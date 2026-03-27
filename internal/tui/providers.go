package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
)

// ProvidersPage displays cloud provider configurations.
type ProvidersPage struct {
	cfg         *config.SpinctlConfig
	sortedNames []string
	cursor      int
}

// NewProvidersPage creates a providers page.
func NewProvidersPage(cfg *config.SpinctlConfig) *ProvidersPage {
	names := make([]string, 0, len(cfg.Providers))
	for name := range cfg.Providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return &ProvidersPage{cfg: cfg, sortedNames: names}
}

func (p *ProvidersPage) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.cursor > 0 {
				p.cursor--
			}
		case "down", "j":
			if p.cursor < len(p.sortedNames)-1 {
				p.cursor++
			}
		}
	}
	return p, nil
}

func (p *ProvidersPage) View() string {
	var b strings.Builder
	b.WriteString("\n  Providers\n\n")

	if len(p.sortedNames) == 0 {
		b.WriteString("  No providers configured.\n")
		b.WriteString("\n  esc: back\n")
		return b.String()
	}

	for i, name := range p.sortedNames {
		prov := p.cfg.Providers[name]
		cursor := "  "
		if i == p.cursor {
			cursor = "▸ "
		}
		status := "OFF"
		if prov.Enabled {
			status = " ON"
		}
		acctCount := len(prov.Accounts)
		b.WriteString(fmt.Sprintf("%s[%s] %-20s %d account(s)\n", cursor, status, name, acctCount))
	}

	b.WriteString("\n  esc: back\n")
	return b.String()
}
