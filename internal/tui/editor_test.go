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
	if !ep.editing {
		t.Error("should be in editing mode")
	}

	// Clear and type new value.
	for range len(ep.editBuffer) {
		ep.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	ep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n', 'e', 'w'}})
	ep.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if ep.editing {
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
