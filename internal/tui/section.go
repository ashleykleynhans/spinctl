package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// SectionPage wraps an EditorPage for a generic config section,
// sending goBackMsg when esc is pressed at the editor root.
type SectionPage struct {
	editor *EditorPage
}

func newSectionPage(editor *EditorPage) *SectionPage {
	return &SectionPage{editor: editor}
}

func (s *SectionPage) Update(msg tea.Msg) (page, tea.Cmd) {
	if s.editor == nil {
		return s, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" && len(s.editor.nodeStack) == 0 {
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
