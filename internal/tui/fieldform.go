package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// fieldFormDoneMsg is sent when the field form is completed.
type fieldFormDoneMsg struct {
	values map[string]string
}

// FieldForm walks through a list of FieldDefs, prompting for each.
type FieldForm struct {
	title  string
	fields []FieldDef
	values map[string]string
	cursor int
	buffer string
	done   bool
}

// NewFieldForm creates a form for the given fields.
func NewFieldForm(title string, fields []FieldDef) *FieldForm {
	values := make(map[string]string, len(fields))
	for _, f := range fields {
		values[f.Name] = f.Default
	}
	ff := &FieldForm{
		title:  title,
		fields: fields,
		values: values,
	}
	if len(fields) > 0 {
		ff.buffer = fields[0].Default
	}
	return ff
}

func (f *FieldForm) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Save current field.
			if f.cursor < len(f.fields) {
				field := f.fields[f.cursor]
				if f.buffer == "" && field.Required {
					// Don't advance if required field is empty.
					return f, nil
				}
				f.values[field.Name] = f.buffer
				f.cursor++
				if f.cursor >= len(f.fields) {
					f.done = true
					return f, func() tea.Msg {
						return fieldFormDoneMsg{values: f.values}
					}
				}
				// Load next field's default/current value.
				next := f.fields[f.cursor]
				f.buffer = f.values[next.Name]
			}
		case tea.KeyEscape:
			// Skip optional field or cancel.
			if f.cursor < len(f.fields) && !f.fields[f.cursor].Required {
				f.values[f.fields[f.cursor].Name] = ""
				f.cursor++
				if f.cursor >= len(f.fields) {
					f.done = true
					return f, func() tea.Msg {
						return fieldFormDoneMsg{values: f.values}
					}
				}
				next := f.fields[f.cursor]
				f.buffer = f.values[next.Name]
			}
		case tea.KeyBackspace:
			if len(f.buffer) > 0 {
				f.buffer = f.buffer[:len(f.buffer)-1]
			}
		case tea.KeyRunes:
			f.buffer += string(msg.Runes)
		}
	}
	return f, nil
}

func (f *FieldForm) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headingStyle.Render(f.title))
	b.WriteString("\n\n")

	for i, field := range f.fields {
		if i > f.cursor {
			break
		}

		reqMark := ""
		if field.Required {
			reqMark = warnStyle.Render("*")
		}

		if i < f.cursor {
			// Already filled.
			val := f.values[field.Name]
			if field.Secret && val != "" {
				val = strings.Repeat("*", len(val))
			}
			if val == "" {
				val = valueStyle.Render("(skipped)")
			} else {
				val = onStyle.Render(val)
			}
			b.WriteString("  " + reqMark + keyStyle.Render(field.Label+": ") + val + "\n")
		} else {
			// Current field being edited.
			display := f.buffer
			if field.Secret {
				display = strings.Repeat("*", len(f.buffer))
			}
			b.WriteString("  " + reqMark + keyStyle.Render(field.Label+": ") + editCursorStyle.Render(display+"█") + "\n")
			if field.Default != "" {
				b.WriteString("    " + menuDescStyle.Render("default: "+field.Default) + "\n")
			}
		}
	}

	b.WriteString("\n")
	if f.cursor < len(f.fields) {
		hints := "enter: confirm"
		if !f.fields[f.cursor].Required {
			hints += "  esc: skip"
		}
		b.WriteString("  " + menuDescStyle.Render(hints) + "\n")
	}

	return b.String()
}
