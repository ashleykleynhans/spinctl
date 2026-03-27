package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/deploy"
	"github.com/spinnaker/spinctl/internal/halimport"
	"github.com/spinnaker/spinctl/internal/model"
)

// wizardStep tracks which step of the setup wizard we're on.
type wizardStep int

const (
	wizardWelcome wizardStep = iota
	wizardImportPath
	wizardImporting
	wizardVersion
	wizardServices
	wizardProvider
	wizardProviderFields
	wizardStorage
	wizardStorageFields
	wizardDone
)

// versionValidMsg is sent after version validation completes.
type versionValidMsg struct {
	version string
	err     error
}

// WizardPage guides new users through initial Spinnaker configuration.
type WizardPage struct {
	cfg          *config.SpinctlConfig
	halDir       string
	step         wizardStep
	cursor       int
	editing      bool
	editBuffer   string
	validating   bool
	validateErr  string

	// Service selection state.
	serviceToggles map[model.ServiceName]bool

	// Provider selection state.
	providerNames    []string
	providerCursor   int
	selectedProvider string
	providerForm     *FieldForm

	// Storage state.
	storageOptions []string
	storageCursor  int
	selectedStore  string
	storageForm    *FieldForm
}

// NewWizardPage creates the setup wizard.
func NewWizardPage(cfg *config.SpinctlConfig, halDir string) *WizardPage {
	toggles := make(map[model.ServiceName]bool)
	for _, name := range model.AllServiceNames() {
		toggles[name] = true // all enabled by default
	}

	if halDir == "" {
		halDir = halimport.DetectHalDir()
	}

	return &WizardPage{
		cfg:            cfg,
		halDir:         halDir,
		step:           wizardWelcome,
		serviceToggles: toggles,
		providerNames:  []string{"kubernetes", "aws", "gcp", "azure", "cloudfoundry", "oracle"},
		storageOptions: []string{"s3", "gcs", "azs", "oracle", "redis"},
	}
}

func (w *WizardPage) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case versionValidMsg:
		w.validating = false
		if msg.err != nil {
			errStr := msg.err.Error()
			if strings.Contains(errStr, "404") {
				w.validateErr = fmt.Sprintf("Version %q not found. Check the version and try again.", msg.version)
			} else {
				w.validateErr = fmt.Sprintf("Error validating version %q: %s", msg.version, errStr)
			}
			w.editing = true
			return w, nil
		}
		w.cfg.Version = msg.version
		w.validateErr = ""
		w.step = wizardServices
		return w, nil
	case importDoneMsg:
		w.validating = false
		if msg.err != nil {
			w.validateErr = fmt.Sprintf("Import failed: %s", msg.err)
			w.step = wizardImportPath
			w.editing = true
			return w, nil
		}
		if msg.result != nil && msg.result.Config != nil {
			*w.cfg = *msg.result.Config
		}
		w.step = wizardDone
		return w, nil
	case fieldFormDoneMsg:
		return w.handleFormDone(msg)
	case tea.KeyMsg:
		if w.validating {
			return w, nil
		}
		if w.editing {
			return w.updateEditing(msg)
		}
		switch w.step {
		case wizardWelcome:
			return w.updateWelcome(msg)
		case wizardImportPath:
			return w.updateImportPath(msg)
		case wizardImporting:
			return w, nil // ignore keys while importing
		case wizardVersion:
			return w.updateVersion(msg)
		case wizardServices:
			return w.updateServices(msg)
		case wizardProvider:
			return w.updateProvider(msg)
		case wizardProviderFields:
			return w.updateProviderFields(msg)
		case wizardStorage:
			return w.updateStorage(msg)
		case wizardStorageFields:
			return w.updateStorageFields(msg)
		case wizardDone:
			return w.updateDone(msg)
		}
	}
	return w, nil
}

