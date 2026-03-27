package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
)

type menuItem struct {
	label       string
	description string
	action      PageID
	separator   bool
}

// HomePage displays the main menu.
type HomePage struct {
	cfg      *config.SpinctlConfig
	items    []menuItem
	cursor   int
	selected PageID
}

// NewHomePage creates a home page with menu items derived from the config.
func NewHomePage(cfg *config.SpinctlConfig) *HomePage {
	serviceCount := len(cfg.Services)
	providerCount := len(cfg.Providers)

	items := []menuItem{
		{label: fmt.Sprintf("Services (%d)", serviceCount), description: "Configure Spinnaker services", action: PageServices},
		{label: fmt.Sprintf("Providers (%d)", providerCount), description: "Configure cloud providers", action: PageProviders},
		{label: "Security", description: "Authentication & authorization", action: PageSecurity},
		{label: "Features", description: "Feature flags", action: PageFeatures},
		{label: fmt.Sprintf("Version: %s", cfg.Version), description: "Spinnaker version", action: PageVersion},
		{separator: true},
		{label: "Import from Halyard", description: "Import existing halconfig", action: PageImport},
		{label: "Deploy", description: "Deploy configuration changes", action: PageDeploy},
	}

	return &HomePage{
		cfg:   cfg,
		items: items,
	}
}

// Update handles input for the home page.
func (h *HomePage) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			h.cursor--
			if h.cursor < 0 {
				h.cursor = len(h.items) - 1
			}
			// Skip separators.
			if h.items[h.cursor].separator {
				h.cursor--
				if h.cursor < 0 {
					h.cursor = len(h.items) - 1
				}
			}
		case "down", "j":
			h.cursor++
			if h.cursor >= len(h.items) {
				h.cursor = 0
			}
			// Skip separators.
			if h.items[h.cursor].separator {
				h.cursor++
				if h.cursor >= len(h.items) {
					h.cursor = 0
				}
			}
		case "enter":
			if h.cursor >= 0 && h.cursor < len(h.items) && !h.items[h.cursor].separator {
				h.selected = h.items[h.cursor].action
			}
		}
	}
	return h, nil
}

// View renders the home page menu.
func (h *HomePage) View() string {
	var b strings.Builder
	b.WriteString("\n")
	for i, item := range h.items {
		if item.separator {
			b.WriteString("  ───\n")
			continue
		}
		cursor := "  "
		if i == h.cursor {
			cursor = "▸ "
		}
		b.WriteString(fmt.Sprintf("%s%-30s %s\n", cursor, item.label, item.description))
	}
	return b.String()
}
