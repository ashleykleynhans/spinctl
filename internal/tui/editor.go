package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

// editorItem represents one visible item in the editor.
type editorItem struct {
	key      string
	value    string
	node     *yaml.Node
	isScalar bool
}

// EditorPage provides drill-down navigation through a yaml.Node tree.
type EditorPage struct {
	root       *yaml.Node
	path       []string
	nodeStack  []*yaml.Node
	current    *yaml.Node
	cursor     int
	editing    bool
	editBuffer string
}

// NewEditorPage creates an editor page for the given YAML node.
func NewEditorPage(node *yaml.Node, breadcrumb string) *EditorPage {
	ep := &EditorPage{
		root:    node,
		current: node,
	}
	if breadcrumb != "" {
		ep.path = []string{breadcrumb}
	}
	return ep
}

// items returns the visible items for the current node.
func (e *EditorPage) items() []editorItem {
	if e.current == nil {
		return nil
	}
	switch e.current.Kind {
	case yaml.MappingNode:
		return e.mapItems()
	case yaml.SequenceNode:
		return e.sequenceItems()
	default:
		return nil
	}
}

// mapItems returns key/value pairs from a mapping node.
func (e *EditorPage) mapItems() []editorItem {
	var items []editorItem
	for i := 0; i+1 < len(e.current.Content); i += 2 {
		keyNode := e.current.Content[i]
		valNode := e.current.Content[i+1]
		item := editorItem{
			key:  keyNode.Value,
			node: valNode,
		}
		switch valNode.Kind {
		case yaml.ScalarNode:
			item.value = valNode.Value
			item.isScalar = true
		case yaml.MappingNode:
			item.value = fmt.Sprintf("{%d keys}", len(valNode.Content)/2)
		case yaml.SequenceNode:
			item.value = fmt.Sprintf("[%d items]", len(valNode.Content))
		}
		items = append(items, item)
	}
	return items
}

// sequenceItems returns indexed items from a sequence node.
func (e *EditorPage) sequenceItems() []editorItem {
	var items []editorItem
	for i, node := range e.current.Content {
		item := editorItem{
			key:  fmt.Sprintf("[%d]", i),
			node: node,
		}
		switch node.Kind {
		case yaml.ScalarNode:
			item.value = node.Value
			item.isScalar = true
		case yaml.MappingNode:
			item.value = fmt.Sprintf("{%d keys}", len(node.Content)/2)
		case yaml.SequenceNode:
			item.value = fmt.Sprintf("[%d items]", len(node.Content))
		}
		items = append(items, item)
	}
	return items
}

// Update handles input for the editor page.
func (e *EditorPage) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if e.editing {
			return e.updateEditing(msg)
		}
		return e.updateBrowsing(msg)
	}
	return e, nil
}

func (e *EditorPage) updateEditing(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Save the edit.
		items := e.items()
		if e.cursor >= 0 && e.cursor < len(items) && items[e.cursor].isScalar {
			items[e.cursor].node.Value = e.editBuffer
		}
		e.editing = false
	case tea.KeyEscape:
		e.editing = false
	case tea.KeyBackspace:
		if len(e.editBuffer) > 0 {
			e.editBuffer = e.editBuffer[:len(e.editBuffer)-1]
		}
	case tea.KeyRunes:
		e.editBuffer += string(msg.Runes)
	}
	return e, nil
}

func (e *EditorPage) updateBrowsing(msg tea.KeyMsg) (page, tea.Cmd) {
	items := e.items()
	switch msg.String() {
	case "up", "k":
		e.cursor--
		if e.cursor < 0 && len(items) > 0 {
			e.cursor = len(items) - 1
		}
	case "down", "j":
		e.cursor++
		if e.cursor >= len(items) {
			e.cursor = 0
		}
	case "enter":
		if e.cursor >= 0 && e.cursor < len(items) {
			item := items[e.cursor]
			if item.isScalar {
				e.editing = true
				e.editBuffer = item.value
			} else if item.node != nil && (item.node.Kind == yaml.MappingNode || item.node.Kind == yaml.SequenceNode) {
				e.nodeStack = append(e.nodeStack, e.current)
				e.path = append(e.path, item.key)
				e.current = item.node
				e.cursor = 0
			}
		}
	case "esc":
		if len(e.nodeStack) > 0 {
			e.current = e.nodeStack[len(e.nodeStack)-1]
			e.nodeStack = e.nodeStack[:len(e.nodeStack)-1]
			if len(e.path) > 0 {
				e.path = e.path[:len(e.path)-1]
			}
			e.cursor = 0
		}
	}
	return e, nil
}

// View renders the editor page.
func (e *EditorPage) View() string {
	var b strings.Builder

	// Breadcrumb.
	breadcrumb := "root"
	if len(e.path) > 0 {
		breadcrumb = strings.Join(e.path, " > ")
	}
	b.WriteString(fmt.Sprintf("\n  %s\n\n", breadcrumb))

	items := e.items()
	if len(items) == 0 {
		b.WriteString("  (empty)\n")
		return b.String()
	}

	for i, item := range items {
		cursor := "  "
		if i == e.cursor {
			cursor = "▸ "
		}
		if e.editing && i == e.cursor {
			b.WriteString(fmt.Sprintf("%s%s: %s█\n", cursor, item.key, e.editBuffer))
		} else {
			b.WriteString(fmt.Sprintf("%s%-20s %s\n", cursor, item.key, item.value))
		}
	}

	b.WriteString("\n  enter: edit/drill  esc: back\n")
	return b.String()
}
