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
	return NewWizardPage(config.NewDefault(), "")
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

func TestWizardWelcomeImport(t *testing.T) {
	w := newTestWizard()
	// Default cursor=0 is "Import from Halyard".
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if w.step != wizardImportPath {
		t.Errorf("step = %v, want wizardImportPath", w.step)
	}
	if !w.editing {
		t.Error("should be in editing mode for import path")
	}
}

func TestWizardWelcomeFreshInstall(t *testing.T) {
	w := newTestWizard()
	// Move to "Fresh install" (cursor=1).
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if w.step != wizardVersion {
		t.Errorf("step = %v, want wizardVersion", w.step)
	}
	if !w.editing {
		t.Error("should be in editing mode after selecting fresh install")
	}
}

func TestWizardVersionEditing(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.editing = true
	w.editBuffer = ""
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

func TestWizardUpdateProviderFields(t *testing.T) {
	w := newTestWizard()
	w.step = wizardProviderFields
	w.selectedProvider = "kubernetes"
	w.providerForm = NewFieldForm("k8s", ProviderFields("kubernetes"))

	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	// The form should have received the key (buffer updated).
	if w.providerForm.buffer != w.providerForm.fields[0].Default+"a" {
		t.Errorf("providerForm buffer = %q, expected delegation to form", w.providerForm.buffer)
	}
	_ = cmd
}

func TestWizardUpdateStorageFields(t *testing.T) {
	w := newTestWizard()
	w.step = wizardStorageFields
	w.selectedStore = "s3"
	w.storageForm = NewFieldForm("s3", StorageFields("s3"))

	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	// The form should have received the key.
	if !strings.Contains(w.storageForm.buffer, "x") {
		t.Errorf("storageForm buffer = %q, expected delegation to form", w.storageForm.buffer)
	}
	_ = cmd
}

func TestWizardStorageNavigationKUp(t *testing.T) {
	w := newTestWizard()
	w.step = wizardStorage
	// Move down then up.
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if w.storageCursor != 1 {
		t.Errorf("storageCursor = %d, want 1", w.storageCursor)
	}
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if w.storageCursor != 0 {
		t.Errorf("storageCursor = %d, want 0 after k", w.storageCursor)
	}
}

func TestWizardHandleFormDoneProviderWithExtras(t *testing.T) {
	w := newTestWizard()
	w.step = wizardProviderFields
	w.selectedProvider = "aws"
	w.providerForm = NewFieldForm("aws", ProviderFields("aws"))

	w.Update(fieldFormDoneMsg{values: map[string]string{
		"name":           "my-aws",
		"accountId":      "123456789",
		"regions":        "us-west-2, us-east-1",
		"context":        "my-context",
		"defaultKeyPair": "my-key",
	}})

	if w.step != wizardStorage {
		t.Errorf("step = %v, want wizardStorage", w.step)
	}
	prov, ok := w.cfg.Providers["aws"]
	if !ok {
		t.Fatal("aws provider config not created")
	}
	acct := prov.Accounts[0]
	if acct.Name != "my-aws" {
		t.Errorf("account name = %q, want my-aws", acct.Name)
	}
	if acct.Context != "my-context" {
		t.Errorf("account context = %q, want my-context", acct.Context)
	}
	// Regions should be stored as a list in Extra.
	regions, ok := acct.Extra["regions"]
	if !ok {
		t.Fatal("regions not in Extra")
	}
	regionList, ok := regions.([]any)
	if !ok {
		t.Fatalf("regions type = %T, want []any", regions)
	}
	if len(regionList) != 2 {
		t.Errorf("regions count = %d, want 2", len(regionList))
	}
	// accountId and defaultKeyPair should be in Extra.
	if acct.Extra["accountId"] != "123456789" {
		t.Errorf("accountId = %v, want 123456789", acct.Extra["accountId"])
	}
	if acct.Extra["defaultKeyPair"] != "my-key" {
		t.Errorf("defaultKeyPair = %v, want my-key", acct.Extra["defaultKeyPair"])
	}
}

func TestWizardHandleFormDoneStorage(t *testing.T) {
	w := newTestWizard()
	w.step = wizardStorageFields
	w.selectedStore = "gcs"
	w.storageForm = NewFieldForm("gcs", StorageFields("gcs"))

	w.Update(fieldFormDoneMsg{values: map[string]string{
		"bucket":  "my-gcs-bucket",
		"project": "my-project",
		"jsonPath": "/path/to/key.json",
	}})

	if w.step != wizardDone {
		t.Errorf("step = %v, want wizardDone", w.step)
	}
	if w.cfg.PersistentStorage == nil {
		t.Fatal("persistent storage not set")
	}
	if w.cfg.PersistentStorage["persistentStoreType"] != "gcs" {
		t.Errorf("store type = %v, want gcs", w.cfg.PersistentStorage["persistentStoreType"])
	}
	storeConfig, ok := w.cfg.PersistentStorage["gcs"].(map[string]any)
	if !ok {
		t.Fatalf("gcs config type = %T, want map[string]any", w.cfg.PersistentStorage["gcs"])
	}
	if storeConfig["bucket"] != "my-gcs-bucket" {
		t.Errorf("bucket = %v, want my-gcs-bucket", storeConfig["bucket"])
	}
}

func TestWizardVersionEditingBackspace(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.editing = true
	w.editBuffer = ""
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if w.editBuffer != "abc" {
		t.Errorf("editBuffer = %q, want abc", w.editBuffer)
	}
	w.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if w.editBuffer != "ab" {
		t.Errorf("editBuffer after backspace = %q, want ab", w.editBuffer)
	}
}

func TestWizardVersionEditingEscape(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.editing = true
	w.editBuffer = "1.35.0"
	w.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if w.editing {
		t.Error("editing should be false after escape")
	}
}

func TestWizardVersionEditingRunes(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.editing = true
	w.editBuffer = ""
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
	if w.editBuffer != "2.0" {
		t.Errorf("editBuffer = %q, want 2.0", w.editBuffer)
	}
}

func TestWizardViewProviderFields(t *testing.T) {
	w := newTestWizard()
	w.step = wizardProviderFields
	w.providerForm = NewFieldForm("Configure kubernetes", ProviderFields("kubernetes"))
	view := w.View()
	if !strings.Contains(view, "Configure kubernetes") {
		t.Error("providerFields view should render the form title")
	}
}

func TestWizardViewStorageFields(t *testing.T) {
	w := newTestWizard()
	w.step = wizardStorageFields
	w.storageForm = NewFieldForm("Configure s3 storage", StorageFields("s3"))
	view := w.View()
	if !strings.Contains(view, "Configure s3 storage") {
		t.Error("storageFields view should render the form title")
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

func TestWizardEditingEmptyEnterDoesNothing(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.editing = true
	w.editBuffer = ""
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// With empty buffer, should stay editing and not start validation.
	if w.validating {
		t.Error("should not start validation with empty buffer")
	}
}

func TestWizardEditingBackspaceOnEmpty(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	w.editing = true
	w.editBuffer = ""
	w.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if w.editBuffer != "" {
		t.Errorf("editBuffer = %q, want empty", w.editBuffer)
	}
}

func TestWizardUpdateVersionReturnsNoCmd(t *testing.T) {
	w := newTestWizard()
	w.step = wizardVersion
	// When not editing, version step returns w, nil.
	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd != nil {
		t.Error("updateVersion should return nil cmd")
	}
}

func TestWizardUpdateProviderFieldsNilForm(t *testing.T) {
	w := newTestWizard()
	w.step = wizardProviderFields
	w.providerForm = nil
	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd != nil {
		t.Error("nil providerForm should return nil cmd")
	}
}

func TestWizardUpdateStorageFieldsNilForm(t *testing.T) {
	w := newTestWizard()
	w.step = wizardStorageFields
	w.storageForm = nil
	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd != nil {
		t.Error("nil storageForm should return nil cmd")
	}
}

func TestWizardUpdateDoneNonEnterKey(t *testing.T) {
	w := newTestWizard()
	w.step = wizardDone
	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd != nil {
		t.Error("non-enter key on done step should return nil cmd")
	}
}

func TestWizardProviderNavigationWrapping(t *testing.T) {
	w := newTestWizard()
	w.step = wizardProvider
	// Wrap up from 0.
	w.providerCursor = 0
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if w.providerCursor != len(w.providerNames)-1 {
		t.Errorf("providerCursor = %d, want %d", w.providerCursor, len(w.providerNames)-1)
	}
	// Wrap down from last.
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if w.providerCursor != 0 {
		t.Errorf("providerCursor = %d, want 0", w.providerCursor)
	}
}

func TestWizardStorageNavigationWrapping(t *testing.T) {
	w := newTestWizard()
	w.step = wizardStorage
	// Wrap up from 0.
	w.storageCursor = 0
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if w.storageCursor != len(w.storageOptions)-1 {
		t.Errorf("storageCursor = %d, want %d", w.storageCursor, len(w.storageOptions)-1)
	}
	// Wrap down from last.
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if w.storageCursor != 0 {
		t.Errorf("storageCursor = %d, want 0", w.storageCursor)
	}
}

func TestWizardServicesCursorWrapping(t *testing.T) {
	w := newTestWizard()
	w.step = wizardServices
	names := model.AllServiceNames()
	// Wrap up from 0.
	w.cursor = 0
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if w.cursor != len(names)-1 {
		t.Errorf("cursor = %d, want %d", w.cursor, len(names)-1)
	}
	// Wrap down from last.
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if w.cursor != 0 {
		t.Errorf("cursor = %d, want 0", w.cursor)
	}
}

func TestWizardViewDoneWithProviders(t *testing.T) {
	w := newTestWizard()
	w.step = wizardDone
	w.cfg.Version = "1.35.0"
	w.cfg.Providers = map[string]config.ProviderConfig{
		"kubernetes": {Enabled: true, Accounts: []config.ProviderAccount{{Name: "prod"}}},
	}
	view := w.View()
	if !strings.Contains(view, "kubernetes") {
		t.Error("done view should show provider name")
	}
	if !strings.Contains(view, "1 account") {
		t.Error("done view should show account count")
	}
}

func TestWizardViewDoneWithStorage(t *testing.T) {
	w := newTestWizard()
	w.step = wizardDone
	w.cfg.Version = "1.35.0"
	w.cfg.PersistentStorage = map[string]any{
		"persistentStoreType": "s3",
	}
	view := w.View()
	if !strings.Contains(view, "s3") {
		t.Error("done view should show storage type")
	}
}

func TestWizardViewDoneNoStorageNoProviders(t *testing.T) {
	w := newTestWizard()
	w.step = wizardDone
	w.cfg.Version = "1.35.0"
	view := w.View()
	if !strings.Contains(view, "none (skipped)") {
		t.Error("done view should show 'none (skipped)' when no providers/storage")
	}
}
