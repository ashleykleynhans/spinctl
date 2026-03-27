package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var statusStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("236")).
	Foreground(lipgloss.Color("252")).
	Padding(0, 1)

var modifiedStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("236")).
	Foreground(lipgloss.Color("214")).
	Bold(true).
	Padding(0, 1)

// StatusBar renders a bottom status bar with key hints and a modified indicator.
type StatusBar struct {
	width    int
	modified bool
}

// NewStatusBar creates a new StatusBar with the given width.
func NewStatusBar(width int) *StatusBar {
	return &StatusBar{width: width}
}

// SetModified sets whether the config has been modified.
func (s *StatusBar) SetModified(modified bool) {
	s.modified = modified
}

// View renders the status bar with hints and an optional modified indicator.
func (s *StatusBar) View(hints string) string {
	right := ""
	if s.modified {
		right = modifiedStyle.Render("[modified]")
	}

	hintsRendered := statusStyle.Render(hints)

	// Calculate padding
	gap := s.width - lipgloss.Width(hintsRendered) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return fmt.Sprintf("%s%s%s", hintsRendered, strings.Repeat(" ", gap), right)
}
