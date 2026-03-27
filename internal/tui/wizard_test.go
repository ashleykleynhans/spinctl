package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

func newTestWizard() *WizardPage {
	return NewWizardPage(config.NewDefault())
}

func TestWizardInitialState(t *testing.T) {
	w := newTestWizard()
	if w.step != wizardWelcome {
		t.Errorf("initial step = %v, want wizardWelcome", w.step)
	}
	for _, name := range model.AllServiceNames() {
		if !w.serviceToggles[name] {
			t.Errorf("service %v should be toggled on by default", name)
		}
	}
}

func TestWizardWelcomeEnter(t *testing.T) {
	w := newTestWizard()
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if w.step != wizardVersion {
		t.Errorf("step = %v, want wizardVersion", w.step)
	}
	if !w.editing {
		t.Error("should be in editing mode after welcome enter")
	}
}

func TestWizardVersionEditing(t *testing.T) {
	w := newTestWizard()
	w.Update(tea.KeyMsg{Type: tea.KeyEnter}) // welcome -> version
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
	if w.editBuffer != "1.0" {
		t.Errorf("editBuffer = %q, want 1.0", w.editBuffer)
	}
}

func TestWizardVersionValidation(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.Update(versionValidMsg{version: "1.35.0", err: nil})
	if w.step != wizardServices {
		t.Errorf("step = %v, want wizardServices", w.step)
	}
	if w.cfg.Version != "1.35.0" {
		t.Errorf("cfg.Version = %q, want 1.35.0", w.cfg.Version)
	}
}

func TestWizardVersionValidationError(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.Update(versionValidMsg{version: "bad", err: errors.New("not found")})
	if w.step != wizardVersion {
		t.Errorf("step = %v, want wizardVersion (should stay)", w.step)
	}
	if w.validateErr == "" {
		t.Error("validateErr should be set on error")
	}
	if !w.editing {
		t.Error("should return to editing on validation error")
	}
}

func TestWizardServicesToggle(t *testing.T) {
	w := newTestWizard()
	w.step = wizardServices
	names := model.AllServiceNames()
	// Toggle first service off.
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if w.serviceToggles[names[0]] {
		t.Errorf("service %v should be toggled off", names[0])
	}
	// Toggle it back on.
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !w.serviceToggles[names[0]] {
		t.Errorf("service %v should be toggled on", names[0])
	}
}

func TestWizardServicesNavigation(t *testing.T) {
	w := newTestWizard()
	w.step = wizardServices
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if w.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after j", w.cursor)
	}
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if w.cursor != 0 {
		t.Errorf("cursor = %d, want 0 after k", w.cursor)
	}
}

func TestWizardServicesEnter(t *testing.T) {
	w := newTestWizard()
	w.step = wizardServices
	// Toggle first service off.
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if w.step != wizardProvider {
		t.Errorf("step = %v, want wizardProvider", w.step)
	}
	names := model.AllServiceNames()
	svc := w.cfg.Services[names[0]]
	if svc.Enabled {
		t.Errorf("service %v should be disabled after toggle + enter", names[0])
	}
}

func TestWizardProviderSelect(t *testing.T) {
	w := newTestWizard()
	w.step = wizardProvider
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if w.step != wizardProviderFields {
		t.Errorf("step = %v, want wizardProviderFields", w.step)
	}
	if w.selectedProvider != "kubernetes" {
		t.Errorf("selectedProvider = %q, want kubernetes", w.selectedProvider)
	}
	if w.providerForm == nil {
		t.Error("providerForm should be created")
	}
}

func TestWizardProviderSkip(t *testing.T) {
	w := newTestWizard()
	w.step = wizardProvider
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if w.step != wizardStorage {
		t.Errorf("step = %v, want wizardStorage", w.step)
	}
}

func TestWizardStorageSelect(t *testing.T) {
	w := newTestWizard()
	w.step = wizardStorage
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if w.step != wizardStorageFields {
		t.Errorf("step = %v, want wizardStorageFields", w.step)
	}
	if w.selectedStore != "s3" {
		t.Errorf("selectedStore = %q, want s3", w.selectedStore)
	}
	if w.storageForm == nil {
		t.Error("storageForm should be created")
	}
}

func TestWizardStorageSkip(t *testing.T) {
	w := newTestWizard()
	w.step = wizardStorage
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if w.step != wizardDone {
		t.Errorf("step = %v, want wizardDone", w.step)
	}
}

func TestWizardDone(t *testing.T) {
	w := newTestWizard()
	w.step = wizardDone
	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected wizardDoneMsg command")
	}
	msg := cmd()
	if _, ok := msg.(wizardDoneMsg); !ok {
		t.Fatalf("expected wizardDoneMsg, got %T", msg)
	}
}

func TestWizardViewWelcome(t *testing.T) {
	w := newTestWizard()
	view := w.View()
	if !strings.Contains(view, "Welcome") {
		t.Error("welcome view should contain Welcome")
	}
	if !strings.Contains(view, "enter") {
		t.Error("welcome view should mention enter")
	}
}

