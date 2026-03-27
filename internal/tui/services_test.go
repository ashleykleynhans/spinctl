package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

func TestServicesPageShowsAllServices(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)
	view := sp.View()
	for _, name := range model.AllServiceNames() {
		if !strings.Contains(view, name.String()) {
			t.Errorf("services page should show %s", name)
		}
	}
}

func TestServicesPageNavigation(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)

	if sp.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", sp.cursor)
	}

	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if sp.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", sp.cursor)
	}

	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if sp.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", sp.cursor)
	}
}

func TestServicesPageToggle(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)

	name := sp.sortedNames[0]
	if cfg.Services[name].Enabled {
		t.Error("service should start disabled")
	}

	// Space toggles enabled/disabled.
	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !cfg.Services[name].Enabled {
		t.Error("service should be enabled after toggle")
	}

	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if cfg.Services[name].Enabled {
		t.Error("service should be disabled after second toggle")
	}
}

func TestServicesPageEnterOpensEditor(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)

	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if sp.editor == nil {
		t.Error("enter should open the editor for the service")
	}
}

func TestServicesPageEnabledDisplay(t *testing.T) {
	cfg := config.NewDefault()
	// Enable one service.
	svc := cfg.Services[model.Gate]
	svc.Enabled = true
	cfg.Services[model.Gate] = svc

	sp := NewServicesPage(cfg)
	view := sp.View()
	if !strings.Contains(view, " ON") {
		t.Error("should display ON for enabled service")
	}
	if !strings.Contains(view, "OFF") {
		t.Error("should display OFF for disabled services")
	}
}

func TestServicesPageSorted(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)

	for i := 1; i < len(sp.sortedNames); i++ {
		if sp.sortedNames[i].String() < sp.sortedNames[i-1].String() {
			t.Errorf("services not sorted: %s before %s", sp.sortedNames[i-1], sp.sortedNames[i])
		}
	}
}

func TestServicesPageWrapDown(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)

	sp.cursor = len(sp.sortedNames) - 1
	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if sp.cursor != 0 {
		t.Errorf("cursor should wrap to 0, got %d", sp.cursor)
	}
}

func TestServicesPageWrapUp(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)

	sp.cursor = 0
	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	expected := len(sp.sortedNames) - 1
	if sp.cursor != expected {
		t.Errorf("cursor should wrap to %d, got %d", expected, sp.cursor)
	}
}

func TestServicesPageUpDownKeys(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)

	sp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if sp.cursor != 1 {
		t.Errorf("cursor after down = %d, want 1", sp.cursor)
	}

	sp.Update(tea.KeyMsg{Type: tea.KeyUp})
	if sp.cursor != 0 {
		t.Errorf("cursor after up = %d, want 0", sp.cursor)
	}
}

func TestServicesPageNonKeyMessage(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)
	type customMsg struct{}
	result, cmd := sp.Update(customMsg{})
	if result != sp {
		t.Error("non-key message should return same page")
	}
	if cmd != nil {
		t.Error("non-key message should return nil cmd")
	}
}

func TestServicesPageViewShowsHints(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)
	view := sp.View()
	if !strings.Contains(view, "enter: configure") {
		t.Error("services page should show configure hint")
	}
	if !strings.Contains(view, "esc: back") {
		t.Error("services page should show back hint")
	}
}

func TestServicesPageEditorEscReturnsToList(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)

	// Enter to open editor.
	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if sp.editor == nil {
		t.Fatal("should have editor after enter")
	}

	// Esc at editor root returns to service list.
	sp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if sp.editor != nil {
		t.Error("esc at editor root should close editor and return to list")
	}
}

func TestServicesPageEditorForwardsNonEscKey(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)

	// Enter to open editor.
	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if sp.editor == nil {
		t.Fatal("should have editor after enter")
	}

	// Send a non-esc key to the editor.
	sp.Update(tea.KeyMsg{Type: tea.KeyDown})
	// Editor should still be active.
	if sp.editor == nil {
		t.Error("editor should remain active after non-esc key")
	}
}

func TestServicesPageEscSendsGoBack(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)

	// Esc on the service list should send goBackMsg.
	_, cmd := sp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should return a cmd")
	}
	msg := cmd()
	if _, ok := msg.(goBackMsg); !ok {
		t.Errorf("expected goBackMsg, got %T", msg)
	}
}
