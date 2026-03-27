package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTextInputRender(t *testing.T) {
	ti := &TextInput{Label: "Name", Value: "hello"}
	view := ti.View()
	if !strings.Contains(view, "Name") || !strings.Contains(view, "hello") {
		t.Error("text input should render label and value")
	}
}

func TestTextInputEdit(t *testing.T) {
	ti := &TextInput{Label: "Name", Value: "hel", Focused: true}
	ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l', 'o'}})
	if ti.Value != "hello" {
		t.Errorf("value = %q, want %q", ti.Value, "hello")
	}
}

func TestTextInputBackspace(t *testing.T) {
	ti := &TextInput{Label: "Name", Value: "hello", Focused: true}
	ti.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if ti.Value != "hell" {
		t.Errorf("value = %q, want %q", ti.Value, "hell")
	}
}

func TestTextInputNotFocused(t *testing.T) {
	ti := &TextInput{Label: "Name", Value: "hello", Focused: false}
	ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if ti.Value != "hello" {
		t.Errorf("unfocused input should not change, got %q", ti.Value)
	}
}

func TestBoolInputToggle(t *testing.T) {
	bi := &BoolInput{Label: "Enabled", Val: false}
	bi.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !bi.Value() {
		t.Error("bool should be true after toggle")
	}
	bi.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if bi.Value() {
		t.Error("bool should be false after second toggle")
	}
}

func TestBoolInputRender(t *testing.T) {
	bi := &BoolInput{Label: "Enabled", Val: true}
	view := bi.View()
	if !strings.Contains(view, "[x]") {
		t.Error("bool input should show [x] when true")
	}
	bi.Val = false
	view = bi.View()
	if !strings.Contains(view, "[ ]") {
		t.Error("bool input should show [ ] when false")
	}
}

func TestNumberInputValidation(t *testing.T) {
	ni := &NumberInput{Label: "Port", RawValue: "8080", Min: 1, Max: 65535}
	v, err := ni.IntValue()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if v != 8080 {
		t.Errorf("value = %d, want 8080", v)
	}
}

func TestNumberInputOutOfRange(t *testing.T) {
	ni := &NumberInput{Label: "Port", RawValue: "99999", Min: 1, Max: 65535}
	_, err := ni.IntValue()
	if err == nil {
		t.Error("expected out of range error")
	}
}

func TestNumberInputInvalid(t *testing.T) {
	ni := &NumberInput{Label: "Port", RawValue: "abc", Min: 1, Max: 65535}
	_, err := ni.IntValue()
	if err == nil {
		t.Error("expected invalid number error")
	}
}
