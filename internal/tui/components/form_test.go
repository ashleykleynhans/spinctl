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

func TestTextInputFocus(t *testing.T) {
	ti := &TextInput{Label: "Name", Value: "hello"}
	ti.Focus()
	if !ti.Focused {
		t.Error("Focus() should set Focused to true")
	}
	view := ti.View()
	if !strings.Contains(view, "\u2588") {
		t.Error("focused input should show cursor block")
	}
}

func TestTextInputBlur(t *testing.T) {
	ti := &TextInput{Label: "Name", Value: "hello", Focused: true}
	ti.Blur()
	if ti.Focused {
		t.Error("Blur() should set Focused to false")
	}
	view := ti.View()
	if strings.Contains(view, "\u2588") {
		t.Error("blurred input should not show cursor block")
	}
}

func TestTextInputBackspaceOnEmpty(t *testing.T) {
	ti := &TextInput{Label: "Name", Value: "", Focused: true}
	ti.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if ti.Value != "" {
		t.Errorf("backspace on empty value should stay empty, got %q", ti.Value)
	}
}

func TestNumberInputFocus(t *testing.T) {
	ni := &NumberInput{Label: "Port", RawValue: "8080", Min: 1, Max: 65535}
	ni.Focus()
	if !ni.Focused {
		t.Error("Focus() should set Focused to true")
	}
}

func TestNumberInputBlur(t *testing.T) {
	ni := &NumberInput{Label: "Port", RawValue: "8080", Min: 1, Max: 65535, Focused: true}
	ni.Blur()
	if ni.Focused {
		t.Error("Blur() should set Focused to false")
	}
}

func TestNumberInputSetRawValue(t *testing.T) {
	ni := &NumberInput{Label: "Port", Min: 1, Max: 65535}
	ni.SetRawValue("9090")
	if ni.RawValue != "9090" {
		t.Errorf("RawValue = %q, want '9090'", ni.RawValue)
	}
	v, err := ni.IntValue()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if v != 9090 {
		t.Errorf("IntValue() = %d, want 9090", v)
	}
}

func TestNumberInputUpdate(t *testing.T) {
	ni := &NumberInput{Label: "Port", RawValue: "808", Min: 1, Max: 65535, Focused: true}
	ni.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
	if ni.RawValue != "8080" {
		t.Errorf("RawValue = %q, want '8080'", ni.RawValue)
	}
}

func TestNumberInputUpdateBackspace(t *testing.T) {
	ni := &NumberInput{Label: "Port", RawValue: "8080", Min: 1, Max: 65535, Focused: true}
	ni.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if ni.RawValue != "808" {
		t.Errorf("RawValue = %q, want '808'", ni.RawValue)
	}
}

func TestNumberInputUpdateNotFocused(t *testing.T) {
	ni := &NumberInput{Label: "Port", RawValue: "8080", Min: 1, Max: 65535, Focused: false}
	ni.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if ni.RawValue != "8080" {
		t.Errorf("unfocused input should not change, got %q", ni.RawValue)
	}
}

func TestNumberInputUpdateBackspaceEmpty(t *testing.T) {
	ni := &NumberInput{Label: "Port", RawValue: "", Min: 1, Max: 65535, Focused: true}
	ni.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if ni.RawValue != "" {
		t.Errorf("backspace on empty should stay empty, got %q", ni.RawValue)
	}
}

func TestNumberInputView(t *testing.T) {
	ni := &NumberInput{Label: "Port", RawValue: "8080", Min: 1, Max: 65535, Focused: false}
	view := ni.View()
	if !strings.Contains(view, "Port") || !strings.Contains(view, "8080") {
		t.Error("view should contain label and value")
	}
}

func TestNumberInputViewFocused(t *testing.T) {
	ni := &NumberInput{Label: "Port", RawValue: "8080", Min: 1, Max: 65535, Focused: true}
	view := ni.View()
	if !strings.Contains(view, "\u2588") {
		t.Error("focused number input should show cursor block")
	}
}

func TestBoolInputNonEnterKey(t *testing.T) {
	bi := &BoolInput{Label: "Enabled", Val: false}
	bi.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if bi.Value() {
		t.Error("non-enter key should not toggle bool")
	}
}

func TestNumberInputBelowMin(t *testing.T) {
	ni := &NumberInput{Label: "Port", RawValue: "0", Min: 1, Max: 65535}
	_, err := ni.IntValue()
	if err == nil {
		t.Error("expected error for value below min")
	}
}
