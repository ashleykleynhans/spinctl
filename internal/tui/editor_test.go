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

func TestEditorAddKeyValueToMap(t *testing.T) {
	node := makeTestNode("a: 1")
	ep := NewEditorPage(&node, "config")

	// Press + to start adding.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	if ep.mode != modeAddKey {
		t.Fatalf("mode = %v, want modeAddKey", ep.mode)
	}

	// Type key name.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ep.mode != modeAddValue {
		t.Fatalf("mode = %v, want modeAddValue", ep.mode)
	}

	// Type value.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	result, cmd := ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if result.(*EditorPage).mode != modeBrowse {
		t.Error("should return to browse mode")
	}
	if cmd == nil {
		t.Error("should return configChangedMsg cmd")
	}

	items := ep.items()
	found := false
	for _, item := range items {
		if item.key == "b" && item.value == "2" {
			found = true
		}
	}
	if !found {
		t.Error("new key/value pair 'b: 2' not found")
	}
}

func TestEditorAddKeyCancel(t *testing.T) {
	node := makeTestNode("a: 1")
	ep := NewEditorPage(&node, "config")

	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if ep.mode != modeBrowse {
		t.Error("should return to browse mode after esc")
	}
	items := ep.items()
	if len(items) != 1 {
		t.Errorf("items count = %d, want 1 (nothing added)", len(items))
	}
}

func TestEditorAddValueCancel(t *testing.T) {
	node := makeTestNode("a: 1")
	ep := NewEditorPage(&node, "config")

	// Start add, enter key, then cancel during value.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ep.mode != modeAddValue {
		t.Fatalf("mode = %v, want modeAddValue", ep.mode)
	}

	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if ep.mode != modeBrowse {
		t.Error("should return to browse mode after esc")
	}
	items := ep.items()
	if len(items) != 1 {
		t.Errorf("items count = %d, want 1 (nothing added)", len(items))
	}
}

func TestEditorAddListItem(t *testing.T) {
	node := makeTestNode("- alpha\n- beta")
	ep := NewEditorPage(&node, "config")

	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	if ep.mode != modeAddListItem {
		t.Fatalf("mode = %v, want modeAddListItem", ep.mode)
	}

	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g', 'a', 'm', 'm', 'a'}})
	_, cmd := ep.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if ep.mode != modeBrowse {
		t.Error("should return to browse mode")
	}
	if cmd == nil {
		t.Error("should return configChangedMsg cmd")
	}

	items := ep.items()
	if len(items) != 3 {
		t.Errorf("items count = %d, want 3", len(items))
	}
	if items[2].value != "gamma" {
		t.Errorf("new item value = %q, want 'gamma'", items[2].value)
	}
}

func TestEditorAddListItemCancel(t *testing.T) {
	node := makeTestNode("- alpha\n- beta")
	ep := NewEditorPage(&node, "config")

	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if ep.mode != modeBrowse {
		t.Error("should return to browse mode")
	}
	items := ep.items()
	if len(items) != 2 {
		t.Errorf("items count = %d, want 2 (nothing added)", len(items))
	}
}

func TestEditorDeleteItem(t *testing.T) {
	node := makeTestNode("a: 1\nb: 2")
	ep := NewEditorPage(&node, "config")

	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if ep.mode != modeConfirmDelete {
		t.Fatalf("mode = %v, want modeConfirmDelete", ep.mode)
	}

	_, cmd := ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if ep.mode != modeBrowse {
		t.Error("should return to browse mode")
	}
	if cmd == nil {
		t.Error("should return configChangedMsg cmd")
	}

	items := ep.items()
	if len(items) != 1 {
		t.Errorf("items count = %d, want 1", len(items))
	}
	if items[0].key != "b" {
		t.Errorf("remaining key = %q, want 'b'", items[0].key)
	}
}

func TestEditorDeleteCancel(t *testing.T) {
	node := makeTestNode("a: 1\nb: 2")
	ep := NewEditorPage(&node, "config")

	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if ep.mode != modeBrowse {
		t.Error("should return to browse mode")
	}
	items := ep.items()
	if len(items) != 2 {
		t.Errorf("items count = %d, want 2 (nothing deleted)", len(items))
	}
}

func TestEditorDeleteFromSequence(t *testing.T) {
	node := makeTestNode("- alpha\n- beta\n- gamma")
	ep := NewEditorPage(&node, "config")

	// Delete the first item.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	items := ep.items()
	if len(items) != 2 {
		t.Errorf("items count = %d, want 2", len(items))
	}
	if items[0].value != "beta" {
		t.Errorf("first item value = %q, want 'beta'", items[0].value)
	}
}

