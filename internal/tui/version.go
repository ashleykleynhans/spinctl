package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
)

// VersionPage displays and allows editing the Spinnaker version.
type VersionPage struct {
	cfg     *config.SpinctlConfig
	editing bool
	buffer  string
}

// NewVersionPage creates a version page.
func NewVersionPage(cfg *config.SpinctlConfig) *VersionPage {
	return &VersionPage{
		cfg:    cfg,
		buffer: cfg.Version,
	}
}

func (v *VersionPage) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if v.editing {
			switch msg.Type {
			case tea.KeyEnter:
				v.cfg.Version = v.buffer
				v.editing = false
				return v, func() tea.Msg { return configChangedMsg{} }
			case tea.KeyBackspace:
				if len(v.buffer) > 0 {
					v.buffer = v.buffer[:len(v.buffer)-1]
				}
			case tea.KeyRunes:
				v.buffer += string(msg.Runes)
			}
		} else {
			switch msg.String() {
			case "enter":
				v.editing = true
			case "esc":
				return v, func() tea.Msg { return goBackMsg{} }
			}
		}
	}
	return v, nil
}

func (v *VersionPage) View() string {
	var b strings.Builder
	b.WriteString("\n  Spinnaker Version\n\n")

	if v.editing {
		b.WriteString(fmt.Sprintf("  Version: %s█\n", v.buffer))
		b.WriteString("\n  enter: save  esc: back\n")
	} else {
		b.WriteString(fmt.Sprintf("  Version: %s\n", v.cfg.Version))
		b.WriteString("\n  enter: edit  esc: back\n")
	}

	return b.String()
}
