package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

func TestServicesPageShowsAllServices(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)
	view := sp.View()
	for _, name := range model.AllServiceNames() {
		if !strings.Contains(view, name.String()) {
			t.Errorf("services page should show %s", name)
		}
	}
}

func TestServicesPageNavigation(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)

	if sp.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", sp.cursor)
	}

	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if sp.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", sp.cursor)
	}

	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if sp.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", sp.cursor)
	}
}

func TestServicesPageToggle(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)

	name := sp.sortedNames[0]
	if cfg.Services[name].Enabled {
		t.Error("service should start disabled")
	}

	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !cfg.Services[name].Enabled {
		t.Error("service should be enabled after toggle")
	}

	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cfg.Services[name].Enabled {
		t.Error("service should be disabled after second toggle")
	}
}

func TestServicesPageEnabledDisplay(t *testing.T) {
	cfg := config.NewDefault()
	// Enable one service.
	svc := cfg.Services[model.Gate]
	svc.Enabled = true
	cfg.Services[model.Gate] = svc

	sp := NewServicesPage(cfg)
	view := sp.View()
	if !strings.Contains(view, " ON") {
		t.Error("should display ON for enabled service")
	}
	if !strings.Contains(view, "OFF") {
		t.Error("should display OFF for disabled services")
	}
}

func TestServicesPageSorted(t *testing.T) {
	cfg := config.NewDefault()
	sp := NewServicesPage(cfg)

	for i := 1; i < len(sp.sortedNames); i++ {
		if sp.sortedNames[i].String() < sp.sortedNames[i-1].String() {
			t.Errorf("services not sorted: %s before %s", sp.sortedNames[i-1], sp.sortedNames[i])
		}
	}
}
