package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spinnaker/spinctl/internal/config"
)

// findSeparatorIndex returns the index of the first separator in the menu.
func findSeparatorIndex(hp *HomePage) int {
	for i, item := range hp.items {
		if item.separator {
			return i
		}
	}
	return -1
}

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

	hp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if hp.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", hp.cursor)
	}

	hp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if hp.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", hp.cursor)
	}
}

func TestHomePageSelect(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)

	hp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if hp.selected != PageServices {
		t.Errorf("selected = %v, want PageServices", hp.selected)
	}
}

func TestHomePageSkipsSeparator(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)

	sepIdx := findSeparatorIndex(hp)
	hp.cursor = sepIdx - 1
	hp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if hp.cursor != sepIdx+1 {
		t.Errorf("cursor should skip separator, got %d, want %d", hp.cursor, sepIdx+1)
	}
}

func TestHomePageSkipsSeparatorGoingUp(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)

	sepIdx := findSeparatorIndex(hp)
	hp.cursor = sepIdx + 1
	hp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if hp.cursor != sepIdx-1 {
		t.Errorf("cursor should skip separator going up, got %d, want %d", hp.cursor, sepIdx-1)
	}
}

func TestHomePageWrapAroundDown(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)

	hp.cursor = len(hp.items) - 1
	hp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if hp.cursor != 0 {
		t.Errorf("cursor should wrap to 0, got %d", hp.cursor)
	}
}

func TestHomePageWrapAroundUp(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)

	hp.cursor = 0
	hp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	last := len(hp.items) - 1
	if hp.items[last].separator {
		last--
	}
	if hp.cursor != last {
		t.Errorf("cursor should wrap to %d, got %d", last, hp.cursor)
	}
}

func TestHomePageUpDownKeys(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)

	hp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if hp.cursor != 1 {
		t.Errorf("cursor after down = %d, want 1", hp.cursor)
	}

	hp.Update(tea.KeyMsg{Type: tea.KeyUp})
	if hp.cursor != 0 {
		t.Errorf("cursor after up = %d, want 0", hp.cursor)
	}
}

func TestHomePageViewShowsSeparator(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)
	view := hp.View()
	if !strings.Contains(view, "·") {
		t.Error("home page should show separator")
	}
}

func TestHomePageViewShowsCursor(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)
	view := hp.View()
	if !strings.Contains(view, "\u25b8") {
		t.Error("home page should show cursor indicator")
	}
}

func TestHomePageSelectDoesNotSelectSeparator(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)
	sepIdx := findSeparatorIndex(hp)
	hp.cursor = sepIdx
	hp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if hp.selected != 0 {
		t.Errorf("should not select separator, selected = %v", hp.selected)
	}
}

func TestHomePageDownArrowWrapping(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)

	// Navigate to last item.
	hp.cursor = len(hp.items) - 1
	// Ensure it's not a separator.
	if hp.items[hp.cursor].separator {
		hp.cursor--
	}
	// Press down arrow to wrap.
	hp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if hp.cursor != 0 {
		t.Errorf("cursor should wrap to 0, got %d", hp.cursor)
	}
}

func TestHomePageNonKeyMessage(t *testing.T) {
	cfg := config.NewDefault()
	hp := NewHomePage(cfg)
	type customMsg struct{}
	result, cmd := hp.Update(customMsg{})
	if result != hp {
		t.Error("non-key message should return same page")
	}
	if cmd != nil {
		t.Error("non-key message should return nil cmd")
	}
}
