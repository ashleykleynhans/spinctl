package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spinnaker/spinctl/internal/config"
)

// findMenuCursor returns the cursor index for a given page action in the home page.
func findMenuCursor(hp *HomePage, target PageID) int {
	for i, item := range hp.items {
		if item.action == target {
			return i
		}
	}
	return -1
}

func TestAppInit(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	app := NewApp(cfg, "", "test")
	if app.currentPage != PageHome {
		t.Errorf("initial page = %v, want PageHome", app.currentPage)
	}
}

func TestAppQuitOnQ(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestAppViewContainsTitle(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	app := NewApp(cfg, "", "test")
	view := app.View()
	if !strings.Contains(view, "spinctl") {
		t.Error("view should contain 'spinctl'")
	}
}

func TestAppWindowResize(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if app.width != 120 || app.height != 40 {
		t.Errorf("size = %dx%d, want 120x40", app.width, app.height)
	}
}

func TestAppQuitOnCtrlC(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("expected quit command for ctrl+c")
	}
}

func TestAppViewDevVersion(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "0.0.3")
	view := app.View()
	if !strings.Contains(view, "0.0.3") {
		t.Error("view should contain spinctl version")
	}
}

func TestAppViewWithVersion(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	app := NewApp(cfg, "", "0.0.3")
	view := app.View()
	if !strings.Contains(view, "0.0.3") {
		t.Error("view should contain the version")
	}
}

func TestAppWindowResizeUpdatesStatusBar(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")
	app.dirty = true
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	view := app.View()
	if !strings.Contains(view, "modified") {
		t.Error("status bar should show 'modified' when dirty")
	}
}

func TestAppViewShowsStatusBar(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")
	view := app.View()
	if !strings.Contains(view, "quit") {
		t.Error("view should contain status bar with quit hint")
	}
}

func TestAppNavigateToServices(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	app := NewApp(cfg, "", "test")

	// Select first item (Services) on home page by pressing enter.
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageServices {
		t.Errorf("currentPage = %v, want PageServices", app.currentPage)
	}
	if app.servicesPage == nil {
		t.Error("services page should be created after navigation")
	}
	view := app.View()
	if !strings.Contains(view, "Services") {
		t.Error("view should show services page")
	}
}

func TestAppNavigateToImport(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")

	hp := app.homePage.(*HomePage)
	hp.cursor = findMenuCursor(hp, PageImport)
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageImport {
		t.Errorf("currentPage = %v, want PageImport", app.currentPage)
	}
	view := app.View()
	if !strings.Contains(view, "Import") {
		t.Error("view should show import page")
	}
}

func TestAppNavigateToDeploy(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")

	hp := app.homePage.(*HomePage)
	hp.cursor = findMenuCursor(hp, PageDeploy)
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageDeploy {
		t.Errorf("currentPage = %v, want PageDeploy", app.currentPage)
	}
	view := app.View()
	if !strings.Contains(view, "Deploy") {
		t.Error("view should show deploy page")
	}
}

func TestAppPageStackTracksHistory(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")

	if len(app.pageStack) != 0 {
		t.Errorf("initial pageStack len = %d, want 0", len(app.pageStack))
	}

	// Navigate to Services.
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if len(app.pageStack) != 1 {
		t.Errorf("pageStack len after navigate = %d, want 1", len(app.pageStack))
	}
	if app.pageStack[0] != PageHome {
		t.Errorf("pageStack[0] = %v, want PageHome", app.pageStack[0])
	}
}

