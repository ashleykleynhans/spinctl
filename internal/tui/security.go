package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
)

// SecurityPage displays authentication and authorization settings.
type SecurityPage struct {
	cfg    *config.SpinctlConfig
	cursor int
}

// NewSecurityPage creates a security settings page.
func NewSecurityPage(cfg *config.SpinctlConfig) *SecurityPage {
	return &SecurityPage{cfg: cfg}
}

func (s *SecurityPage) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < 1 {
				s.cursor++
			}
		case "enter":
			switch s.cursor {
			case 0:
				s.cfg.Security.Authn.Enabled = !s.cfg.Security.Authn.Enabled
			case 1:
				s.cfg.Security.Authz.Enabled = !s.cfg.Security.Authz.Enabled
			}
		}
	}
	return s, nil
}

func (s *SecurityPage) View() string {
	var b strings.Builder
	b.WriteString("\n  Security\n\n")

	items := []struct {
		label   string
		enabled bool
	}{
		{"Authentication (authn)", s.cfg.Security.Authn.Enabled},
		{"Authorization (authz)", s.cfg.Security.Authz.Enabled},
	}

	for i, item := range items {
		cursor := "  "
		if i == s.cursor {
			cursor = "▸ "
		}
		status := "OFF"
		if item.enabled {
			status = " ON"
		}
		b.WriteString(fmt.Sprintf("%s[%s] %s\n", cursor, status, item.label))
	}

	b.WriteString("\n  enter: toggle  esc: back\n")
	return b.String()
}
