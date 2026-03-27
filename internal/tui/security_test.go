package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spinnaker/spinctl/internal/config"
)

func TestSecurityPageView(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Security = config.SecurityConfig{
		Authn: config.AuthnConfig{Enabled: true},
		Authz: config.AuthzConfig{Enabled: true},
	}

	sp := NewSecurityPage(cfg)
	view := sp.View()
	if !strings.Contains(view, "authn") {
		t.Error("view should show 'authn'")
	}
	if !strings.Contains(view, "authz") {
		t.Error("view should show 'authz'")
	}
}

func TestSecurityPageEscAtRoot(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Security = config.SecurityConfig{
		Authn: config.AuthnConfig{Enabled: true},
	}

	sp := NewSecurityPage(cfg)
	_, cmd := sp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc at root should return a cmd")
	}
	// Execute the cmd to verify it returns goBackMsg.
	msg := cmd()
	if _, ok := msg.(goBackMsg); !ok {
		t.Errorf("expected goBackMsg, got %T", msg)
	}
}

func TestSecurityPageDrillAndEsc(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Security = config.SecurityConfig{
		Authn: config.AuthnConfig{Enabled: true},
		Authz: config.AuthzConfig{Enabled: false},
	}

	sp := NewSecurityPage(cfg)

	// Drill into authn (first item).
	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Now esc should go back within editor (not send goBackMsg).
	_, cmd := sp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	// After drilling in and pressing esc, should go back to editor root, no goBackMsg.
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(goBackMsg); ok {
			t.Error("esc inside editor should not send goBackMsg")
		}
	}
}

func TestSecurityPageNoConfig(t *testing.T) {
	cfg := config.NewDefault()
	// Security is zero-value, so toYAMLNode produces a node.
	// But let's test the page still renders.
	sp := NewSecurityPage(cfg)
	view := sp.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestSecurityPageNilEditor(t *testing.T) {
	cfg := config.NewDefault()
	sp := &SecurityPage{cfg: cfg, editor: nil}
	view := sp.View()
	if !strings.Contains(view, "No security configuration") {
		t.Error("nil editor should show 'No security configuration'")
	}
	if !strings.Contains(view, "esc: back") {
		t.Error("nil editor should show back hint")
	}
}

func TestSecurityPageNilEditorUpdate(t *testing.T) {
	cfg := config.NewDefault()
	sp := &SecurityPage{cfg: cfg, editor: nil}
	result, cmd := sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if result != sp {
		t.Error("should return same page")
	}
	if cmd != nil {
		t.Error("should return nil cmd")
	}
}