func TestAppInitReturnsNil(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")
	cmd := app.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestAppDelegatesNonKeyMessages(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")
	// Send a non-key, non-window message; should not panic.
	type customMsg struct{}
	_, cmd := app.Update(customMsg{})
	if cmd != nil {
		t.Error("non-key message should not produce a command")
	}
}

func TestAppDelegatesToServicesPage(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	app := NewApp(cfg, "", "test")

	// Navigate to services page.
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageServices {
		t.Fatalf("expected PageServices, got %v", app.currentPage)
	}

	// Send a key to the services page (down arrow).
	app.Update(tea.KeyMsg{Type: tea.KeyDown})
	sp := app.servicesPage.(*ServicesPage)
	if sp.cursor != 1 {
		t.Errorf("services cursor = %d, want 1", sp.cursor)
	}
}

func TestAppDelegatesToImportPage(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")

	// Navigate to import page.
	hp := app.homePage.(*HomePage)
	hp.cursor = findMenuCursor(hp, PageImport)
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageImport {
		t.Fatalf("expected PageImport, got %v", app.currentPage)
	}

	// Send a key to the import page (y to confirm).
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	ip := app.importPage.(*ImportPage)
	if !ip.confirmed {
		t.Error("import page should be confirmed after 'y'")
	}
}

func TestAppDelegatesToDeployPage(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")

	// Navigate to deploy page.
	hp := app.homePage.(*HomePage)
	hp.cursor = findMenuCursor(hp, PageDeploy)
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageDeploy {
		t.Fatalf("expected PageDeploy, got %v", app.currentPage)
	}

	// Send 'y' to confirm deploy.
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	dp := app.deployPage.(*DeployPage)
	if !dp.confirmed {
		t.Error("deploy page should be confirmed after 'y'")
	}
}

func TestAppNavigateToEditor(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")

	// Navigate to Providers (index 1).
	hp := app.homePage.(*HomePage)
	hp.cursor = 1
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageProviders {
		t.Errorf("expected PageProviders, got %v", app.currentPage)
	}
}

func TestAppViewEditorPage(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")

	// Navigate to editor page (Providers).
	hp := app.homePage.(*HomePage)
	hp.cursor = 1
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// View should not panic even if editorPage is nil (navigateTo doesn't create it for PageEditor).
	view := app.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestAppSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	app := NewApp(cfg, configPath, "test")
	app.dirty = true

	// Press 's' on home page to trigger save.
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if cmd == nil {
		t.Fatal("expected save command")
	}

	// Execute the cmd to get the saveResultMsg.
	msg := cmd()
	result, ok := msg.(saveResultMsg)
	if !ok {
		t.Fatalf("expected saveResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Errorf("save error = %v", result.err)
	}
}

func TestAppGoBack(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	app := NewApp(cfg, "", "test")

	// Navigate to services.
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageServices {
		t.Fatalf("expected PageServices, got %v", app.currentPage)
	}

	// Send goBackMsg.
	app.Update(goBackMsg{})
	if app.currentPage != PageHome {
		t.Errorf("after goBack, currentPage = %v, want PageHome", app.currentPage)
	}
}

func TestAppConfigChangedMsg(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")

	if app.dirty {
		t.Error("should not be dirty initially")
	}

	// Actually change the config so checkDirty detects a difference.
	cfg.Version = "changed"
	app.Update(configChangedMsg{})
	if !app.dirty {
		t.Error("should be dirty after config was modified")
	}
}

func TestAppConfigChangedMsgRevert(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")

	// Change config.
	cfg.Version = "changed"
	app.Update(configChangedMsg{})
	if !app.dirty {
		t.Error("should be dirty after change")
	}

	// Revert config to original state.
	cfg.Version = ""
	app.Update(configChangedMsg{})
	if app.dirty {
		t.Error("should not be dirty after reverting to original state")
	}
}

func TestAppQuitConfirmDirty(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")
	app.dirty = true

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if !app.confirmQuit {
		t.Error("should show confirmQuit when dirty")
	}
}

func TestAppQuitConfirmYes(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")
	app.dirty = true
	app.confirmQuit = true

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestAppQuitConfirmNo(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")
	app.confirmQuit = true

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if app.confirmQuit {
		t.Error("confirmQuit should be false after 'n'")
	}
}

func TestAppSaveResultSuccess(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")
	app.dirty = true

	app.Update(saveResultMsg{err: nil})
	if app.dirty {
		t.Error("dirty should be false after successful save")
	}
}

func TestAppSaveResultError(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")
	app.dirty = true

	app.Update(saveResultMsg{err: fmt.Errorf("write failed")})
	if !app.dirty {
		t.Error("dirty should still be true after failed save")
	}
	if !strings.Contains(app.saveMessage, "Save failed") {
		t.Error("saveMessage should contain error info")
	}
}

func TestAppNavigateToVersion(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	app := NewApp(cfg, "", "test")

	hp := app.homePage.(*HomePage)
	hp.cursor = findMenuCursor(hp, PageVersion)
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageVersion {
		t.Errorf("currentPage = %v, want PageVersion", app.currentPage)
	}
}

func TestAppNavigateToSecurity(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")

	hp := app.homePage.(*HomePage)
	hp.cursor = findMenuCursor(hp, PageSecurity)
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageSecurity {
		t.Errorf("currentPage = %v, want PageSecurity", app.currentPage)
	}
}

func TestAppNavigateToFeatures(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")

	hp := app.homePage.(*HomePage)
	hp.cursor = findMenuCursor(hp, PageFeatures)
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageFeatures {
		t.Errorf("currentPage = %v, want PageFeatures", app.currentPage)
	}
}

func TestAppConfirmQuitView(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "", "test")
	app.confirmQuit = true

	view := app.View()
	if !strings.Contains(view, "unsaved changes") {
		t.Error("confirm quit view should mention unsaved changes")
	}
}

func TestAppViewAllPages(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"

	// Test each page renders without panic.
	pages := []struct {
		name   string
		pageID PageID
	}{
		{"Services", PageServices},
		{"Import", PageImport},
		{"Deploy", PageDeploy},
	}
	for _, tc := range pages {
		t.Run(tc.name, func(t *testing.T) {
			app := NewApp(cfg, "", "test")
			hp := app.homePage.(*HomePage)
			hp.cursor = findMenuCursor(hp, tc.pageID)
			app.Update(tea.KeyMsg{Type: tea.KeyEnter})
			view := app.View()
			if view == "" {
				t.Errorf("view for %s should not be empty", tc.name)
			}
		})
	}
}

func TestAppNavigateToArtifacts(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Artifacts = map[string]any{"github": map[string]any{"enabled": true}}
	app := NewApp(cfg, "", "test")

	hp := app.homePage.(*HomePage)
	hp.cursor = findMenuCursor(hp, PageArtifacts)
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageArtifacts {
		t.Errorf("currentPage = %v, want PageArtifacts", app.currentPage)
	}
	if app.editorPage == nil {
		t.Error("editor page should be created for Artifacts")
	}
}

func TestAppNavigateToNotifications(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Notifications = map[string]any{"slack": map[string]any{"enabled": true}}
	app := NewApp(cfg, "", "test")

	hp := app.homePage.(*HomePage)
	hp.cursor = findMenuCursor(hp, PageNotifications)
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageNotifications {
		t.Errorf("currentPage = %v, want PageNotifications", app.currentPage)
	}
	if app.editorPage == nil {
		t.Error("editor page should be created for Notifications")
	}
}

func TestAppNavigateToCI(t *testing.T) {
	cfg := config.NewDefault()
	cfg.CI = map[string]any{"jenkins": map[string]any{"enabled": true}}
	app := NewApp(cfg, "", "test")

	hp := app.homePage.(*HomePage)
	hp.cursor = findMenuCursor(hp, PageCI)
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageCI {
		t.Errorf("currentPage = %v, want PageCI", app.currentPage)
	}
	if app.editorPage == nil {
		t.Error("editor page should be created for CI")
	}
}

func TestAppNavigateToPersistentStorage(t *testing.T) {
	cfg := config.NewDefault()
	cfg.PersistentStorage = map[string]any{"s3": map[string]any{"enabled": true}}
	app := NewApp(cfg, "", "test")

	hp := app.homePage.(*HomePage)
	hp.cursor = findMenuCursor(hp, PagePersistentStorage)
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PagePersistentStorage {
		t.Errorf("currentPage = %v, want PagePersistentStorage", app.currentPage)
	}
	if app.editorPage == nil {
		t.Error("editor page should be created for PersistentStorage")
	}
}

func TestAppNavigateToCanary(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Canary = map[string]any{"enabled": true}
	app := NewApp(cfg, "", "test")

	hp := app.homePage.(*HomePage)
	hp.cursor = findMenuCursor(hp, PageCanary)
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentPage != PageCanary {
		t.Errorf("currentPage = %v, want PageCanary", app.currentPage)
	}
	if app.editorPage == nil {
		t.Error("editor page should be created for Canary")
	}
}

func TestNewConfigSectionPageWithData(t *testing.T) {
	data := map[string]any{"github": map[string]any{"enabled": true}}
	p := newConfigSectionPage(data, "Artifacts")
	view := p.View()
	if !strings.Contains(view, "github") {
		t.Error("section page should show data keys")
	}
	if !strings.Contains(view, "Artifacts") {
		t.Error("section page should show label in breadcrumb")
	}
}

func TestNewConfigSectionPageEmpty(t *testing.T) {
	p := newConfigSectionPage(nil, "Empty")
	view := p.View()
	if !strings.Contains(view, "(empty)") {
		t.Errorf("empty section page view = %q, should contain '(empty)'", view)
	}
}
