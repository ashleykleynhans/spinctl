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
	app := NewApp(cfg, "")
	if app.currentPage != PageHome {
		t.Errorf("initial page = %v, want PageHome", app.currentPage)
	}
}

func TestAppQuitOnQ(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "")
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestAppViewContainsTitle(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	app := NewApp(cfg, "")
	view := app.View()
	if !strings.Contains(view, "spinctl") {
		t.Error("view should contain 'spinctl'")
	}
}

func TestAppWindowResize(t *testing.T) {
	cfg := config.NewDefault()
	app := NewApp(cfg, "")
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if app.width != 120 || app.height != 40 {
		t.Errorf("size = %dx%d, want 120x40", app.width, app.height)
	}
}