func (w *WizardPage) handleFormDone(msg fieldFormDoneMsg) (page, tea.Cmd) {
	switch w.step {
	case wizardProviderFields:
		// Build provider config from form values.
		account := config.ProviderAccount{
			Name: msg.values["name"],
		}
		extra := make(map[string]any)
		for k, v := range msg.values {
			if k == "name" || v == "" {
				continue
			}
			// Handle comma-separated lists.
			if k == "namespaces" || k == "regions" {
				parts := strings.Split(v, ",")
				trimmed := make([]any, 0, len(parts))
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						trimmed = append(trimmed, p)
					}
				}
				if len(trimmed) > 0 {
					extra[k] = trimmed
				}
				continue
			}
			if k == "context" {
				account.Context = v
				continue
			}
			extra[k] = v
		}
		if len(extra) > 0 {
			account.Extra = extra
		}

		if w.cfg.Providers == nil {
			w.cfg.Providers = make(map[string]config.ProviderConfig)
		}
		w.cfg.Providers[w.selectedProvider] = config.ProviderConfig{
			Enabled:  true,
			Accounts: []config.ProviderAccount{account},
		}
		w.providerForm = nil
		w.step = wizardStorage
		w.cursor = 0

	case wizardStorageFields:
		// Build storage config from form values.
		storeConfig := make(map[string]any)
		for k, v := range msg.values {
			if v != "" {
				storeConfig[k] = v
			}
		}
		w.cfg.PersistentStorage = map[string]any{
			"persistentStoreType": w.selectedStore,
			w.selectedStore:       storeConfig,
		}
		w.storageForm = nil
		w.step = wizardDone
	}
	return w, nil
}

func (w *WizardPage) updateEditing(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if w.editBuffer == "" {
			return w, nil
		}
		w.editing = false
		w.validateErr = ""

		if w.step == wizardImportPath {
			// Run halyard import.
			w.step = wizardImporting
			halPath := w.editBuffer
			outputPath := config.DefaultConfigPath()
			return w, func() tea.Msg {
				result, err := halimport.Import(halimport.ImportOptions{
					HalDir:     halPath,
					OutputPath: outputPath,
				})
				return importDoneMsg{result: result, err: err}
			}
		}

		// Version validation.
		w.validating = true
		ver := w.editBuffer
		return w, func() tea.Msg {
			cacheDir := filepath.Join(config.DefaultConfigDir(), "cache", "bom")
			fetcher := deploy.NewBOMFetcher(deploy.DefaultBOMURLPattern, cacheDir)
			_, err := fetcher.Fetch(ver)
			return versionValidMsg{version: ver, err: err}
		}
	case tea.KeyEscape:
		if w.step == wizardImportPath {
			w.step = wizardWelcome
			w.cursor = 0
		}
		w.editing = false
	case tea.KeyBackspace:
		if len(w.editBuffer) > 0 {
			w.editBuffer = w.editBuffer[:len(w.editBuffer)-1]
		}
	case tea.KeyRunes:
		w.editBuffer += string(msg.Runes)
	}
	return w, nil
}

func (w *WizardPage) updateImportPath(msg tea.KeyMsg) (page, tea.Cmd) {
	// Handled by editing mode.
	return w, nil
}

func (w *WizardPage) updateWelcome(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		w.cursor = 0
	case "down", "j":
		w.cursor = 1
	case "enter":
		if w.cursor == 0 {
			// Import from Halyard.
			w.step = wizardImportPath
			w.editing = true
			w.editBuffer = w.halDir
		} else {
			// Fresh install.
			w.step = wizardVersion
			w.editing = true
			w.editBuffer = w.cfg.Version
		}
	}
	return w, nil
}

func (w *WizardPage) updateVersion(msg tea.KeyMsg) (page, tea.Cmd) {
	return w, nil
}

func (w *WizardPage) updateServices(msg tea.KeyMsg) (page, tea.Cmd) {
	names := model.AllServiceNames()
	switch msg.String() {
	case "up", "k":
		w.cursor--
		if w.cursor < 0 {
			w.cursor = len(names) - 1
		}
	case "down", "j":
		w.cursor++
		if w.cursor >= len(names) {
			w.cursor = 0
		}
	case " ":
		if w.cursor >= 0 && w.cursor < len(names) {
			name := names[w.cursor]
			w.serviceToggles[name] = !w.serviceToggles[name]
		}
	case "enter":
		for _, name := range names {
			svc := w.cfg.Services[name]
			svc.Enabled = w.serviceToggles[name]
			w.cfg.Services[name] = svc
		}
		w.cursor = 0
		w.step = wizardProvider
	}
	return w, nil
}

