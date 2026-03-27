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
	halDir    string
	outputPath string
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
		detected := halimport.DetectHalDir()
		if detected != "" {
			halDir = detected
		} else {
			halDir = "~/.hal"
		}
	}
	return &ImportPage{
		halDir:     halDir,
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
		if !p.confirmed && !p.cancelled && !p.importing {
			switch msg.String() {
			case "y":
				p.confirmed = true
				p.importing = true
				return p, p.runImport()
			case "n", "esc":
				p.cancelled = true
			}
		}
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
	b.WriteString("  " + keyStyle.Render("Source: ") + valueStyle.Render(p.halDir) + "\n")
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
		b.WriteString("  " + keyStyle.Render("Proceed with import? ") + menuDescStyle.Render("(y/n)") + "\n")
	}

	return b.String()
}