func TestEditorBoolToggle(t *testing.T) {
	node := makeTestNode("enabled: true")
	ep := NewEditorPage(&node, "config")

	_, cmd := ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("should return configChangedMsg cmd")
	}

	items := ep.items()
	if items[0].value != "false" {
		t.Errorf("value = %q, want 'false' after toggle", items[0].value)
	}

	// Toggle back.
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	items = ep.items()
	if items[0].value != "true" {
		t.Errorf("value = %q, want 'true' after second toggle", items[0].value)
	}
}

func TestEditorViewAddKeyMode(t *testing.T) {
	node := makeTestNode("a: 1")
	ep := NewEditorPage(&node, "config")
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})

	view := ep.View()
	if !strings.Contains(view, "New key:") {
		t.Error("add key mode should show 'New key:' prompt")
	}
	if !strings.Contains(view, "esc: cancel") {
		t.Error("add key mode should show cancel hint")
	}
}

func TestEditorViewAddValueMode(t *testing.T) {
	node := makeTestNode("a: 1")
	ep := NewEditorPage(&node, "config")
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})

	view := ep.View()
	if !strings.Contains(view, "Key: k") {
		t.Error("add value mode should show the key name")
	}
	if !strings.Contains(view, "Value:") {
		t.Error("add value mode should show 'Value:' prompt")
	}
}

func TestEditorViewAddListItemMode(t *testing.T) {
	node := makeTestNode("- alpha")
	ep := NewEditorPage(&node, "config")
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})

	view := ep.View()
	if !strings.Contains(view, "New item:") {
		t.Error("add list item mode should show 'New item:' prompt")
	}
}

func TestEditorViewConfirmDeleteMode(t *testing.T) {
	node := makeTestNode("a: 1\nb: 2")
	ep := NewEditorPage(&node, "config")
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	view := ep.View()
	if !strings.Contains(view, "Delete") {
		t.Error("confirm delete mode should show 'Delete' prompt")
	}
	if !strings.Contains(view, "y/n") {
		t.Error("confirm delete mode should show y/n options")
	}
}

func TestEditorViewBoolDisplay(t *testing.T) {
	node := makeTestNode("enabled: true\ndisabled: false")
	ep := NewEditorPage(&node, "config")
	view := ep.View()

	if !strings.Contains(view, "[ ON]") {
		t.Error("true boolean should show [ ON]")
	}
	if !strings.Contains(view, "[OFF]") {
		t.Error("false boolean should show [OFF]")
	}
}

func TestEditorAddKeyBackspace(t *testing.T) {
	node := makeTestNode("a: 1")
	ep := NewEditorPage(&node, "config")

	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a', 'b'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if ep.editBuffer != "a" {
		t.Errorf("editBuffer = %q, want 'a'", ep.editBuffer)
	}
}

func TestEditorAddValueBackspace(t *testing.T) {
	node := makeTestNode("a: 1")
	ep := NewEditorPage(&node, "config")

	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v', 'x'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if ep.editBuffer != "v" {
		t.Errorf("editBuffer = %q, want 'v'", ep.editBuffer)
	}
}

func TestEditorAddListItemBackspace(t *testing.T) {
	node := makeTestNode("- alpha")
	ep := NewEditorPage(&node, "config")

	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x', 'y'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if ep.editBuffer != "x" {
		t.Errorf("editBuffer = %q, want 'x'", ep.editBuffer)
	}
}

func TestEditorAddKeyEmptyEnterDoesNothing(t *testing.T) {
	node := makeTestNode("a: 1")
	ep := NewEditorPage(&node, "config")

	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	// Enter with empty key should stay in modeAddKey.
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ep.mode != modeAddKey {
		t.Errorf("mode = %v, want modeAddKey (empty key should not advance)", ep.mode)
	}
}

func TestEditorDeleteOnEmptyList(t *testing.T) {
	node := makeTestNode("{}")
	ep := NewEditorPage(&node, "config")

	// d on empty list should not enter confirm delete.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if ep.mode == modeConfirmDelete {
		t.Error("should not enter confirmDelete on empty map")
	}
}

func TestEditorConfirmDeleteEsc(t *testing.T) {
	node := makeTestNode("a: 1\nb: 2")
	ep := NewEditorPage(&node, "config")

	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if ep.mode != modeBrowse {
		t.Error("esc should return to browse mode")
	}
	items := ep.items()
	if len(items) != 2 {
		t.Errorf("items count = %d, want 2", len(items))
	}
}