func (w *WizardPage) updateProvider(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		w.providerCursor--
		if w.providerCursor < 0 {
			w.providerCursor = len(w.providerNames) - 1
		}
	case "down", "j":
		w.providerCursor++
		if w.providerCursor >= len(w.providerNames) {
			w.providerCursor = 0
		}
	case "enter":
		w.selectedProvider = w.providerNames[w.providerCursor]
		fields := ProviderFields(w.selectedProvider)
		w.providerForm = NewFieldForm(
			fmt.Sprintf("Configure %s", w.selectedProvider),
			fields,
		)
		w.step = wizardProviderFields
	case "s":
		w.step = wizardStorage
	}
	return w, nil
}

func (w *WizardPage) updateProviderFields(msg tea.KeyMsg) (page, tea.Cmd) {
	if w.providerForm != nil {
		_, cmd := w.providerForm.Update(msg)
		return w, cmd
	}
	return w, nil
}

func (w *WizardPage) updateStorage(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		w.storageCursor--
		if w.storageCursor < 0 {
			w.storageCursor = len(w.storageOptions) - 1
		}
	case "down", "j":
		w.storageCursor++
		if w.storageCursor >= len(w.storageOptions) {
			w.storageCursor = 0
		}
	case "enter":
		w.selectedStore = w.storageOptions[w.storageCursor]
		fields := StorageFields(w.selectedStore)
		w.storageForm = NewFieldForm(
			fmt.Sprintf("Configure %s storage", w.selectedStore),
			fields,
		)
		w.step = wizardStorageFields
	case "s":
		w.step = wizardDone
	}
	return w, nil
}

func (w *WizardPage) updateStorageFields(msg tea.KeyMsg) (page, tea.Cmd) {
	if w.storageForm != nil {
		_, cmd := w.storageForm.Update(msg)
		return w, cmd
	}
	return w, nil
}

func (w *WizardPage) updateDone(msg tea.KeyMsg) (page, tea.Cmd) {
	if msg.String() == "enter" {
		return w, func() tea.Msg { return wizardDoneMsg{} }
	}
	return w, nil
}

// wizardDoneMsg signals the wizard is complete.
type wizardDoneMsg struct{}

