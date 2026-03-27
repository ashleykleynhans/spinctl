package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

// ServicesPage displays the list of Spinnaker services with toggle support.
type ServicesPage struct {
	cfg         *config.SpinctlConfig
	sortedNames []model.ServiceName
	cursor      int
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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
		case "enter":
			if s.cursor >= 0 && s.cursor < len(s.sortedNames) {
				name := s.sortedNames[s.cursor]
				svc := s.cfg.Services[name]
				svc.Enabled = !svc.Enabled
				s.cfg.Services[name] = svc
			}
		}
	}
	return s, nil
}

// View renders the services list.
func (s *ServicesPage) View() string {
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
	b.WriteString("\n  enter: toggle  esc: back\n")
	return b.String()
}
