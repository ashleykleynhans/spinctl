package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

func makeTestNode(yamlStr string) yaml.Node {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
		panic(err)
	}
	// Unwrap document node.
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return *doc.Content[0]
	}
	return doc
}

func TestEditorPageShowsMapKeys(t *testing.T) {
	node := makeTestNode("name: test\nport: 8080\nhost: localhost")
	ep := NewEditorPage(&node, "config")
	view := ep.View()
	if !strings.Contains(view, "name") {
		t.Error("editor should show 'name' key")
	}
	if !strings.Contains(view, "port") {
		t.Error("editor should show 'port' key")
	}
	if !strings.Contains(view, "host") {
		t.Error("editor should show 'host' key")
	}
}

func TestEditorPageDrillIntoMap(t *testing.T) {
	node := makeTestNode("server:\n  host: localhost\n  port: 8080")
	ep := NewEditorPage(&node, "config")

	// Enter to drill into 'server'.
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view := ep.View()
	if !strings.Contains(view, "host") {
		t.Error("after drill, should show nested keys")
	}
	if !strings.Contains(view, "localhost") {
		t.Error("after drill, should show nested values")
	}
}

func TestEditorPageBreadcrumb(t *testing.T) {
	node := makeTestNode("server:\n  host: localhost\n  port: 8080")
	ep := NewEditorPage(&node, "config")

	view := ep.View()
	if !strings.Contains(view, "config") {
		t.Error("breadcrumb should show 'config'")
	}

	// Drill into server.
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view = ep.View()
	if !strings.Contains(view, "server") {
		t.Error("breadcrumb should show 'server' after drilling in")
	}
}

func TestEditorPageGoBack(t *testing.T) {
	node := makeTestNode("server:\n  host: localhost")
	ep := NewEditorPage(&node, "config")

	// Drill in.
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Go back.
	ep.Update(tea.KeyMsg{Type: tea.KeyEscape})
	view := ep.View()
	if !strings.Contains(view, "server") {
		t.Error("after going back, should show parent keys")
	}
}

func TestEditorPageEditScalar(t *testing.T) {
	node := makeTestNode("name: original")
	ep := NewEditorPage(&node, "config")

	// Enter to start editing.
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ep.mode != modeEdit {
		t.Error("should be in editing mode")
	}

	// Clear and type new value.
	for range len(ep.editBuffer) {
		ep.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n', 'e', 'w'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if ep.mode == modeEdit {
		t.Error("should exit editing mode after enter")
	}

	// Check the node was updated.
	items := ep.items()
	if items[0].value != "new" {
		t.Errorf("value = %q, want %q", items[0].value, "new")
	}
}

func TestEditorPageShowsList(t *testing.T) {
	node := makeTestNode("items:\n  - alpha\n  - beta\n  - gamma")
	ep := NewEditorPage(&node, "config")

	// Drill into the sequence.
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view := ep.View()
	if !strings.Contains(view, "[0]") {
		t.Error("should show sequence indices")
	}
	if !strings.Contains(view, "alpha") {
		t.Error("should show sequence values")
	}
}

func TestEditorPageNoBreadcrumb(t *testing.T) {
	node := makeTestNode("key: value")
	ep := NewEditorPage(&node, "")
	view := ep.View()
	if !strings.Contains(view, "root") {
		t.Error("empty breadcrumb should show 'root'")
	}
}

func TestEditorPageEmptyNode(t *testing.T) {
	node := makeTestNode("{}")
	ep := NewEditorPage(&node, "config")
	view := ep.View()
	if !strings.Contains(view, "(empty)") {
		t.Error("empty map should show '(empty)'")
	}
}

func TestEditorPageNilNode(t *testing.T) {
	ep := NewEditorPage(nil, "config")
	items := ep.items()
	if items != nil {
		t.Error("nil node should return nil items")
	}
}

