package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestImportPageShowsStatus(t *testing.T) {
	ip := NewImportPage("/tmp/hal")
	view := ip.View()
	if !strings.Contains(view, "/tmp/hal") {
		t.Error("should show source directory")
	}
	if !strings.Contains(view, "backup") {
		t.Error("should mention backup")
	}
	if !strings.Contains(view, "y/n") {
		t.Error("should show confirmation prompt")
	}
}

func TestImportPageConfirm(t *testing.T) {
	ip := NewImportPage("")
	ip.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if !ip.confirmed {
		t.Error("should be confirmed after 'y'")
	}
	if !ip.importing {
		t.Error("should be importing after confirm")
	}
}

func TestImportPageCancel(t *testing.T) {
	ip := NewImportPage("")
	ip.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !ip.cancelled {
		t.Error("should be cancelled after 'n'")
	}
	view := ip.View()
	if !strings.Contains(view, "cancelled") {
		t.Error("should show cancelled message")
	}
}

func TestImportPageDefaultDir(t *testing.T) {
	ip := NewImportPage("")
	if ip.halDir != "~/.hal" {
		t.Errorf("default dir = %q, want ~/.hal", ip.halDir)
	}
}
