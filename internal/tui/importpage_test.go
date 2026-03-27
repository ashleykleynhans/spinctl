package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/halimport"
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
	if !strings.Contains(view, "import") {
		t.Error("should show import option")
	}
}

func TestImportPageConfirm(t *testing.T) {
	ip := NewImportPage("/tmp/hal")
	result, cmd := ip.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	p := result.(*ImportPage)
	if !p.confirmed {
		t.Error("should be confirmed after 'y'")
	}
	if !p.importing {
		t.Error("should be importing after confirm")
	}
	if cmd == nil {
		t.Error("should return a command to run the import")
	}
}

func TestImportPageCancel(t *testing.T) {
	ip := NewImportPage("/tmp/hal")
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
	// If .hal exists, halDir is set and not editing. If not, halDir is empty and editing.
	if ip.halDir == "" {
		if !ip.editing {
			t.Error("should be in editing mode when no .hal detected")
		}
	} else {
		if ip.editing {
			t.Error("should not be in editing mode when .hal detected")
		}
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
	ip := NewImportPage("/tmp/hal")
	ip.Update(importDoneMsg{
		result: &halimport.ImportResult{
			DeploymentName: "default",
			BackupPath:     "/tmp/hal.backup.123",
		},
	})
	if !ip.done {
		t.Error("should be done after importDoneMsg")
	}
	view := ip.View()
	if !strings.Contains(view, "complete") {
		t.Error("done page should show 'complete'")
	}
	if !strings.Contains(view, "default") {
		t.Error("should show deployment name")
	}
}

func TestImportPageDoneError(t *testing.T) {
	ip := NewImportPage("/tmp/hal")
	ip.Update(importDoneMsg{err: fmt.Errorf("parse error")})
	if !ip.done {
		t.Error("should be done after error")
	}
	view := ip.View()
	if !strings.Contains(view, "failed") {
		t.Error("done with error should show 'failed'")
	}
	if !strings.Contains(view, "parse error") {
		t.Error("done with error should show the error message")
	}
}

func TestImportPageCancelWithEsc(t *testing.T) {
	ip := NewImportPage("/tmp/hal")
	ip.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if !ip.cancelled {
		t.Error("should be cancelled after esc")
	}
}

func TestImportPageIgnoresKeysAfterImporting(t *testing.T) {
	ip := NewImportPage("/tmp/hal")
	ip.confirmed = true
	ip.importing = true
	ip.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if ip.cancelled {
		t.Error("should not cancel during importing")
	}
}

func TestImportPageNonKeyMessage(t *testing.T) {
	ip := NewImportPage("/tmp/hal")
	type customMsg struct{}
	result, cmd := ip.Update(customMsg{})
	if result != ip {
		t.Error("non-key message should return same page")
	}
	if cmd != nil {
		t.Error("non-key message should return nil cmd")
	}
}

func TestImportPageEditPath(t *testing.T) {
	ip := NewImportPage("/tmp/hal")

	// Press e to enter edit mode.
	ip.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if !ip.editing {
		t.Fatal("should be in editing mode after 'e'")
	}

	view := ip.View()
	if !strings.Contains(view, "confirm path") {
		t.Error("editing view should show 'confirm path'")
	}

	// Clear and type new path.
	for range len(ip.editBuffer) {
		ip.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	ip.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/new/path")})
	ip.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if ip.editing {
		t.Error("should exit editing after enter")
	}
	if ip.halDir != "/new/path" {
		t.Errorf("halDir = %q, want /new/path", ip.halDir)
	}
}

func TestImportPageEditPathCancel(t *testing.T) {
	ip := NewImportPage("/tmp/hal")
	ip.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	ip.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	ip.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if ip.editing {
		t.Error("should exit editing after esc")
	}
	if ip.halDir != "/tmp/hal" {
		t.Errorf("halDir = %q, want /tmp/hal (unchanged)", ip.halDir)
	}
}

func TestImportPageWithExplicitPath(t *testing.T) {
	ip := NewImportPage("/custom/hal")
	if ip.halDir != "/custom/hal" {
		t.Errorf("halDir = %q, want /custom/hal", ip.halDir)
	}
	if ip.editing {
		t.Error("should not start in editing mode when path is provided")
	}
}

func TestImportPageWithExplicitExistingPath(t *testing.T) {
	tmpDir := t.TempDir()
	ip := NewImportPage(tmpDir)
	if ip.halDir != tmpDir {
		t.Errorf("halDir = %q, want %q", ip.halDir, tmpDir)
	}
	if ip.editing {
		t.Error("should not be in editing mode when explicit path is provided")
	}
	if ip.editBuffer != tmpDir {
		t.Errorf("editBuffer = %q, want %q", ip.editBuffer, tmpDir)
	}
}

func TestImportPageDoneSuccessWithUnmapped(t *testing.T) {
	ip := NewImportPage("/tmp/hal")
	ip.Update(importDoneMsg{
		result: &halimport.ImportResult{
			DeploymentName: "default",
			BackupPath:     "/tmp/hal.backup.123",
			UnmappedFields: []string{"customField1", "customField2"},
		},
	})
	if !ip.done {
		t.Error("should be done")
	}
	if !strings.Contains(ip.result, "customField1") {
		t.Error("result should include unmapped fields")
	}
}