func TestEditorPageCursorNavigation(t *testing.T) {
	node := makeTestNode("a: 1\nb: 2\nc: 3")
	ep := NewEditorPage(&node, "config")

	// Move down.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if ep.cursor != 1 {
		t.Errorf("cursor = %d, want 1", ep.cursor)
	}

	// Move down again.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if ep.cursor != 2 {
		t.Errorf("cursor = %d, want 2", ep.cursor)
	}

	// Move down past end wraps to 0.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if ep.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (wrapped)", ep.cursor)
	}

	// Move up from 0 wraps to last.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if ep.cursor != 2 {
		t.Errorf("cursor = %d, want 2 (wrapped up)", ep.cursor)
	}
}

func TestEditorPageCursorUpKey(t *testing.T) {
	node := makeTestNode("a: 1\nb: 2")
	ep := NewEditorPage(&node, "config")
	ep.cursor = 1

	ep.Update(tea.KeyMsg{Type: tea.KeyUp})
	if ep.cursor != 0 {
		t.Errorf("cursor = %d, want 0", ep.cursor)
	}
}

func TestEditorPageCursorDownKey(t *testing.T) {
	node := makeTestNode("a: 1\nb: 2")
	ep := NewEditorPage(&node, "config")

	ep.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ep.cursor != 1 {
		t.Errorf("cursor = %d, want 1", ep.cursor)
	}
}

func TestEditorPageEditViewShowsCursor(t *testing.T) {
	node := makeTestNode("name: test")
	ep := NewEditorPage(&node, "config")

	// Enter editing mode.
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ep.mode != modeEdit {
		t.Error("should be editing")
	}
	view := ep.View()
	if !strings.Contains(view, "\u2588") {
		t.Error("editing view should show cursor block")
	}
}

func TestEditorPageEscapeFromEditMode(t *testing.T) {
	node := makeTestNode("name: original")
	ep := NewEditorPage(&node, "config")

	// Enter editing mode.
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ep.mode != modeEdit {
		t.Error("should be in editing mode")
	}

	// Type something.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Escape should cancel edit (exit editing mode without saving).
	ep.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if ep.mode == modeEdit {
		t.Error("should exit editing mode after escape")
	}

	// Original value should be preserved.
	items := ep.items()
	if items[0].value != "original" {
		t.Errorf("value = %q, want 'original' (escape should not save)", items[0].value)
	}
}

func TestEditorPageEscAtRootDoesNothing(t *testing.T) {
	node := makeTestNode("a: 1")
	ep := NewEditorPage(&node, "config")

	// Esc at root level with empty nodeStack should not panic.
	ep.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if ep.cursor != 0 {
		t.Errorf("cursor changed unexpectedly")
	}
}

func TestEditorPageSequenceWithNestedMap(t *testing.T) {
	node := makeTestNode("items:\n  - name: foo\n    value: bar")
	ep := NewEditorPage(&node, "config")

	// Drill into the sequence.
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view := ep.View()
	// Should show the name field value instead of [0].
	if !strings.Contains(view, "foo") {
		t.Error("should show name value for named sequence items")
	}
	// The item is a mapping with 2 keys.
	if !strings.Contains(view, "2 keys") {
		t.Error("should show key count for nested mapping")
	}
}

func TestEditorPageNonKeyMessage(t *testing.T) {
	node := makeTestNode("a: 1")
	ep := NewEditorPage(&node, "config")

	type customMsg struct{}
	result, cmd := ep.Update(customMsg{})
	if result != ep {
		t.Error("non-key message should return same page")
	}
	if cmd != nil {
		t.Error("non-key message should return nil cmd")
	}
}

func TestEditorPageNestedSequenceInSequence(t *testing.T) {
	node := makeTestNode("lists:\n  - - inner1\n    - inner2")
	ep := NewEditorPage(&node, "config")

	// Drill into top-level mapping (lists key).
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Now inside the sequence, first item is also a sequence.
	view := ep.View()
	if !strings.Contains(view, "[0]") {
		t.Error("should show sequence index")
	}
	if !strings.Contains(view, "2 items") {
		t.Error("should show item count for nested sequence")
	}
}