func TestEditorEmptyWithAddHint(t *testing.T) {
	node := makeTestNode("{}")
	ep := NewEditorPage(&node, "config")
	view := ep.View()

	if !strings.Contains(view, "+: add") {
		t.Error("empty map should show '+: add' hint")
	}
}

func TestFindItemLabelFallback(t *testing.T) {
	// Create a mapping node with no name/id/title/key/region/label fields,
	// but with a scalar value that should be used as fallback.
	node := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "customField"},
			{Kind: yaml.ScalarNode, Value: "myValue"},
		},
	}
	label := findItemLabel(node)
	if label != "myValue" {
		t.Errorf("findItemLabel fallback = %q, want %q", label, "myValue")
	}
}

func TestFindItemLabelNonMapping(t *testing.T) {
	node := &yaml.Node{Kind: yaml.ScalarNode, Value: "hello"}
	label := findItemLabel(node)
	if label != "" {
		t.Errorf("findItemLabel(scalar) = %q, want empty", label)
	}
}

func TestDeleteLastItem(t *testing.T) {
	node := makeTestNode("a: 1\nb: 2\nc: 3")
	ep := NewEditorPage(&node, "config")

	// Move cursor to last item (index 2).
	ep.cursor = 2

	// Delete.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	items := ep.items()
	if len(items) != 2 {
		t.Fatalf("items count = %d, want 2", len(items))
	}
	// Cursor should have adjusted down to the new last item.
	if ep.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after deleting last item", ep.cursor)
	}
}

func TestDeleteFromEmptyDoesNothing(t *testing.T) {
	node := makeTestNode("{}")
	ep := NewEditorPage(&node, "config")

	// Directly call deleteCurrentItem on empty node.
	ep.deleteCurrentItem()
	items := ep.items()
	if len(items) != 0 {
		t.Errorf("items count = %d, want 0", len(items))
	}
}

func TestHasTypeSelector(t *testing.T) {
	node := makeTestNode("persistentStoreType: s3\ns3:\n  bucket: my-bucket")
	ep := NewEditorPage(&node, "config")
	if !ep.hasTypeSelector() {
		t.Error("expected hasTypeSelector to return true for node with *Type key")
	}
}

func TestHasTypeSelectorFalse(t *testing.T) {
	node := makeTestNode("name: test\nport: 8080")
	ep := NewEditorPage(&node, "config")
	if ep.hasTypeSelector() {
		t.Error("expected hasTypeSelector to return false for node without *Type key")
	}
}

func TestIsActiveByTypeSelectorNoMatch(t *testing.T) {
	node := makeTestNode("persistentStoreType: s3\ngcs:\n  bucket: my-bucket")
	ep := NewEditorPage(&node, "config")
	if ep.isActiveByTypeSelector("gcs") {
		t.Error("gcs should not be active when persistentStoreType is s3")
	}
}

func TestSequenceItemsWithTypeSelector(t *testing.T) {
	node := makeTestNode("- name: alpha\n  enabled: true\n- name: beta\n  enabled: false")
	ep := NewEditorPage(&node, "config")
	items := ep.sequenceItems()
	if len(items) != 2 {
		t.Fatalf("items count = %d, want 2", len(items))
	}
	if !strings.Contains(items[0].value, " ON") {
		t.Errorf("first item value = %q, should contain ON", items[0].value)
	}
	if !strings.Contains(items[1].value, "OFF") {
		t.Errorf("second item value = %q, should contain OFF", items[1].value)
	}
}

func TestMapItemsEnabledSortsToTop(t *testing.T) {
	node := makeTestNode("name: test\nenabled: true\nport: 8080")
	ep := NewEditorPage(&node, "config")
	items := ep.items()
	if len(items) < 1 {
		t.Fatal("expected items")
	}
	if items[0].key != "enabled" {
		t.Errorf("first item key = %q, want 'enabled' (should sort to top)", items[0].key)
	}
}

func TestMapItemsTypeSelectorInactive(t *testing.T) {
	node := makeTestNode("persistentStoreType: s3\ns3:\n  bucket: my-bucket\ngcs:\n  bucket: other")
	ep := NewEditorPage(&node, "config")
	items := ep.items()
	// Find the gcs item.
	for _, item := range items {
		if item.key == "gcs" {
			if !strings.Contains(item.value, "OFF") {
				t.Errorf("inactive type selector sibling should show [OFF], got %q", item.value)
			}
			return
		}
	}
	t.Error("gcs item not found")
}

func TestSpacebarOnNonBool(t *testing.T) {
	node := makeTestNode("name: hello")
	ep := NewEditorPage(&node, "config")

	// Press spacebar on a non-boolean scalar; should do nothing.
	_, cmd := ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if cmd != nil {
		t.Error("spacebar on non-boolean should return nil cmd")
	}
	items := ep.items()
	if items[0].value != "hello" {
		t.Errorf("value should remain 'hello', got %q", items[0].value)
	}
}

