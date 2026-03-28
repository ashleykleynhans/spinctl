package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

// SectionPage wraps an EditorPage for a generic config section,
// sending goBackMsg when esc is pressed at the editor root.
// It includes a save callback to write changes back to the config.
type SectionPage struct {
	editor *EditorPage
	onSave func(*yaml.Node) // called when navigating back to persist changes
}

func newSectionPage(editor *EditorPage, onSave func(*yaml.Node)) *SectionPage {
	return &SectionPage{editor: editor, onSave: onSave}
}

func (s *SectionPage) Update(msg tea.Msg) (page, tea.Cmd) {
	if s.editor == nil {
		return s, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" && len(s.editor.nodeStack) == 0 {
			// Persist changes before going back.
			if s.onSave != nil {
				s.onSave(s.editor.root)
			}
			return s, func() tea.Msg { return goBackMsg{} }
		}
	}
	var cmd tea.Cmd
	_, cmd = s.editor.Update(msg)
	return s, cmd
}

func (s *SectionPage) View() string {
	if s.editor != nil {
		return s.editor.View()
	}
	return "\n  (empty)\n"
}
