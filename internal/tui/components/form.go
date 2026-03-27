package components

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

// TextInput is a simple single-line text input component.
type TextInput struct {
	Label   string
	Value   string
	Focused bool
}

// Focus sets the input as focused.
func (t *TextInput) Focus() {
	t.Focused = true
}

// Blur removes focus from the input.
func (t *TextInput) Blur() {
	t.Focused = false
}

// Update handles key messages for the text input.
func (t *TextInput) Update(msg tea.KeyMsg) {
	if !t.Focused {
		return
	}
	switch msg.Type {
	case tea.KeyBackspace:
		if len(t.Value) > 0 {
			t.Value = t.Value[:len(t.Value)-1]
		}
	case tea.KeyRunes:
		t.Value += string(msg.Runes)
	}
}

// View renders the text input.
func (t *TextInput) View() string {
	cursor := ""
	if t.Focused {
		cursor = "█"
	}
	return fmt.Sprintf("%s: %s%s", t.Label, t.Value, cursor)
}

// BoolInput is a toggleable boolean input component.
type BoolInput struct {
	Label string
	Val   bool
}

// Value returns the current boolean value.
func (b *BoolInput) Value() bool {
	return b.Val
}

// Update handles key messages for the bool input (enter toggles).
func (b *BoolInput) Update(msg tea.KeyMsg) {
	if msg.Type == tea.KeyEnter {
		b.Val = !b.Val
	}
}

// View renders the bool input.
func (b *BoolInput) View() string {
	indicator := "[ ]"
	if b.Val {
		indicator = "[x]"
	}
	return fmt.Sprintf("%s %s", indicator, b.Label)
}

// NumberInput is a numeric text input with min/max validation.
type NumberInput struct {
	Label    string
	RawValue string
	Min      int
	Max      int
	Focused  bool
}

// Focus sets the input as focused.
func (n *NumberInput) Focus() {
	n.Focused = true
}

// Blur removes focus from the input.
func (n *NumberInput) Blur() {
	n.Focused = false
}

// SetRawValue sets the raw string value.
func (n *NumberInput) SetRawValue(s string) {
	n.RawValue = s
}

// IntValue parses and validates the current value, returning the integer
// and an error if the value is not a valid number or is out of range.
func (n *NumberInput) IntValue() (int, error) {
	v, err := strconv.Atoi(n.RawValue)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", n.RawValue)
	}
	if v < n.Min || v > n.Max {
		return 0, fmt.Errorf("value %d out of range [%d, %d]", v, n.Min, n.Max)
	}
	return v, nil
}

// Update handles key messages for the number input.
func (n *NumberInput) Update(msg tea.KeyMsg) {
	if !n.Focused {
		return
	}
	switch msg.Type {
	case tea.KeyBackspace:
		if len(n.RawValue) > 0 {
			n.RawValue = n.RawValue[:len(n.RawValue)-1]
		}
	case tea.KeyRunes:
		n.RawValue += string(msg.Runes)
	}
}

// View renders the number input.
func (n *NumberInput) View() string {
	cursor := ""
	if n.Focused {
		cursor = "█"
	}
	return fmt.Sprintf("%s: %s%s", n.Label, n.RawValue, cursor)
}
