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

// editorMode tracks what input mode the editor is in.
type editorMode int

const (
	modeBrowse editorMode = iota
	modeEdit              // editing a scalar value
	modeAddKey            // entering a key name for a new map entry
	modeAddValue          // entering a value for the new map entry
	modeAddListItem       // entering a value for a new list item
	modeConfirmDelete     // confirming deletion
)

// EditorPage provides drill-down navigation through a yaml.Node tree.
type EditorPage struct {
	root       *yaml.Node
	path       []string
	nodeStack  []*yaml.Node
	current    *yaml.Node
	cursor     int
	mode       editorMode
	editBuffer string
	addKeyBuf  string // stashed key name during modeAddValue
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
// If items are mappings with a "name" key, use the name as the label.
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
			// Try to find a "name" key to use as label.
			if name := findMapValue(node, "name"); name != "" {
				item.key = name
			}
			item.value = fmt.Sprintf("{%d keys}", len(node.Content)/2)
		case yaml.SequenceNode:
			item.value = fmt.Sprintf("[%d items]", len(node.Content))
		}
		items = append(items, item)
	}
	return items
}

// isBoolNode checks if a yaml.Node is a boolean scalar.
func isBoolNode(node *yaml.Node) bool {
	if node == nil || node.Kind != yaml.ScalarNode {
		return false
	}
	if node.Tag == "!!bool" {
		return true
	}
	v := strings.ToLower(node.Value)
	return v == "true" || v == "false"
}

// findMapValue returns the scalar value for a given key in a mapping node.
func findMapValue(node *yaml.Node, key string) string {
	if node.Kind != yaml.MappingNode {
		return ""
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1].Value
		}
	}
	return ""
}

// Update handles input for the editor page.
func (e *EditorPage) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch e.mode {
		case modeEdit:
			return e.updateEditing(msg)
		case modeAddKey:
			return e.updateAddKey(msg)
		case modeAddValue:
			return e.updateAddValue(msg)
		case modeAddListItem:
			return e.updateAddListItem(msg)
		case modeConfirmDelete:
			return e.updateConfirmDelete(msg)
		default:
			return e.updateBrowsing(msg)
		}
	}
	return e, nil
}

func (e *EditorPage) updateEditing(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		items := e.items()
		if e.cursor >= 0 && e.cursor < len(items) && items[e.cursor].isScalar {
			items[e.cursor].node.Value = e.editBuffer
			e.mode = modeBrowse
			return e, func() tea.Msg { return configChangedMsg{} }
		}
		e.mode = modeBrowse
	case tea.KeyEscape:
		e.mode = modeBrowse
	case tea.KeyBackspace:
		if len(e.editBuffer) > 0 {
			e.editBuffer = e.editBuffer[:len(e.editBuffer)-1]
		}
	case tea.KeyRunes:
		e.editBuffer += string(msg.Runes)
	}
	return e, nil
}

func (e *EditorPage) updateAddKey(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if e.editBuffer != "" {
			e.addKeyBuf = e.editBuffer
			e.editBuffer = ""
			e.mode = modeAddValue
		}
	case tea.KeyEscape:
		e.mode = modeBrowse
		e.editBuffer = ""
	case tea.KeyBackspace:
		if len(e.editBuffer) > 0 {
			e.editBuffer = e.editBuffer[:len(e.editBuffer)-1]
		}
	case tea.KeyRunes:
		e.editBuffer += string(msg.Runes)
	}
	return e, nil
}

func (e *EditorPage) updateAddValue(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Add key/value pair to the current mapping node.
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: e.addKeyBuf, Tag: "!!str"}
		valNode := &yaml.Node{Kind: yaml.ScalarNode, Value: e.editBuffer, Tag: "!!str"}
		e.current.Content = append(e.current.Content, keyNode, valNode)
		e.editBuffer = ""
		e.addKeyBuf = ""
		e.mode = modeBrowse
		e.cursor = len(e.current.Content)/2 - 1
		return e, func() tea.Msg { return configChangedMsg{} }
	case tea.KeyEscape:
		e.mode = modeBrowse
		e.editBuffer = ""
		e.addKeyBuf = ""
	case tea.KeyBackspace:
		if len(e.editBuffer) > 0 {
			e.editBuffer = e.editBuffer[:len(e.editBuffer)-1]
		}
	case tea.KeyRunes:
		e.editBuffer += string(msg.Runes)
	}
	return e, nil
}

