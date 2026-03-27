package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
)

// SecurityPage displays authentication and authorization settings
// via the YAML editor for full configuration access.
type SecurityPage struct {
	cfg    *config.SpinctlConfig
	editor *EditorPage
}

// NewSecurityPage creates a security settings page.
func NewSecurityPage(cfg *config.SpinctlConfig) *SecurityPage {
	node, _ := toYAMLNode(cfg.Security)
	var editor *EditorPage
	if node != nil {
		editor = NewEditorPage(node, "Security")
	}
	return &SecurityPage{cfg: cfg, editor: editor}
}

func (s *SecurityPage) Update(msg tea.Msg) (page, tea.Cmd) {
	if s.editor != nil {
		var cmd tea.Cmd
		_, cmd = s.editor.Update(msg)
		return s, cmd
	}
	return s, nil
}

func (s *SecurityPage) View() string {
	if s.editor != nil {
		return s.editor.View()
	}
	var b strings.Builder
	b.WriteString("\n  Security\n\n")
	b.WriteString("  No security configuration.\n")
	b.WriteString("\n  esc: back\n")
	return b.String()
}
