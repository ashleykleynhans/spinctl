package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

// ServicesPage displays the list of Spinnaker services.
// Enter drills into a service's full configuration via the editor.
// Space toggles enabled/disabled.
type ServicesPage struct {
	cfg         *config.SpinctlConfig
	sortedNames []model.ServiceName
	cursor      int
	editor      *EditorPage
	editingName model.ServiceName
}

// NewServicesPage creates a services page with alphabetically sorted services.
func NewServicesPage(cfg *config.SpinctlConfig) *ServicesPage {
	names := make([]model.ServiceName, 0, len(cfg.Services))
	for name := range cfg.Services {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return names[i].String() < names[j].String()
	})
	return &ServicesPage{
		cfg:         cfg,
		sortedNames: names,
	}
}

// Update handles input for the services page.
func (s *ServicesPage) Update(msg tea.Msg) (page, tea.Cmd) {
	// If we're in the editor, delegate to it.
	if s.editor != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" && len(s.editor.nodeStack) == 0 {
				// At editor root level, esc goes back to service list.
				s.editor = nil
				return s, nil
			}
		}
		var cmd tea.Cmd
		_, cmd = s.editor.Update(msg)
		return s, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, func() tea.Msg { return goBackMsg{} }
		case "up", "k":
			s.cursor--
			if s.cursor < 0 {
				s.cursor = len(s.sortedNames) - 1
			}
		case "down", "j":
			s.cursor++
			if s.cursor >= len(s.sortedNames) {
				s.cursor = 0
			}
		case " ":
			// Space toggles enabled/disabled.
			if s.cursor >= 0 && s.cursor < len(s.sortedNames) {
				name := s.sortedNames[s.cursor]
				svc := s.cfg.Services[name]
				svc.Enabled = !svc.Enabled
				s.cfg.Services[name] = svc
				return s, func() tea.Msg { return configChangedMsg{} }
			}
		case "enter":
			// Enter drills into service config.
			if s.cursor >= 0 && s.cursor < len(s.sortedNames) {
				name := s.sortedNames[s.cursor]
				svc := s.cfg.Services[name]
				node, err := toYAMLNode(svc)
				if err == nil {
					s.editor = NewEditorPage(node, name.String())
					s.editingName = name
				}
			}
		}
	}
	return s, nil
}

// View renders the services list or the editor if drilling in.
func (s *ServicesPage) View() string {
	if s.editor != nil {
		return s.editor.View()
	}

	var b strings.Builder
	b.WriteString("\n  Services\n\n")
	for i, name := range s.sortedNames {
		svc := s.cfg.Services[name]
		cursor := "  "
		if i == s.cursor {
			cursor = "▸ "
		}
		status := "OFF"
		if svc.Enabled {
			status = " ON"
		}
		b.WriteString(fmt.Sprintf("%s[%s] %-15s %s:%d\n", cursor, status, name, svc.Host, svc.Port))
	}
	b.WriteString("\n  enter: configure  space: toggle  esc: back\n")
	return b.String()
}