func (e *EditorPage) updateAddListItem(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Add scalar item to the current sequence node.
		valNode := &yaml.Node{Kind: yaml.ScalarNode, Value: e.editBuffer, Tag: "!!str"}
		e.current.Content = append(e.current.Content, valNode)
		e.editBuffer = ""
		e.mode = modeBrowse
		e.cursor = len(e.current.Content) - 1
		return e, func() tea.Msg { return configChangedMsg{} }
	case tea.KeyEscape:
		e.mode = modeBrowse
		e.editBuffer = ""
	case tea.KeyBackspace:
		if len(e.editBuffer) > 0 {
			e.editBuffer = e.editBuffer[:len(e.editBuffer)-1]
		}
	case tea.KeyRunes:
		e.editBuffer += string(msg.Runes)
	}
	return e, nil
}

func (e *EditorPage) updateConfirmDelete(msg tea.KeyMsg) (page, tea.Cmd) {
	switch msg.String() {
	case "y":
		e.deleteCurrentItem()
		e.mode = modeBrowse
		return e, func() tea.Msg { return configChangedMsg{} }
	case "n", "esc":
		e.mode = modeBrowse
	}
	return e, nil
}

func (e *EditorPage) deleteCurrentItem() {
	items := e.items()
	if e.cursor < 0 || e.cursor >= len(items) {
		return
	}

	switch e.current.Kind {
	case yaml.MappingNode:
		// Remove key + value (2 nodes per entry).
		idx := e.cursor * 2
		if idx+1 < len(e.current.Content) {
			e.current.Content = append(e.current.Content[:idx], e.current.Content[idx+2:]...)
		}
	case yaml.SequenceNode:
		e.current.Content = append(e.current.Content[:e.cursor], e.current.Content[e.cursor+1:]...)
	}

	// Adjust cursor.
	if e.cursor >= len(e.items()) && e.cursor > 0 {
		e.cursor--
	}
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
			if item.isScalar && isBoolNode(item.node) {
				// Toggle boolean values directly.
				if item.node.Value == "true" {
					item.node.Value = "false"
				} else {
					item.node.Value = "true"
				}
				return e, func() tea.Msg { return configChangedMsg{} }
			} else if item.isScalar {
				e.mode = modeEdit
				e.editBuffer = item.value
			} else if item.node != nil && (item.node.Kind == yaml.MappingNode || item.node.Kind == yaml.SequenceNode) {
				e.nodeStack = append(e.nodeStack, e.current)
				e.path = append(e.path, item.key)
				e.current = item.node
				e.cursor = 0
			}
		}
	case "+":
		if e.current != nil {
			switch e.current.Kind {
			case yaml.MappingNode:
				e.mode = modeAddKey
				e.editBuffer = ""
			case yaml.SequenceNode:
				e.mode = modeAddListItem
				e.editBuffer = ""
			}
		}
	case "d":
		if len(items) > 0 && e.cursor >= 0 && e.cursor < len(items) {
			e.mode = modeConfirmDelete
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

	// Input prompts for add/delete modes.
	switch e.mode {
	case modeAddKey:
		b.WriteString(fmt.Sprintf("  New key: %s█\n\n", e.editBuffer))
		b.WriteString("  enter: next  esc: cancel\n")
		return b.String()
	case modeAddValue:
		b.WriteString(fmt.Sprintf("  Key: %s\n", e.addKeyBuf))
		b.WriteString(fmt.Sprintf("  Value: %s█\n\n", e.editBuffer))
		b.WriteString("  enter: add  esc: cancel\n")
		return b.String()
	case modeAddListItem:
		b.WriteString(fmt.Sprintf("  New item: %s█\n\n", e.editBuffer))
		b.WriteString("  enter: add  esc: cancel\n")
		return b.String()
	case modeConfirmDelete:
		items := e.items()
		if e.cursor >= 0 && e.cursor < len(items) {
			b.WriteString(fmt.Sprintf("  Delete '%s'? (y/n)\n", items[e.cursor].key))
		}
		return b.String()
	}

	items := e.items()
	if len(items) == 0 {
		b.WriteString("  (empty)\n")
		b.WriteString("\n  +: add  esc: back\n")
		return b.String()
	}

	for i, item := range items {
		cursor := "  "
		if i == e.cursor {
			cursor = "▸ "
		}
		if e.mode == modeEdit && i == e.cursor {
			b.WriteString(fmt.Sprintf("%s%s: %s█\n", cursor, item.key, e.editBuffer))
		} else if item.isScalar && isBoolNode(item.node) {
			status := "OFF"
			if item.node.Value == "true" {
				status = " ON"
			}
			b.WriteString(fmt.Sprintf("%s[%s] %s\n", cursor, status, item.key))
		} else {
			b.WriteString(fmt.Sprintf("%s%-20s %s\n", cursor, item.key, item.value))
		}
	}

	b.WriteString("\n  enter: edit/drill  +: add  d: delete  esc: back\n")
	return b.String()
}
