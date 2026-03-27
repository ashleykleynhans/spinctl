package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spinnaker/spinctl/internal/config"
)

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

	// Navigate to Import from Halyard (index 6, after separator).
	hp := app.homePage.(*HomePage)
	hp.cursor = 6
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

	// Navigate to Deploy (index 7).
	hp := app.homePage.(*HomePage)
	hp.cursor = 7
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
	hp.cursor = 6 // Import from Halyard
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
	hp.cursor = 7 // Deploy
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

func TestAppViewAllPages(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"

	// Test each page renders without panic.
	pages := []struct {
		name   string
		cursor int
	}{
		{"Services", 0},
		{"Import", 6},
		{"Deploy", 7},
	}
	for _, tc := range pages {
		t.Run(tc.name, func(t *testing.T) {
			app := NewApp(cfg, "", "test")
			hp := app.homePage.(*HomePage)
			hp.cursor = tc.cursor
			app.Update(tea.KeyMsg{Type: tea.KeyEnter})
			view := app.View()
			if view == "" {
				t.Errorf("view for %s should not be empty", tc.name)
			}
		})
	}
}
