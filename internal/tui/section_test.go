package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

func TestSectionPageView(t *testing.T) {
	node := makeTestNode("name: test\nport: 8080")
	editor := NewEditorPage(&node, "TestSection")
	sp := newSectionPage(editor)

	view := sp.View()
	if !strings.Contains(view, "name") {
		t.Error("view should contain 'name'")
	}
	if !strings.Contains(view, "port") {
		t.Error("view should contain 'port'")
	}
	if !strings.Contains(view, "TestSection") {
		t.Error("view should contain breadcrumb 'TestSection'")
	}
}

func TestSectionPageViewNilEditor(t *testing.T) {
	sp := newSectionPage(nil)
	view := sp.View()
	if !strings.Contains(view, "(empty)") {
		t.Errorf("nil editor view = %q, want '(empty)'", view)
	}
}

func TestSectionPageEscAtRootSendsGoBack(t *testing.T) {
	node := makeTestNode("a: 1")
	editor := NewEditorPage(&node, "Test")
	sp := newSectionPage(editor)

	// nodeStack is empty, so esc should send goBackMsg.
	_, cmd := sp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected a command for goBackMsg")
	}
	msg := cmd()
	if _, ok := msg.(goBackMsg); !ok {
		t.Errorf("expected goBackMsg, got %T", msg)
	}
}

func TestSectionPageEscDrillsUp(t *testing.T) {
	node := makeTestNode("server:\n  host: localhost\n  port: 8080")
	editor := NewEditorPage(&node, "Test")
	sp := newSectionPage(editor)

	// Drill into 'server'.
	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if len(editor.nodeStack) == 0 {
		t.Fatal("expected nodeStack to be non-empty after drill-in")
	}

	// Esc should drill up within the editor, not send goBackMsg.
	_, cmd := sp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd != nil {
		// The editor's esc in browse mode with nodeStack returns nil cmd.
		// But if a cmd is returned, it should NOT be goBackMsg.
		msg := cmd()
		if _, ok := msg.(goBackMsg); ok {
			t.Error("esc with non-empty nodeStack should not send goBackMsg")
		}
	}

	if len(editor.nodeStack) != 0 {
		t.Errorf("nodeStack len = %d, want 0 after drilling up", len(editor.nodeStack))
	}
}

func TestSectionPageUpdate(t *testing.T) {
	node := makeTestNode("a: 1\nb: 2\nc: 3")
	editor := NewEditorPage(&node, "Test")
	sp := newSectionPage(editor)

	// Forward a down key to the editor.
	sp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if editor.cursor != 1 {
		t.Errorf("editor cursor = %d, want 1", editor.cursor)
	}
}

func TestSectionPageUpdateNilEditor(t *testing.T) {
	sp := newSectionPage(nil)
	result, cmd := sp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if result != sp {
		t.Error("nil editor update should return same page")
	}
	if cmd != nil {
		t.Error("nil editor update should return nil cmd")
	}
}

func TestSectionPageUpdateNonKeyMsg(t *testing.T) {
	node := makeTestNode("a: 1")
	editor := NewEditorPage(&node, "Test")
	sp := newSectionPage(editor)

	type customMsg struct{}
	result, cmd := sp.Update(customMsg{})
	if result != sp {
		t.Error("non-key message should return same page")
	}
	if cmd != nil {
		t.Error("non-key message should return nil cmd")
	}
}

func TestNewSectionPageHelper(t *testing.T) {
	node := &yaml.Node{Kind: yaml.MappingNode}
	editor := NewEditorPage(node, "Test")
	sp := newSectionPage(editor)
	if sp.editor != editor {
		t.Error("newSectionPage should set the editor field")
	}
}
