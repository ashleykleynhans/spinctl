package components

import (
	"strings"
	"testing"
)

func TestStatusBarRender(t *testing.T) {
	sb := NewStatusBar(80)
	view := sb.View("q: quit  ?: help")
	if !strings.Contains(view, "quit") {
		t.Error("status bar should contain hints text")
	}
}

func TestStatusBarModified(t *testing.T) {
	sb := NewStatusBar(80)
	sb.SetModified(true)
	view := sb.View("q: quit")
	if !strings.Contains(view, "modified") {
		t.Error("status bar should show modified indicator")
	}
}

func TestStatusBarNotModified(t *testing.T) {
	sb := NewStatusBar(80)
	sb.SetModified(false)
	view := sb.View("q: quit")
	if strings.Contains(view, "modified") {
		t.Error("status bar should not show modified indicator when not modified")
	}
}