func (w *WizardPage) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headingStyle.Render("Spinnaker Setup"))
	b.WriteString("\n\n")

	switch w.step {
	case wizardWelcome:
		b.WriteString("  " + keyStyle.Render("Welcome to spinctl!") + "\n\n")
		b.WriteString("  " + valueStyle.Render("No existing configuration found.") + "\n")
		b.WriteString("  " + valueStyle.Render("How would you like to get started?") + "\n\n")

		options := []string{"Import from Halyard", "Fresh install"}
		for i, opt := range options {
			selected := i == w.cursor
			cursor := "  "
			if selected {
				cursor = menuCursorStyle.Render("▸ ")
			}
			label := keyStyle.Render(opt)
			if selected {
				label = keySelectedStyle.Render(opt)
			}
			b.WriteString(cursor + label + "\n")
		}
		b.WriteString("\n  " + menuDescStyle.Render("enter: select") + "\n")

	case wizardImportPath:
		b.WriteString("  " + stepIndicator(1, 1) + "\n\n")
		if w.validateErr != "" {
			b.WriteString("  " + errorStyle.Render(w.validateErr) + "\n\n")
		}
		b.WriteString("  " + keyStyle.Render("Halyard config path: ") + editCursorStyle.Render(w.editBuffer+"█") + "\n\n")
		if w.halDir != "" {
			b.WriteString("  " + menuDescStyle.Render("Detected: "+w.halDir) + "\n")
		}
		b.WriteString("  " + menuDescStyle.Render("enter: import  esc: back") + "\n")

	case wizardImporting:
		b.WriteString("  " + keyStyle.Render("Importing from "+w.editBuffer+"...") + "\n")

	case wizardVersion:
		b.WriteString("  " + stepIndicator(1, 5) + "\n\n")
		if w.validating {
			b.WriteString("  " + keyStyle.Render("Spinnaker version: ") + valueStyle.Render(w.editBuffer) + "\n\n")
			b.WriteString("  " + keyStyle.Render("Validating version...") + "\n")
		} else {
			if w.validateErr != "" {
				b.WriteString("  " + errorStyle.Render(w.validateErr) + "\n\n")
			}
			b.WriteString("  " + keyStyle.Render("Spinnaker version: ") + editCursorStyle.Render(w.editBuffer+"█") + "\n\n")
			b.WriteString("  " + menuDescStyle.Render("Example: 2025.3.2") + "\n")
			b.WriteString("  " + menuDescStyle.Render("enter: validate and continue") + "\n")
		}

	case wizardServices:
		b.WriteString("  " + stepIndicator(2, 5) + "\n\n")
		b.WriteString("  " + keyStyle.Render("Which services do you want to enable?") + "\n\n")
		names := model.AllServiceNames()
		for i, name := range names {
			selected := i == w.cursor
			cursor := "  "
			if selected {
				cursor = menuCursorStyle.Render("▸ ")
			}
			status := offStyle.Render("[OFF]")
			if w.serviceToggles[name] {
				status = onStyle.Render("[ ON]")
			}
			label := keyStyle.Render(fmt.Sprintf("%-15s", name))
			if selected {
				label = keySelectedStyle.Render(fmt.Sprintf("%-15s", name))
			}
			port := valueStyle.Render(fmt.Sprintf(":%d", config.DefaultPort(name)))
			b.WriteString(cursor + status + " " + label + " " + port + "\n")
		}
		b.WriteString("\n  " + menuDescStyle.Render("space: toggle  enter: continue") + "\n")

	case wizardProvider:
		b.WriteString("  " + stepIndicator(3, 5) + "\n\n")
		b.WriteString("  " + keyStyle.Render("Select a cloud provider:") + "\n\n")
		for i, name := range w.providerNames {
			selected := i == w.providerCursor
			cursor := "  "
			if selected {
				cursor = menuCursorStyle.Render("▸ ")
			}
			label := keyStyle.Render(name)
			if selected {
				label = keySelectedStyle.Render(name)
			}
			b.WriteString(cursor + label + "\n")
		}
		b.WriteString("\n  " + menuDescStyle.Render("enter: select  s: skip") + "\n")

	case wizardProviderFields:
		if w.providerForm != nil {
			return w.providerForm.View()
		}

	case wizardStorage:
		b.WriteString("  " + stepIndicator(4, 5) + "\n\n")
		b.WriteString("  " + keyStyle.Render("Select persistent storage backend:") + "\n\n")
		for i, name := range w.storageOptions {
			selected := i == w.storageCursor
			cursor := "  "
			if selected {
				cursor = menuCursorStyle.Render("▸ ")
			}
			label := keyStyle.Render(name)
			if selected {
				label = keySelectedStyle.Render(name)
			}
			b.WriteString(cursor + label + "\n")
		}
		b.WriteString("\n  " + menuDescStyle.Render("enter: select  s: skip") + "\n")

	case wizardStorageFields:
		if w.storageForm != nil {
			return w.storageForm.View()
		}

	case wizardDone:
		b.WriteString("  " + successStyle.Render("Setup complete!") + "\n\n")
		b.WriteString("  " + keyStyle.Render("Version:  ") + onStyle.Render(w.cfg.Version) + "\n")

		enabledCount := 0
		for _, svc := range w.cfg.Services {
			if svc.Enabled {
				enabledCount++
			}
		}
		b.WriteString("  " + keyStyle.Render("Services: ") + valueStyle.Render(fmt.Sprintf("%d enabled", enabledCount)) + "\n")

		if len(w.cfg.Providers) > 0 {
			for name, prov := range w.cfg.Providers {
				b.WriteString("  " + keyStyle.Render("Provider: ") + onStyle.Render(name) +
					valueStyle.Render(fmt.Sprintf(" (%d account(s))", len(prov.Accounts))) + "\n")
			}
		} else {
			b.WriteString("  " + keyStyle.Render("Provider: ") + valueStyle.Render("none (skipped)") + "\n")
		}

		if w.cfg.PersistentStorage != nil {
			if st, ok := w.cfg.PersistentStorage["persistentStoreType"].(string); ok {
				b.WriteString("  " + keyStyle.Render("Storage:  ") + onStyle.Render(st) + "\n")
			}
		} else {
			b.WriteString("  " + keyStyle.Render("Storage:  ") + valueStyle.Render("none (skipped)") + "\n")
		}

		b.WriteString("\n  " + menuDescStyle.Render("Press enter to save and continue to the main screen.") + "\n")
	}

	return b.String()
}

func stepIndicator(current, total int) string {
	return valueStyle.Render(fmt.Sprintf("Step %d of %d", current, total))
}
