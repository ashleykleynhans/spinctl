package tui

import (
	"fmt"
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

func TestImportPageImportingView(t *testing.T) {
	ip := NewImportPage("/tmp/hal")
	ip.confirmed = true
	ip.importing = true
	view := ip.View()
	if !strings.Contains(view, "Importing") {
		t.Error("should show 'Importing' during import")
	}
}

func TestImportPageDoneSuccess(t *testing.T) {
	ip := NewImportPage("")
	ip.done = true
	ip.result = "5 services imported"
	view := ip.View()
	if !strings.Contains(view, "complete") {
		t.Error("done page should show 'complete'")
	}
	if !strings.Contains(view, "5 services imported") {
		t.Error("done page should show result message")
	}
}

func TestImportPageDoneError(t *testing.T) {
	ip := NewImportPage("")
	ip.done = true
	ip.err = fmt.Errorf("parse error")
	view := ip.View()
	if !strings.Contains(view, "failed") {
		t.Error("done with error should show 'failed'")
	}
	if !strings.Contains(view, "parse error") {
		t.Error("done with error should show the error message")
	}
}

func TestImportPageCancelWithEsc(t *testing.T) {
	ip := NewImportPage("")
	ip.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if !ip.cancelled {
		t.Error("should be cancelled after esc")
	}
}

func TestImportPageIgnoresKeysAfterImporting(t *testing.T) {
	ip := NewImportPage("")
	ip.confirmed = true
	ip.importing = true
	ip.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if ip.cancelled {
		t.Error("should not cancel during importing")
	}
}

func TestImportPageNonKeyMessage(t *testing.T) {
	ip := NewImportPage("")
	type customMsg struct{}
	result, cmd := ip.Update(customMsg{})
	if result != ip {
		t.Error("non-key message should return same page")
	}
	if cmd != nil {
		t.Error("non-key message should return nil cmd")
	}
}
