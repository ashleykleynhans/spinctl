package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

// editorItem represents one visible item in the editor.
type editorItem struct {
	key        string
	value      string
	node       *yaml.Node
	isScalar   bool
	contentIdx int // index in parent's Content array (for maps: key index, for sequences: item index)
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
	root        *yaml.Node
	path        []string
	nodeStack   []*yaml.Node
	cursorStack []int
	current     *yaml.Node
	cursor      int
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
	var enabledItem *editorItem
	var items []editorItem
	for i := 0; i+1 < len(e.current.Content); i += 2 {
		keyNode := e.current.Content[i]
		valNode := e.current.Content[i+1]
		item := editorItem{
			key:        keyNode.Value,
			node:       valNode,
			contentIdx: i,
		}
		switch valNode.Kind {
		case yaml.ScalarNode:
			item.value = valNode.Value
			item.isScalar = true
		case yaml.MappingNode:
			keyCount := len(valNode.Content) / 2
			// Show enabled status if the map has an "enabled" key.
			if enabled := findMapValue(valNode, "enabled"); enabled != "" {
				status := "OFF"
				if enabled == "true" {
					status = " ON"
				}
				item.value = fmt.Sprintf("[%s] {%d keys}", status, keyCount)
			} else if active := e.isActiveByTypeSelector(keyNode.Value); active {
				item.value = fmt.Sprintf("[ ON] {%d keys}", keyCount)
			} else if e.hasTypeSelector() {
				// A type selector exists — this entry is not the active one.
				if keyCount == 0 {
					item.value = "[OFF]"
				} else {
					item.value = fmt.Sprintf("[OFF] {%d keys}", keyCount)
				}
			} else if keyCount == 0 {
				item.value = "(empty)"
			} else {
				item.value = fmt.Sprintf("{%d keys}", keyCount)
			}
		case yaml.SequenceNode:
			item.value = fmt.Sprintf("[%d items]", len(valNode.Content))
		}
		// Sort "enabled" to the top.
		if keyNode.Value == "enabled" {
			copied := item
			enabledItem = &copied
		} else {
			items = append(items, item)
		}
	}
	if enabledItem != nil {
		items = append([]editorItem{*enabledItem}, items...)
	}
	return items
}

