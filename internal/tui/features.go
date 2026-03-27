package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
)

// FeaturesPage displays feature flags.
type FeaturesPage struct {
	cfg         *config.SpinctlConfig
	sortedNames []string
	cursor      int
}

// NewFeaturesPage creates a features page.
func NewFeaturesPage(cfg *config.SpinctlConfig) *FeaturesPage {
	names := make([]string, 0, len(cfg.Features))
	for name := range cfg.Features {
		names = append(names, name)
	}
	sort.Strings(names)
	return &FeaturesPage{cfg: cfg, sortedNames: names}
}

func (f *FeaturesPage) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if f.cursor > 0 {
				f.cursor--
			}
		case "down", "j":
			if f.cursor < len(f.sortedNames)-1 {
				f.cursor++
			}
		case "enter":
			if f.cursor >= 0 && f.cursor < len(f.sortedNames) {
				name := f.sortedNames[f.cursor]
				f.cfg.Features[name] = !f.cfg.Features[name]
			}
		}
	}
	return f, nil
}

func (f *FeaturesPage) View() string {
	var b strings.Builder
	b.WriteString("\n  Features\n\n")

	if len(f.sortedNames) == 0 {
		b.WriteString("  No feature flags configured.\n")
		b.WriteString("\n  esc: back\n")
		return b.String()
	}

	for i, name := range f.sortedNames {
		cursor := "  "
		if i == f.cursor {
			cursor = "▸ "
		}
		status := "OFF"
		if f.cfg.Features[name] {
			status = " ON"
		}
		b.WriteString(fmt.Sprintf("%s[%s] %s\n", cursor, status, name))
	}

	b.WriteString("\n  enter: toggle  esc: back\n")
	return b.String()
}
