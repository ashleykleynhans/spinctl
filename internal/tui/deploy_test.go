package tui

import (
	"fmt"
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

func TestDeployStatusString(t *testing.T) {
	tests := []struct {
		status DeployStatus
		want   string
	}{
		{Pending, "pending"},
		{Installing, "installing"},
		{Restarting, "restarting"},
		{Done, "done"},
		{Failed, "failed"},
		{DeployStatus(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeployStatusIcon(t *testing.T) {
	tests := []struct {
		status DeployStatus
		want   string
	}{
		{Pending, "\u25cb"},
		{Installing, "\u25d0"},
		{Restarting, "\u25d1"},
		{Done, "\u25cf"},
		{Failed, "\u2717"},
		{DeployStatus(99), "?"},
	}
	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			if got := tt.status.icon(); got != tt.want {
				t.Errorf("icon() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeployPageConfirmedView(t *testing.T) {
	plan := &deploy.DeployPlan{
		Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Front50}}},
	}
	dp := NewDeployPage(plan)
	dp.confirmed = true
	view := dp.View()
	if !strings.Contains(view, "Deploying") {
		t.Error("confirmed page should show 'Deploying'")
	}
}

func TestDeployPageDoneSuccess(t *testing.T) {
	plan := &deploy.DeployPlan{
		Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Front50}}},
	}
	dp := NewDeployPage(plan)
	dp.done = true
	view := dp.View()
	if !strings.Contains(view, "complete") {
		t.Error("done page should show 'complete'")
	}
}

func TestDeployPageDoneWithError(t *testing.T) {
	plan := &deploy.DeployPlan{
		Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Front50}}},
	}
	dp := NewDeployPage(plan)
	dp.done = true
	dp.err = fmt.Errorf("something went wrong")
	view := dp.View()
	if !strings.Contains(view, "failed") {
		t.Error("done page with error should show 'failed'")
	}
	if !strings.Contains(view, "something went wrong") {
		t.Error("done page should show the error message")
	}
}

func TestDeployPageCancelWithEsc(t *testing.T) {
	plan := &deploy.DeployPlan{
		Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Front50}}},
	}
	dp := NewDeployPage(plan)
	dp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if !dp.cancelled {
		t.Error("should be cancelled after esc")
	}
}

func TestDeployPageIgnoresKeysAfterConfirm(t *testing.T) {
	plan := &deploy.DeployPlan{
		Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Front50}}},
	}
	dp := NewDeployPage(plan)
	dp.confirmed = true
	dp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if dp.cancelled {
		t.Error("should not cancel after already confirmed")
	}
}

func TestDeployPageStatusesMatchSteps(t *testing.T) {
	plan := &deploy.DeployPlan{
		Steps: []deploy.DeployStep{
			{Services: []model.ServiceName{model.Front50}},
			{Services: []model.ServiceName{model.Gate}},
		},
	}
	dp := NewDeployPage(plan)
	if len(dp.statuses) != 2 {
		t.Errorf("statuses len = %d, want 2", len(dp.statuses))
	}
}
