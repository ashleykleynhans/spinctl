package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func testFields() []FieldDef {
	return []FieldDef{
		{Name: "name", Label: "Account name", Required: true},
		{Name: "region", Label: "Region", Default: "us-west-2"},
		{Name: "secret", Label: "Secret key", Secret: true},
	}
}

func TestNewFieldForm(t *testing.T) {
	ff := NewFieldForm("Test", testFields())
	if len(ff.fields) != 3 {
		t.Errorf("fields count = %d, want 3", len(ff.fields))
	}
	if ff.values["region"] != "us-west-2" {
		t.Errorf("region default = %q, want us-west-2", ff.values["region"])
	}
	if ff.buffer != "" {
		t.Errorf("initial buffer = %q, want empty (first field has no default)", ff.buffer)
	}
}

func TestNewFieldFormDefaultBuffer(t *testing.T) {
	fields := []FieldDef{
		{Name: "host", Label: "Host", Default: "localhost"},
	}
	ff := NewFieldForm("Test", fields)
	if ff.buffer != "localhost" {
		t.Errorf("buffer = %q, want localhost", ff.buffer)
	}
}

func TestFieldFormUpdateEnter(t *testing.T) {
	ff := NewFieldForm("Test", testFields())
	// Type "myaccount" into the first field.
	typeText(ff, "myaccount")
	ff.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ff.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after enter", ff.cursor)
	}
	if ff.values["name"] != "myaccount" {
		t.Errorf("name = %q, want myaccount", ff.values["name"])
	}
}

func TestFieldFormUpdateEnterRequired(t *testing.T) {
	ff := NewFieldForm("Test", testFields())
	// Press enter with empty buffer on required field.
	ff.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ff.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (should not advance on empty required)", ff.cursor)
	}
}

func TestFieldFormUpdateEscSkipsOptional(t *testing.T) {
	ff := NewFieldForm("Test", testFields())
	// Fill required field first.
	typeText(ff, "myaccount")
	ff.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Now on "region" (optional), press esc to skip.
	ff.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if ff.cursor != 2 {
		t.Errorf("cursor = %d, want 2 after esc skip", ff.cursor)
	}
	if ff.values["region"] != "" {
		t.Errorf("region = %q, want empty after skip", ff.values["region"])
	}
}

func TestFieldFormUpdateBackspace(t *testing.T) {
	ff := NewFieldForm("Test", testFields())
	typeText(ff, "abc")
	ff.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if ff.buffer != "ab" {
		t.Errorf("buffer = %q, want ab", ff.buffer)
	}
}

func TestFieldFormUpdateRunes(t *testing.T) {
	ff := NewFieldForm("Test", testFields())
	typeText(ff, "hello")
	if ff.buffer != "hello" {
		t.Errorf("buffer = %q, want hello", ff.buffer)
	}
}

func TestFieldFormCompletion(t *testing.T) {
	ff := NewFieldForm("Test", testFields())
	// Fill "name" (required).
	typeText(ff, "acct")
	ff.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Fill "region" (optional with default).
	ff.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Fill "secret" (optional).
	_, cmd := ff.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected fieldFormDoneMsg command on completion")
	}
	msg := cmd()
	done, ok := msg.(fieldFormDoneMsg)
	if !ok {
		t.Fatalf("expected fieldFormDoneMsg, got %T", msg)
	}
	if done.values["name"] != "acct" {
		t.Errorf("done name = %q, want acct", done.values["name"])
	}
	if !ff.done {
		t.Error("form should be marked done")
	}
}

func TestFieldFormViewShowsFields(t *testing.T) {
	ff := NewFieldForm("Test Form", testFields())
	view := ff.View()
	if !strings.Contains(view, "Account name") {
		t.Error("view should show first field label")
	}
	if !strings.Contains(view, "Test Form") {
		t.Error("view should show title")
	}
}

func TestFieldFormViewShowsRequired(t *testing.T) {
	ff := NewFieldForm("Test", testFields())
	view := ff.View()
	// Required fields show * marker (rendered through warnStyle).
	if !strings.Contains(view, "*") {
		t.Error("view should show * for required fields")
	}
}

func TestFieldFormViewShowsSecret(t *testing.T) {
	ff := NewFieldForm("Test", testFields())
	// Advance to secret field and type something.
	typeText(ff, "acct")
	ff.Update(tea.KeyMsg{Type: tea.KeyEnter})
	ff.Update(tea.KeyMsg{Type: tea.KeyEnter}) // accept region default
	typeText(ff, "mysecret")

	view := ff.View()
	if strings.Contains(view, "mysecret") {
		t.Error("view should not show secret value in plaintext")
	}
	if !strings.Contains(view, "***") {
		t.Error("view should show masked secret value")
	}
}

func TestFieldFormViewShowsDefault(t *testing.T) {
	ff := NewFieldForm("Test", testFields())
	// Advance to region field which has a default.
	typeText(ff, "acct")
	ff.Update(tea.KeyMsg{Type: tea.KeyEnter})

	view := ff.View()
	if !strings.Contains(view, "default: us-west-2") {
		t.Error("view should show default value hint")
	}
}

func TestFieldFormEscOnRequiredSendsGoBack(t *testing.T) {
	ff := NewFieldForm("Test", testFields())
	// First field is required; esc should send goBackMsg.
	_, cmd := ff.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected goBackMsg command on esc from required field")
	}
	msg := cmd()
	if _, ok := msg.(goBackMsg); !ok {
		t.Fatalf("expected goBackMsg, got %T", msg)
	}
	if ff.cursor != 0 {
		t.Errorf("cursor = %d, want 0", ff.cursor)
	}
}

func TestFieldFormEscOnFirstFieldSendsGoBack(t *testing.T) {
	fields := []FieldDef{
		{Name: "opt1", Label: "Optional 1"},
		{Name: "opt2", Label: "Optional 2"},
	}
	ff := NewFieldForm("Test", fields)
	// Even though first field is optional, esc on cursor=0 sends goBack.
	_, cmd := ff.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected goBackMsg on esc from first field")
	}
	msg := cmd()
	if _, ok := msg.(goBackMsg); !ok {
		t.Fatalf("expected goBackMsg, got %T", msg)
	}
}

func TestFieldFormCompletionViaEsc(t *testing.T) {
	fields := []FieldDef{
		{Name: "name", Label: "Name", Required: true},
		{Name: "opt", Label: "Optional"},
	}
	ff := NewFieldForm("Test", fields)
	typeText(ff, "val")
	ff.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Esc on last optional field completes the form.
	_, cmd := ff.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected done command when esc completes form")
	}
	msg := cmd()
	if _, ok := msg.(fieldFormDoneMsg); !ok {
		t.Fatalf("expected fieldFormDoneMsg, got %T", msg)
	}
}

// typeText sends runes one at a time.
func typeText(ff *FieldForm, s string) {
	for _, r := range s {
		ff.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}