// sequenceItems returns indexed items from a sequence node.
// If items are mappings with a "name" key, use the name as the label.
func (e *EditorPage) sequenceItems() []editorItem {
	var items []editorItem
	for i, node := range e.current.Content {
		item := editorItem{
			key:        fmt.Sprintf("[%d]", i),
			node:       node,
			contentIdx: i,
		}
		switch node.Kind {
		case yaml.ScalarNode:
			item.value = node.Value
			item.isScalar = true
		case yaml.MappingNode:
			// Try common identifier keys to use as label.
			if label := findItemLabel(node); label != "" {
				item.key = label
			}
			keyCount := len(node.Content) / 2
			if enabled := findMapValue(node, "enabled"); enabled != "" {
				status := "OFF"
				if enabled == "true" {
					status = " ON"
				}
				item.value = fmt.Sprintf("[%s] {%d keys}", status, keyCount)
			} else if keyCount == 0 {
				item.value = "(empty)"
			} else {
				item.value = fmt.Sprintf("{%d keys}", keyCount)
			}
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

// findItemLabel tries common identifier keys to find a human-readable label
// for a mapping node in a list.
func findItemLabel(node *yaml.Node) string {
	if node.Kind != yaml.MappingNode {
		return ""
	}
	// Try these keys in order of preference.
	labelKeys := []string{"name", "accountName", "id", "title", "key", "region", "label"}
	for _, key := range labelKeys {
		if val := findMapValue(node, key); val != "" {
			return val
		}
	}
	// Fallback: use the first scalar value in the map.
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i+1].Kind == yaml.ScalarNode && node.Content[i+1].Value != "" {
			return node.Content[i+1].Value
		}
	}
	return ""
}

// isActiveByTypeSelector checks if a key's name matches a type-selector value
// in the current mapping node. For example, if the current node has
// "persistentStoreType: s3", then the "s3" entry is considered active.
func (e *EditorPage) isActiveByTypeSelector(name string) bool {
	if e.current == nil || e.current.Kind != yaml.MappingNode {
		return false
	}
	// Look for keys ending in "Type" whose scalar value matches the name.
	for i := 0; i+1 < len(e.current.Content); i += 2 {
		key := e.current.Content[i].Value
		val := e.current.Content[i+1]
		if val.Kind == yaml.ScalarNode && strings.HasSuffix(key, "Type") && val.Value == name {
			return true
		}
	}
	return false
}

// hasTypeSelector checks if the current mapping node has a key ending in "Type".
func (e *EditorPage) hasTypeSelector() bool {
	if e.current == nil || e.current.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i+1 < len(e.current.Content); i += 2 {
		if strings.HasSuffix(e.current.Content[i].Value, "Type") {
			return true
		}
	}
	return false
}

// toggleMapEnabled toggles the "enabled" boolean inside a mapping node.
// Returns true if the toggle was performed.
func toggleMapEnabled(node *yaml.Node) bool {
	if node.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == "enabled" && node.Content[i+1].Kind == yaml.ScalarNode {
			if node.Content[i+1].Value == "true" {
				node.Content[i+1].Value = "false"
			} else {
				node.Content[i+1].Value = "true"
			}
			return true
		}
	}
	return false
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

	item := items[e.cursor]

	switch e.current.Kind {
	case yaml.MappingNode:
		// Use contentIdx to find the correct position in the Content array,
		// since visual order may differ from Content order (enabled sorted to top).
		idx := item.contentIdx
		if idx+1 < len(e.current.Content) {
			e.current.Content = append(e.current.Content[:idx], e.current.Content[idx+2:]...)
		}
	case yaml.SequenceNode:
		idx := item.contentIdx
		if idx < len(e.current.Content) {
			e.current.Content = append(e.current.Content[:idx], e.current.Content[idx+1:]...)
		}
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
	case " ":
		// Space toggles booleans and enabled status on maps.
		if e.cursor >= 0 && e.cursor < len(items) {
			item := items[e.cursor]
			if item.isScalar && isBoolNode(item.node) {
				if item.node.Value == "true" {
					item.node.Value = "false"
				} else {
					item.node.Value = "true"
				}
				return e, func() tea.Msg { return configChangedMsg{} }
			} else if item.node != nil && item.node.Kind == yaml.MappingNode {
				// Toggle the "enabled" key inside the map.
				if toggled := toggleMapEnabled(item.node); toggled {
					return e, func() tea.Msg { return configChangedMsg{} }
				}
			}
		}
	case "enter":
		if e.cursor >= 0 && e.cursor < len(items) {
			item := items[e.cursor]
			if item.isScalar && isBoolNode(item.node) {
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
				e.cursorStack = append(e.cursorStack, e.cursor)
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
			if len(e.cursorStack) > 0 {
				e.cursor = e.cursorStack[len(e.cursorStack)-1]
				e.cursorStack = e.cursorStack[:len(e.cursorStack)-1]
			} else {
				e.cursor = 0
			}
		}
	}
	return e, nil
}

// View renders the editor page.
func (e *EditorPage) View() string {
	var b strings.Builder

	// Breadcrumb.
	b.WriteString("\n")
	if len(e.path) > 0 {
		parts := make([]string, len(e.path))
		for i, p := range e.path {
			if i == len(e.path)-1 {
				parts[i] = headingStyle.Render(p)
			} else {
				parts[i] = valueStyle.Render(p)
			}
		}
		b.WriteString(strings.Join(parts, valueStyle.Render(" > ")))
	} else {
		b.WriteString(headingStyle.Render("root"))
	}
	b.WriteString("\n\n")

	// Input prompts for add/delete modes.
	switch e.mode {
	case modeAddKey:
		b.WriteString("  " + keyStyle.Render("New key: ") + editCursorStyle.Render(e.editBuffer+"█") + "\n\n")
		b.WriteString("  " + menuDescStyle.Render("enter: next  esc: cancel") + "\n")
		return b.String()
	case modeAddValue:
		b.WriteString("  " + keyStyle.Render("Key: ") + valueStyle.Render(e.addKeyBuf) + "\n")
		b.WriteString("  " + keyStyle.Render("Value: ") + editCursorStyle.Render(e.editBuffer+"█") + "\n\n")
		b.WriteString("  " + menuDescStyle.Render("enter: add  esc: cancel") + "\n")
		return b.String()
	case modeAddListItem:
		b.WriteString("  " + keyStyle.Render("New item: ") + editCursorStyle.Render(e.editBuffer+"█") + "\n\n")
		b.WriteString("  " + menuDescStyle.Render("enter: add  esc: cancel") + "\n")
		return b.String()
	case modeConfirmDelete:
		items := e.items()
		if e.cursor >= 0 && e.cursor < len(items) {
			b.WriteString("  " + warnStyle.Render(fmt.Sprintf("Delete '%s'?", items[e.cursor].key)) + "  " + menuDescStyle.Render("y/n") + "\n")
		}
		return b.String()
	}

	items := e.items()
	if len(items) == 0 {
		b.WriteString("  " + valueStyle.Render("(empty)") + "\n")
		b.WriteString("\n  " + menuDescStyle.Render("+: add  esc: back") + "\n")
		return b.String()
	}

	for i, item := range items {
		selected := i == e.cursor
		cursor := "  "
		if selected {
			cursor = menuCursorStyle.Render("▸ ")
		}
		if e.mode == modeEdit && selected {
			b.WriteString(cursor + keySelectedStyle.Render(fmt.Sprintf("%-25s", item.key)) + " " + editCursorStyle.Render(e.editBuffer+"█") + "\n")
		} else if item.isScalar && isBoolNode(item.node) {
			status := offStyle.Render("[OFF]")
			if item.node.Value == "true" {
				status = onStyle.Render("[ ON]")
			}
			label := keyStyle.Render(fmt.Sprintf("%-25s", item.key))
			if selected {
				label = keySelectedStyle.Render(fmt.Sprintf("%-25s", item.key))
			}
			b.WriteString(cursor + label + " " + status + "\n")
		} else {
			label := keyStyle.Render(fmt.Sprintf("%-25s", item.key))
			if selected {
				label = keySelectedStyle.Render(fmt.Sprintf("%-25s", item.key))
			}
			val := valueStyle.Render(item.value)
			// Color ON/OFF badges in item values.
			if strings.Contains(item.value, "[ ON]") {
				val = strings.Replace(item.value, "[ ON]", onStyle.Render("[ ON]"), 1)
				val = strings.Replace(val, "[OFF]", offStyle.Render("[OFF]"), -1)
			} else if strings.Contains(item.value, "[OFF]") {
				val = strings.Replace(item.value, "[OFF]", offStyle.Render("[OFF]"), -1)
			}
			b.WriteString(cursor + label + " " + val + "\n")
		}
	}

	b.WriteString("\n  " + menuDescStyle.Render("enter: edit/drill  space: toggle  +: add  d: delete  esc: back") + "\n")
	return b.String()
}
