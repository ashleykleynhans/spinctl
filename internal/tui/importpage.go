package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ImportPage is the import wizard for importing Halyard configurations.
type ImportPage struct {
	halDir    string
	confirmed bool
	cancelled bool
	importing bool
	done      bool
	result    string
	err       error
}

// NewImportPage creates an import page targeting the given halconfig directory.
func NewImportPage(halDir string) *ImportPage {
	if halDir == "" {
		halDir = "~/.hal"
	}
	return &ImportPage{
		halDir: halDir,
	}
}

// Update handles input for the import page.
func (p *ImportPage) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !p.confirmed && !p.cancelled && !p.importing {
			switch msg.String() {
			case "y":
				p.confirmed = true
				p.importing = true
			case "n", "esc":
				p.cancelled = true
			}
		}
	}
	return p, nil
}

// View renders the import page.
func (p *ImportPage) View() string {
	var b strings.Builder
	b.WriteString("\n  Import from Halyard\n\n")
	b.WriteString(fmt.Sprintf("  Source: %s\n", p.halDir))
	b.WriteString("  A backup will be created before importing.\n\n")

	if p.cancelled {
		b.WriteString("  Import cancelled.\n")
	} else if p.done {
		if p.err != nil {
			b.WriteString(fmt.Sprintf("  Import failed: %s\n", p.err))
		} else {
			b.WriteString(fmt.Sprintf("  Import complete. %s\n", p.result))
		}
	} else if p.importing {
		b.WriteString("  Importing...\n")
	} else {
		b.WriteString("  Proceed with import? (y/n)\n")
	}

	return b.String()
}