func TestToggleMapEnabled(t *testing.T) {
	node := makeTestNode("enabled: true\nname: test")
	toggled := toggleMapEnabled(&node)
	if !toggled {
		t.Error("should toggle enabled")
	}
	if findMapValue(&node, "enabled") != "false" {
		t.Error("enabled should be false after toggle")
	}
	toggleMapEnabled(&node)
	if findMapValue(&node, "enabled") != "true" {
		t.Error("enabled should be true after second toggle")
	}
}

func TestToggleMapEnabledNoKey(t *testing.T) {
	node := makeTestNode("name: test")
	if toggleMapEnabled(&node) {
		t.Error("should not toggle when no enabled key")
	}
}

func TestToggleMapEnabledNonMapping(t *testing.T) {
	node := makeTestNode("- a\n- b")
	if toggleMapEnabled(&node) {
		t.Error("should not toggle on sequence node")
	}
}

func TestEditorItemsScalarNode(t *testing.T) {
	// A scalar node at the top level should return nil items.
	node := &yaml.Node{Kind: yaml.ScalarNode, Value: "hello"}
	ep := NewEditorPage(node, "config")
	items := ep.items()
	if items != nil {
		t.Errorf("scalar node should return nil items, got %d", len(items))
	}
}

func TestIsBoolNodeWithTag(t *testing.T) {
	node := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"}
	if !isBoolNode(node) {
		t.Error("node with !!bool tag should be recognized as bool")
	}
	node2 := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "yes"}
	if !isBoolNode(node2) {
		t.Error("node with !!bool tag should be recognized as bool regardless of value")
	}
}

func TestIsBoolNodeNilNode(t *testing.T) {
	if isBoolNode(nil) {
		t.Error("nil node should not be a bool")
	}
}

func TestIsBoolNodeNonScalar(t *testing.T) {
	node := &yaml.Node{Kind: yaml.MappingNode}
	if isBoolNode(node) {
		t.Error("mapping node should not be a bool")
	}
}

func TestFindMapValueNonMapping(t *testing.T) {
	node := &yaml.Node{Kind: yaml.SequenceNode}
	val := findMapValue(node, "key")
	if val != "" {
		t.Errorf("findMapValue on non-mapping should return empty, got %q", val)
	}
}

func TestHasTypeSelectorNilCurrent(t *testing.T) {
	ep := NewEditorPage(nil, "config")
	if ep.hasTypeSelector() {
		t.Error("nil current should return false for hasTypeSelector")
	}
}

func TestIsActiveByTypeSelectorNilCurrent(t *testing.T) {
	ep := NewEditorPage(nil, "config")
	if ep.isActiveByTypeSelector("s3") {
		t.Error("nil current should return false for isActiveByTypeSelector")
	}
}

func TestIsActiveByTypeSelectorNonMapping(t *testing.T) {
	node := &yaml.Node{Kind: yaml.SequenceNode}
	ep := NewEditorPage(node, "config")
	if ep.isActiveByTypeSelector("s3") {
		t.Error("sequence node should return false for isActiveByTypeSelector")
	}
}

func TestEditorUpdateEditingEscapeDuringEdit(t *testing.T) {
	node := makeTestNode("name: original")
	ep := NewEditorPage(&node, "config")

	// Enter editing mode.
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ep.mode != modeEdit {
		t.Fatal("should be in edit mode")
	}

	// Type something then escape.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x', 'y'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if ep.mode != modeBrowse {
		t.Error("escape should return to browse mode")
	}
	// Value should NOT have been saved.
	items := ep.items()
	if items[0].value != "original" {
		t.Errorf("value = %q, want 'original'", items[0].value)
	}
}

func TestEditorUpdateConfirmDeleteEsc(t *testing.T) {
	node := makeTestNode("a: 1\nb: 2")
	ep := NewEditorPage(&node, "config")

	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if ep.mode != modeConfirmDelete {
		t.Fatal("should be in confirm delete mode")
	}
	ep.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if ep.mode != modeBrowse {
		t.Error("esc should cancel delete")
	}
	items := ep.items()
	if len(items) != 2 {
		t.Errorf("items count = %d, want 2", len(items))
	}
}

func TestEditorBrowsingPlusOnNilCurrent(t *testing.T) {
	ep := NewEditorPage(nil, "config")
	// Press + on nil current should not panic.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	if ep.mode != modeBrowse {
		t.Error("should stay in browse mode")
	}
}

