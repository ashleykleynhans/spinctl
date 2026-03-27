package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/halimport"
)

// importDoneMsg is sent when the import completes.
type importDoneMsg struct {
	result *halimport.ImportResult
	err    error
}

// ImportPage is the import wizard for importing Halyard configurations.
type ImportPage struct {
	halDir     string
	outputPath string
	editing    bool
	editBuffer string
	confirmed  bool
	cancelled  bool
	importing  bool
	done       bool
	result     string
	err        error
}

// NewImportPage creates an import page targeting the given halconfig directory.
func NewImportPage(halDir string) *ImportPage {
	if halDir == "" {
		detected := halimport.DetectHalDir()
		if detected != "" {
			halDir = detected
		} else {
			halDir = ""
		}
	}
	return &ImportPage{
		halDir:     halDir,
		editBuffer: halDir,
		editing:    halDir == "",
		outputPath: config.DefaultConfigPath(),
	}
}

// Update handles input for the import page.
func (p *ImportPage) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case importDoneMsg:
		p.importing = false
		p.done = true
		if msg.err != nil {
			p.err = msg.err
		} else {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Deployment: %s\n", msg.result.DeploymentName))
			sb.WriteString(fmt.Sprintf("  Backup: %s\n", msg.result.BackupPath))
			if len(msg.result.UnmappedFields) > 0 {
				sb.WriteString(fmt.Sprintf("  Unmapped fields: %s\n", strings.Join(msg.result.UnmappedFields, ", ")))
			}
			sb.WriteString(fmt.Sprintf("  Config saved to: %s", p.outputPath))
			p.result = sb.String()
		}
		return p, nil

	case tea.KeyMsg:
		if p.editing {
			return p.updateEditing(msg)
		}
		if !p.confirmed && !p.cancelled && !p.importing && !p.done {
			switch msg.String() {
			case "y":
				p.confirmed = true
				p.importing = true
				return p, p.runImport()
			case "e":
				p.editing = true
				p.editBuffer = p.halDir
			case "n", "esc":
				p.cancelled = true
			}
		}
	}
	return p, nil
}

func (p *ImportPage) updateEditing(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		p.halDir = p.editBuffer
		p.editing = false
	case tea.KeyEscape:
		p.editBuffer = p.halDir
		p.editing = false
	case tea.KeyBackspace:
		if len(p.editBuffer) > 0 {
			p.editBuffer = p.editBuffer[:len(p.editBuffer)-1]
		}
	case tea.KeyRunes:
		p.editBuffer += string(msg.Runes)
	}
	return p, nil
}

func (p *ImportPage) runImport() tea.Cmd {
	return func() tea.Msg {
		result, err := halimport.Import(halimport.ImportOptions{
			HalDir:         p.halDir,
			DeploymentName: "default",
			OutputPath:     p.outputPath,
		})
		return importDoneMsg{result: result, err: err}
	}
}

// View renders the import page.
func (p *ImportPage) View() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(headingStyle.Render("Import from Halyard"))
	b.WriteString("\n\n")

	if p.editing {
		b.WriteString("  " + keyStyle.Render("Source: ") + editCursorStyle.Render(p.editBuffer+"█") + "\n\n")
		b.WriteString("  " + menuDescStyle.Render("enter: confirm path  esc: cancel") + "\n")
		return b.String()
	}

	b.WriteString("  " + keyStyle.Render("Source: ") + onStyle.Render(p.halDir) + "\n")
	b.WriteString("  " + menuDescStyle.Render("A backup will be created before importing.") + "\n\n")

	if p.cancelled {
		b.WriteString("  " + warnStyle.Render("Import cancelled.") + "\n")
	} else if p.done {
		if p.err != nil {
			b.WriteString("  " + warnStyle.Render(fmt.Sprintf("Import failed: %s", p.err)) + "\n")
		} else {
			b.WriteString("  " + successStyle.Render("Import complete.") + "\n  " + p.result + "\n")
		}
	} else if p.importing {
		b.WriteString("  " + keyStyle.Render("Importing...") + "\n")
	} else {
		b.WriteString("  " + keyStyle.Render("y") + menuDescStyle.Render(": import  ") +
			keyStyle.Render("e") + menuDescStyle.Render(": edit path  ") +
			keyStyle.Render("esc") + menuDescStyle.Render(": cancel") + "\n")
	}

	return b.String()
}
