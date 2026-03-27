package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

// wizardStep tracks which step of the setup wizard we're on.
type wizardStep int

const (
	wizardWelcome wizardStep = iota
	wizardVersion
	wizardServices
	wizardProvider
	wizardProviderAccount
	wizardStorage
	wizardDone
)

// WizardPage guides new users through initial Spinnaker configuration.
type WizardPage struct {
	cfg        *config.SpinctlConfig
	step       wizardStep
	cursor     int
	editing    bool
	editBuffer string

	// Service selection state.
	serviceToggles map[model.ServiceName]bool

	// Provider selection state.
	providerNames  []string
	providerCursor int

	// Provider account state.
	selectedProvider string
	accountName      string
	accountStep      int // 0=name

	// Storage state.
	storageOptions []string
	storageCursor  int
}

// NewWizardPage creates the setup wizard.
func NewWizardPage(cfg *config.SpinctlConfig) *WizardPage {
	toggles := make(map[model.ServiceName]bool)
	for _, name := range model.AllServiceNames() {
		toggles[name] = true // all enabled by default
	}

	return &WizardPage{
		cfg:            cfg,
		step:           wizardWelcome,
		serviceToggles: toggles,
		providerNames:  []string{"kubernetes", "aws", "gcp", "azure", "cloudfoundry", "oracle"},
		storageOptions: []string{"s3", "gcs", "azs", "oracle", "redis"},
	}
}

func (w *WizardPage) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if w.editing {
			return w.updateEditing(msg)
		}
		switch w.step {
		case wizardWelcome:
			return w.updateWelcome(msg)
		case wizardVersion:
			return w.updateVersion(msg)
		case wizardServices:
			return w.updateServices(msg)
		case wizardProvider:
			return w.updateProvider(msg)
		case wizardProviderAccount:
			return w.updateProviderAccount(msg)
		case wizardStorage:
			return w.updateStorage(msg)
		case wizardDone:
			return w.updateDone(msg)
		}
	}
	return w, nil
}

func (w *WizardPage) updateEditing(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		w.editing = false
		w.applyEdit()
	case tea.KeyEscape:
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

func (w *WizardPage) applyEdit() {
	switch w.step {
	case wizardVersion:
		w.cfg.Version = w.editBuffer
		w.step = wizardServices
	case wizardProviderAccount:
		w.accountName = w.editBuffer
		// Create the provider and account.
		w.cfg.Providers[w.selectedProvider] = config.ProviderConfig{
			Enabled:  true,
			Accounts: []config.ProviderAccount{{Name: w.accountName}},
		}
		w.step = wizardStorage
	}
}

func (w *WizardPage) updateWelcome(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.String() {
	case "enter":
		w.step = wizardVersion
		w.editing = true
		w.editBuffer = w.cfg.Version
	}
	return w, nil
}

func (w *WizardPage) updateVersion(msg tea.KeyMsg) (page, tea.Cmd) {
	// Handled by editing mode.
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
		// Apply selections and move on.
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
		if w.cfg.Providers == nil {
			w.cfg.Providers = make(map[string]config.ProviderConfig)
		}
		w.step = wizardProviderAccount
		w.editing = true
		w.editBuffer = ""
	case "s":
		// Skip provider setup.
		w.step = wizardStorage
	}
	return w, nil
}

func (w *WizardPage) updateProviderAccount(msg tea.KeyMsg) (page, tea.Cmd) {
	// Handled by editing mode.
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
		selected := w.storageOptions[w.storageCursor]
		w.cfg.PersistentStorage = map[string]any{
			"persistentStoreType": selected,
			selected:              map[string]any{},
		}
		w.step = wizardDone
	case "s":
		// Skip storage setup.
		w.step = wizardDone
	}
	return w, nil
}

func (w *WizardPage) updateDone(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.String() {
	case "enter":
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
		b.WriteString("  " + valueStyle.Render("This wizard will help you set up Spinnaker.") + "\n\n")
		b.WriteString("  " + menuDescStyle.Render("Press enter to begin.") + "\n")

	case wizardVersion:
		b.WriteString("  " + stepIndicator(1, 5) + "\n\n")
		b.WriteString("  " + keyStyle.Render("Spinnaker version: ") + editCursorStyle.Render(w.editBuffer+"█") + "\n\n")
		b.WriteString("  " + menuDescStyle.Render("Example: 2025.3.2") + "\n")
		b.WriteString("  " + menuDescStyle.Render("enter: confirm") + "\n")

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

	case wizardProviderAccount:
		b.WriteString("  " + stepIndicator(4, 5) + "\n\n")
		b.WriteString("  " + keyStyle.Render("Provider: ") + onStyle.Render(w.selectedProvider) + "\n\n")
		b.WriteString("  " + keyStyle.Render("Account name: ") + editCursorStyle.Render(w.editBuffer+"█") + "\n\n")
		b.WriteString("  " + menuDescStyle.Render("enter: confirm") + "\n")

	case wizardStorage:
		b.WriteString("  " + stepIndicator(5, 5) + "\n\n")
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

		providerCount := len(w.cfg.Providers)
		b.WriteString("  " + keyStyle.Render("Providers:") + valueStyle.Render(fmt.Sprintf(" %d configured", providerCount)) + "\n")

		if w.cfg.PersistentStorage != nil {
			if st, ok := w.cfg.PersistentStorage["persistentStoreType"].(string); ok {
				b.WriteString("  " + keyStyle.Render("Storage:  ") + valueStyle.Render(st) + "\n")
			}
		}

		b.WriteString("\n  " + menuDescStyle.Render("Press enter to save and continue to the main screen.") + "\n")
	}

	return b.String()
}

func stepIndicator(current, total int) string {
	return valueStyle.Render(fmt.Sprintf("Step %d of %d", current, total))
}