func TestEditorBrowsingDOnEmptyList(t *testing.T) {
	node := makeTestNode("{}")
	ep := NewEditorPage(&node, "config")
	// d on empty map should not enter confirm delete.
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if ep.mode == modeConfirmDelete {
		t.Error("should not enter confirm delete on empty map")
	}
}

func TestEditorMapItemsEmptyMappingWithTypeSelector(t *testing.T) {
	// A map with a type selector where a sibling has 0 keys should show [OFF].
	node := makeTestNode("persistentStoreType: s3\ngcs: {}")
	ep := NewEditorPage(&node, "config")
	items := ep.items()
	for _, item := range items {
		if item.key == "gcs" {
			if item.value != "[OFF]" {
				t.Errorf("empty inactive sibling should show [OFF], got %q", item.value)
			}
			return
		}
	}
	t.Error("gcs item not found")
}

func TestEditorMapItemsEmptyMappingNoTypeSelector(t *testing.T) {
	node := makeTestNode("inner: {}")
	ep := NewEditorPage(&node, "config")
	items := ep.items()
	if len(items) != 1 {
		t.Fatalf("items count = %d, want 1", len(items))
	}
	if items[0].value != "(empty)" {
		t.Errorf("empty mapping value = %q, want '(empty)'", items[0].value)
	}
}

func TestEditorSpaceTogglesBool(t *testing.T) {
	node := makeTestNode("enabled: true")
	ep := NewEditorPage(&node, "config")

	_, cmd := ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if cmd == nil {
		t.Error("space on bool should return configChangedMsg cmd")
	}
	items := ep.items()
	if items[0].value != "false" {
		t.Errorf("value = %q, want 'false' after space toggle", items[0].value)
	}
}

func TestEditorUpdateEditingEnterOnNonScalar(t *testing.T) {
	// When in edit mode but cursor points to a non-scalar item, enter should
	// just return to browse mode.
	node := makeTestNode("inner:\n  key: val")
	ep := NewEditorPage(&node, "config")
	// Manually set to edit mode on the mapping entry.
	ep.mode = modeEdit
	ep.editBuffer = "something"
	ep.cursor = 0

	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ep.mode != modeBrowse {
		t.Error("enter on non-scalar in edit mode should return to browse")
	}
}

func TestEditorBrowsingEnterOnBoolDoesNotEditText(t *testing.T) {
	// Enter on a boolean scalar should toggle, not enter text edit mode.
	node := makeTestNode("enabled: false")
	ep := NewEditorPage(&node, "config")
	_, cmd := ep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("enter on bool should return configChangedMsg")
	}
	if ep.mode == modeEdit {
		t.Error("should not enter edit mode for bool toggle")
	}
	items := ep.items()
	if items[0].value != "true" {
		t.Errorf("value = %q, want 'true'", items[0].value)
	}
}

func TestEditorSequenceItemScalarInSequence(t *testing.T) {
	// Test that sequenceItems handles scalar items.
	node := makeTestNode("- hello\n- world")
	ep := NewEditorPage(&node, "config")
	items := ep.sequenceItems()
	if len(items) != 2 {
		t.Fatalf("items = %d, want 2", len(items))
	}
	if !items[0].isScalar {
		t.Error("scalar items in sequence should be marked as scalar")
	}
}

func TestEditorMapItemsKeyCountDisplay(t *testing.T) {
	// A map value that is a mapping without 'enabled' key should show {N keys}.
	node := makeTestNode("server:\n  host: localhost\n  port: 8080")
	ep := NewEditorPage(&node, "config")
	items := ep.items()
	if len(items) != 1 {
		t.Fatalf("items = %d, want 1", len(items))
	}
	if !strings.Contains(items[0].value, "2 keys") {
		t.Errorf("value = %q, should contain '2 keys'", items[0].value)
	}
}

func TestEditorMapItemsSequenceChild(t *testing.T) {
	// A map value that is a sequence should show [N items].
	node := makeTestNode("items:\n  - a\n  - b")
	ep := NewEditorPage(&node, "config")
	items := ep.items()
	if len(items) != 1 {
		t.Fatalf("items = %d, want 1", len(items))
	}
	if !strings.Contains(items[0].value, "2 items") {
		t.Errorf("value = %q, should contain '2 items'", items[0].value)
	}
}

func TestSpacebarTogglesMapEnabled(t *testing.T) {
	node := makeTestNode("slack:\n  enabled: true\n  botName: spinbot")
	ep := NewEditorPage(&node, "notifications")

	// Press space on slack (map with enabled key).
	_, cmd := ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if cmd == nil {
		t.Error("should return configChangedMsg")
	}
}