func TestWizardViewVersion(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.editing = true
	w.editBuffer = "1.35.0"
	view := w.View()
	if !strings.Contains(strings.ToLower(view), "version") {
		t.Error("version view should mention version")
	}
	if !strings.Contains(view, "1.35.0") {
		t.Error("version view should show the edit buffer")
	}
}

func TestWizardViewVersionValidating(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.validating = true
	w.editBuffer = "1.35.0"
	view := w.View()
	if !strings.Contains(view, "Validating") {
		t.Error("should show Validating when validating=true")
	}
}

func TestWizardViewVersionError(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.editing = true
	w.validateErr = "Version not found"
	view := w.View()
	if !strings.Contains(view, "Version not found") {
		t.Error("should show validation error")
	}
}

func TestWizardViewServices(t *testing.T) {
	w := newTestWizard()
	w.step = wizardServices
	view := w.View()
	if !strings.Contains(strings.ToLower(view), "services") {
		t.Error("services view should mention services")
	}
	if !strings.Contains(view, "ON") {
		t.Error("services view should show ON toggles")
	}
}

func TestWizardViewProvider(t *testing.T) {
	w := newTestWizard()
	w.step = wizardProvider
	view := w.View()
	if !strings.Contains(view, "kubernetes") {
		t.Error("provider view should list kubernetes")
	}
	if !strings.Contains(view, "aws") {
		t.Error("provider view should list aws")
	}
}

func TestWizardViewStorage(t *testing.T) {
	w := newTestWizard()
	w.step = wizardStorage
	view := w.View()
	if !strings.Contains(view, "s3") {
		t.Error("storage view should list s3")
	}
	if !strings.Contains(view, "gcs") {
		t.Error("storage view should list gcs")
	}
}

func TestWizardViewDone(t *testing.T) {
	w := newTestWizard()
	w.step = wizardDone
	w.cfg.Version = "1.35.0"
	view := w.View()
	if !strings.Contains(strings.ToLower(view), "complete") {
		t.Error("done view should mention complete")
	}
	if !strings.Contains(view, "1.35.0") {
		t.Error("done view should show version")
	}
}

func TestWizardProviderFieldsDone(t *testing.T) {
	w := newTestWizard()
	w.step = wizardProviderFields
	w.selectedProvider = "kubernetes"
	w.providerForm = NewFieldForm("k8s", ProviderFields("kubernetes"))

	w.Update(fieldFormDoneMsg{values: map[string]string{
		"name":    "my-k8s",
		"context": "minikube",
	}})

	if w.step != wizardStorage {
		t.Errorf("step = %v, want wizardStorage", w.step)
	}
	prov, ok := w.cfg.Providers["kubernetes"]
	if !ok {
		t.Fatal("kubernetes provider config not created")
	}
	if len(prov.Accounts) != 1 || prov.Accounts[0].Name != "my-k8s" {
		t.Errorf("account = %+v, want name=my-k8s", prov.Accounts)
	}
	if prov.Accounts[0].Context != "minikube" {
		t.Errorf("context = %q, want minikube", prov.Accounts[0].Context)
	}
}

func TestWizardStorageFieldsDone(t *testing.T) {
	w := newTestWizard()
	w.step = wizardStorageFields
	w.selectedStore = "s3"
	w.storageForm = NewFieldForm("s3", StorageFields("s3"))

	w.Update(fieldFormDoneMsg{values: map[string]string{
		"bucket": "my-bucket",
		"region": "us-east-1",
	}})

	if w.step != wizardDone {
		t.Errorf("step = %v, want wizardDone", w.step)
	}
	if w.cfg.PersistentStorage == nil {
		t.Fatal("persistent storage config not set")
	}
	if w.cfg.PersistentStorage["persistentStoreType"] != "s3" {
		t.Errorf("store type = %v, want s3", w.cfg.PersistentStorage["persistentStoreType"])
	}
}

func TestWizardIgnoresKeysDuringValidation(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.validating = true
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !w.validating {
		t.Error("should still be validating")
	}
	if w.step != wizardVersion {
		t.Errorf("step = %v, want wizardVersion", w.step)
	}
}

func TestWizardProviderNavigation(t *testing.T) {
	w := newTestWizard()
	w.step = wizardProvider
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if w.providerCursor != 1 {
		t.Errorf("providerCursor = %d, want 1", w.providerCursor)
	}
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if w.providerCursor != 0 {
		t.Errorf("providerCursor = %d, want 0", w.providerCursor)
	}
}

func TestWizardStorageNavigation(t *testing.T) {
	w := newTestWizard()
	w.step = wizardStorage
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if w.storageCursor != 1 {
		t.Errorf("storageCursor = %d, want 1", w.storageCursor)
	}
}

func TestWizardVersionBackspace(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.editing = true
	w.editBuffer = "1.35"
	w.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if w.editBuffer != "1.3" {
		t.Errorf("editBuffer = %q, want 1.3", w.editBuffer)
	}
}

func TestWizardVersionEscCancelsEditing(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.editing = true
	w.editBuffer = "1.35.0"
	w.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if w.editing {
		t.Error("editing should be false after esc")
	}
}
