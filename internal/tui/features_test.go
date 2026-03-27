package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spinnaker/spinctl/internal/config"
)

func TestFeaturesPageView(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Features["artifacts"] = true
	cfg.Features["chaos"] = false

	fp := NewFeaturesPage(cfg)
	view := fp.View()
	if !strings.Contains(view, "artifacts") {
		t.Error("view should show 'artifacts' feature")
	}
	if !strings.Contains(view, "chaos") {
		t.Error("view should show 'chaos' feature")
	}
	if !strings.Contains(view, "Features") {
		t.Error("view should show 'Features' title")
	}
}

func TestFeaturesPageToggle(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Features["artifacts"] = false

	fp := NewFeaturesPage(cfg)
	_, cmd := fp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("toggle should return a configChangedMsg cmd")
	}
	if !cfg.Features["artifacts"] {
		t.Error("artifacts should be true after toggle")
	}

	// Toggle again.
	fp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cfg.Features["artifacts"] {
		t.Error("artifacts should be false after second toggle")
	}
}

func TestFeaturesPageEmpty(t *testing.T) {
	cfg := config.NewDefault()
	// Features map is empty by default.

	fp := NewFeaturesPage(cfg)
	view := fp.View()
	if !strings.Contains(view, "No feature flags") {
		t.Error("empty features should show 'No feature flags' message")
	}
}

func TestFeaturesPageNavigation(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Features["alpha"] = true
	cfg.Features["beta"] = false
	cfg.Features["gamma"] = true

	fp := NewFeaturesPage(cfg)
	if fp.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", fp.cursor)
	}

	// Move down.
	fp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if fp.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", fp.cursor)
	}

	// Move down again.
	fp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if fp.cursor != 2 {
		t.Errorf("cursor after j = %d, want 2", fp.cursor)
	}

	// Move down at end should not go past.
	fp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if fp.cursor != 2 {
		t.Errorf("cursor after j at end = %d, want 2", fp.cursor)
	}

	// Move up.
	fp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if fp.cursor != 1 {
		t.Errorf("cursor after k = %d, want 1", fp.cursor)
	}

	// Move up to 0.
	fp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if fp.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", fp.cursor)
	}

	// Move up at top should not go past.
	fp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if fp.cursor != 0 {
		t.Errorf("cursor after k at top = %d, want 0", fp.cursor)
	}
}
