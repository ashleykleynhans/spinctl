package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spinnaker/spinctl/internal/config"
)

func TestProvidersPageView(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Providers = map[string]config.ProviderConfig{
		"kubernetes": {Enabled: true, Accounts: []config.ProviderAccount{{Name: "prod"}}},
		"aws":        {Enabled: false},
	}

	pp := NewProvidersPage(cfg)
	view := pp.View()
	if !strings.Contains(view, "kubernetes") {
		t.Error("view should show 'kubernetes'")
	}
	if !strings.Contains(view, "aws") {
		t.Error("view should show 'aws'")
	}
	if !strings.Contains(view, "Providers") {
		t.Error("view should show title")
	}
}

func TestProvidersPageEmpty(t *testing.T) {
	cfg := config.NewDefault()
	// Providers is nil by default.

	pp := NewProvidersPage(cfg)
	view := pp.View()
	if !strings.Contains(view, "No providers configured") {
		t.Error("empty providers should show 'No providers configured'")
	}
}

func TestProvidersPageDrillIn(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Providers = map[string]config.ProviderConfig{
		"kubernetes": {Enabled: true, Accounts: []config.ProviderAccount{{Name: "prod"}}},
	}

	pp := NewProvidersPage(cfg)

	// Enter drills into the provider.
	pp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if pp.editor == nil {
		t.Error("enter should open the editor for the provider")
	}

	// View should now show editor content.
	view := pp.View()
	if !strings.Contains(view, "kubernetes") {
		t.Error("editor should show provider name in breadcrumb")
	}
}

func TestProvidersPageEsc(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Providers = map[string]config.ProviderConfig{
		"kubernetes": {Enabled: true},
	}

	pp := NewProvidersPage(cfg)

	// Esc on list should send goBackMsg.
	_, cmd := pp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should return a cmd")
	}
	msg := cmd()
	if _, ok := msg.(goBackMsg); !ok {
		t.Errorf("expected goBackMsg, got %T", msg)
	}
}

func TestProvidersPageNavigation(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Providers = map[string]config.ProviderConfig{
		"aws":        {Enabled: false},
		"gce":        {Enabled: false},
		"kubernetes": {Enabled: true},
	}

	pp := NewProvidersPage(cfg)
	if pp.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", pp.cursor)
	}

	// Move down.
	pp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if pp.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", pp.cursor)
	}

	// Move up.
	pp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if pp.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", pp.cursor)
	}

	// Move up at top should not go below 0.
	pp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if pp.cursor != 0 {
		t.Errorf("cursor after k at top = %d, want 0", pp.cursor)
	}
}

func TestProvidersPageNavigationWrapping(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Providers = map[string]config.ProviderConfig{
		"aws":        {Enabled: false},
		"gce":        {Enabled: false},
		"kubernetes": {Enabled: true},
	}

	pp := NewProvidersPage(cfg)

	// Move down to last.
	pp.cursor = len(pp.sortedNames) - 1
	pp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	// Should NOT wrap (no wrapping in providers).
	if pp.cursor != len(pp.sortedNames)-1 {
		t.Errorf("cursor should stay at last, got %d", pp.cursor)
	}
}

func TestProvidersPageEditorForwardsNonEscKey(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Providers = map[string]config.ProviderConfig{
		"kubernetes": {Enabled: true, Accounts: []config.ProviderAccount{{Name: "prod"}}},
	}

	pp := NewProvidersPage(cfg)
	// Drill in.
	pp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if pp.editor == nil {
		t.Fatal("should have editor")
	}

	// Send a non-esc key (e.g., down arrow).
	pp.Update(tea.KeyMsg{Type: tea.KeyDown})
	// Editor should still be active.
	if pp.editor == nil {
		t.Error("editor should remain active after non-esc key")
	}
}

func TestProvidersPageEditorEscReturnsToList(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Providers = map[string]config.ProviderConfig{
		"kubernetes": {Enabled: true},
	}

	pp := NewProvidersPage(cfg)
	// Drill in.
	pp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if pp.editor == nil {
		t.Fatal("should have editor after enter")
	}

	// Esc at editor root returns to list.
	pp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if pp.editor != nil {
		t.Error("esc at editor root should close editor")
	}
}
