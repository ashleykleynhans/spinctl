package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/deploy"
	"github.com/spinnaker/spinctl/internal/model"
)

func TestDeployPageShowsPlan(t *testing.T) {
	plan := &deploy.DeployPlan{
		Steps: []deploy.DeployStep{
			{Services: []model.ServiceName{model.Front50}},
			{Services: []model.ServiceName{model.Gate, model.Orca}},
		},
	}
	dp := NewDeployPage(plan)
	view := dp.View()
	if !strings.Contains(view, "front50") {
		t.Error("should show front50 in deploy plan")
	}
	if !strings.Contains(view, "gate") {
		t.Error("should show gate in deploy plan")
	}
	if !strings.Contains(view, "Step 1") {
		t.Error("should show step numbers")
	}
}

func TestDeployPageShowsWarnings(t *testing.T) {
	plan := &deploy.DeployPlan{
		Steps:    []deploy.DeployStep{{Services: []model.ServiceName{model.Gate}}},
		Warnings: []string{"gate depends on orca"},
	}
	dp := NewDeployPage(plan)
	view := dp.View()
	if !strings.Contains(view, "gate depends on orca") {
		t.Error("should show warning")
	}
}

func TestDeployPageConfirm(t *testing.T) {
	plan := &deploy.DeployPlan{
		Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Front50}}},
	}
	dp := NewDeployPage(plan)
	dp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if !dp.confirmed {
		t.Error("should be confirmed after 'y'")
	}
}

func TestDeployPageCancel(t *testing.T) {
	plan := &deploy.DeployPlan{
		Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Front50}}},
	}
	dp := NewDeployPage(plan)
	dp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !dp.cancelled {
		t.Error("should be cancelled after 'n'")
	}
	view := dp.View()
	if !strings.Contains(view, "cancelled") {
		t.Error("should show cancelled message")
	}
}

func TestDeployPageNoPlan(t *testing.T) {
	dp := NewDeployPage(nil)
	view := dp.View()
	if !strings.Contains(view, "No deploy plan") {
		t.Error("should show no plan message")
	}
}
