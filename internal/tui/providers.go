package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
)

// ProvidersPage displays cloud provider configurations.
// Enter drills into a provider's full config via the editor.
type ProvidersPage struct {
	cfg         *config.SpinctlConfig
	sortedNames []string
	cursor      int
	editor      *EditorPage
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
	if p.editor != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" && len(p.editor.nodeStack) == 0 {
				p.editor = nil
				return p, nil
			}
		}
		var cmd tea.Cmd
		_, cmd = p.editor.Update(msg)
		return p, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return p, func() tea.Msg { return goBackMsg{} }
		case "up", "k":
			if p.cursor > 0 {
				p.cursor--
			}
		case "down", "j":
			if p.cursor < len(p.sortedNames)-1 {
				p.cursor++
			}
		case "enter":
			if p.cursor >= 0 && p.cursor < len(p.sortedNames) {
				name := p.sortedNames[p.cursor]
				prov := p.cfg.Providers[name]
				node, err := toYAMLNode(prov)
				if err == nil {
					p.editor = NewEditorPage(node, name)
				}
			}
		}
	}
	return p, nil
}

func (p *ProvidersPage) View() string {
	if p.editor != nil {
		return p.editor.View()
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(headingStyle.Render("Providers"))
	b.WriteString("\n\n")

	if len(p.sortedNames) == 0 {
		b.WriteString("  " + valueStyle.Render("No providers configured.") + "\n")
		b.WriteString("\n  " + menuDescStyle.Render("esc: back") + "\n")
		return b.String()
	}

	for i, name := range p.sortedNames {
		prov := p.cfg.Providers[name]
		selected := i == p.cursor
		cursor := "  "
		if selected {
			cursor = menuCursorStyle.Render("▸ ")
		}
		status := offStyle.Render("[OFF]")
		if prov.Enabled {
			status = onStyle.Render("[ ON]")
		}
		label := keyStyle.Render(fmt.Sprintf("%-20s", name))
		if selected {
			label = keySelectedStyle.Render(fmt.Sprintf("%-20s", name))
		}
		acctCount := len(prov.Accounts)
		info := valueStyle.Render(fmt.Sprintf("%d account(s)", acctCount))
		b.WriteString(fmt.Sprintf("%s%s %s  %s\n", cursor, status, label, info))
	}

	b.WriteString("\n  " + menuDescStyle.Render("enter: configure  esc: back") + "\n")
	return b.String()
}
