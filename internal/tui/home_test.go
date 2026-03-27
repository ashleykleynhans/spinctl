package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spinnaker/spinctl/internal/config"
)

func TestHomePageView(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)
	view := hp.View()
	if !strings.Contains(view, "Services") {
		t.Error("home page should show Services")
	}
	if !strings.Contains(view, "Deploy") {
		t.Error("home page should show Deploy")
	}
}

func TestHomePageNavigation(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)

	if hp.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", hp.cursor)
	}

	// Move down.
	hp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if hp.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", hp.cursor)
	}

	// Move up.
	hp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if hp.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", hp.cursor)
	}
}

func TestHomePageSelect(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)

	// First item is Services.
	hp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if hp.selected != PageServices {
		t.Errorf("selected = %v, want PageServices", hp.selected)
	}
}

func TestHomePageSkipsSeparator(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)

	// Navigate to item before separator (index 4 = Version).
	hp.cursor = 4
	hp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	// Should skip separator (index 5) and land on index 6.
	if hp.cursor != 6 {
		t.Errorf("cursor should skip separator, got %d, want 6", hp.cursor)
	}
}
