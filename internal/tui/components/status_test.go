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

func TestStatusBarNarrowWidth(t *testing.T) {
	// Very narrow width should not panic.
	sb := NewStatusBar(5)
	sb.SetModified(true)
	view := sb.View("q: quit")
	if view == "" {
		t.Error("narrow status bar should still render something")
	}
}

func TestStatusBarZeroWidth(t *testing.T) {
	sb := NewStatusBar(0)
	view := sb.View("hints")
	if view == "" {
		t.Error("zero-width status bar should still render something")
	}
}

func TestStatusBarToggleModified(t *testing.T) {
	sb := NewStatusBar(80)
	sb.SetModified(true)
	view1 := sb.View("q: quit")
	if !strings.Contains(view1, "modified") {
		t.Error("should show modified after SetModified(true)")
	}
	sb.SetModified(false)
	view2 := sb.View("q: quit")
	if strings.Contains(view2, "modified") {
		t.Error("should not show modified after SetModified(false)")
	}
}
