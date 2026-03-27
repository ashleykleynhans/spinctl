package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/deploy"
	"github.com/spinnaker/spinctl/internal/model"
)

func TestDeployPageBuilding(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	dp := NewDeployPage(cfg)

	if !dp.building {
		t.Error("should start in building state")
	}

	view := dp.View()
	if !strings.Contains(view, "Building") {
		t.Error("view should show building message")
	}
}

func TestDeployPageBuildSuccess(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	dp := NewDeployPage(cfg)

	// Simulate build completion.
	plan := &deploy.DeployPlan{
		Steps: []deploy.DeployStep{
			{Services: []model.ServiceName{model.Gate}},
		},
	}
	versions := map[model.ServiceName]string{model.Gate: "6.62.0"}
	dp.Update(deployBuildMsg{plan: plan, versions: versions})

	if dp.building {
		t.Error("should not be building after build msg")
	}
	if dp.plan == nil {
		t.Error("plan should be set")
	}

	view := dp.View()
	if !strings.Contains(view, "gate") {
		t.Error("should show gate in plan")
	}
	if !strings.Contains(view, "6.62.0") {
		t.Error("should show version")
	}
	if !strings.Contains(view, "deploy") {
		t.Error("should show deploy prompt")
	}
}

func TestDeployPageBuildError(t *testing.T) {
	cfg := config.NewDefault()
	dp := NewDeployPage(cfg)

	dp.Update(deployBuildMsg{err: fmt.Errorf("BOM not found")})

	if dp.building {
		t.Error("should not be building")
	}
	view := dp.View()
	if !strings.Contains(view, "BOM not found") {
		t.Error("should show error")
	}
}

func TestDeployPageCancel(t *testing.T) {
	cfg := config.NewDefault()
	dp := NewDeployPage(cfg)

	// Complete build first.
	dp.Update(deployBuildMsg{
		plan:     &deploy.DeployPlan{Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Gate}}}},
		versions: map[model.ServiceName]string{model.Gate: "1.0.0"},
	})

	dp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !dp.cancelled {
		t.Error("should be cancelled after 'n'")
	}
	view := dp.View()
	if !strings.Contains(view, "cancelled") {
		t.Error("should show cancelled message")
	}
}

func TestDeployPageCancelWithEsc(t *testing.T) {
	cfg := config.NewDefault()
	dp := NewDeployPage(cfg)

	dp.Update(deployBuildMsg{
		plan:     &deploy.DeployPlan{Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Gate}}}},
		versions: map[model.ServiceName]string{model.Gate: "1.0.0"},
	})

	dp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if !dp.cancelled {
		t.Error("should be cancelled after esc")
	}
}

func TestDeployPageConfirm(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	dp := NewDeployPage(cfg)

	dp.Update(deployBuildMsg{
		plan:     &deploy.DeployPlan{Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Gate}}}},
		versions: map[model.ServiceName]string{model.Gate: "6.62.0"},
	})

	_, cmd := dp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if !dp.confirmed {
		t.Error("should be confirmed after 'y'")
	}
	if !dp.deploying {
		t.Error("should be deploying")
	}
	if cmd == nil {
		t.Error("should return deploy command")
	}
}

func TestDeployPageDoneSuccess(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	dp := NewDeployPage(cfg)
	dp.building = false
	dp.confirmed = true
	dp.plan = &deploy.DeployPlan{Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Gate}}}}
	dp.versions = map[model.ServiceName]string{model.Gate: "6.62.0"}

	dp.Update(deployDoneMsg{
		results: []deploy.DeployResult{
			{Service: model.Gate, Version: "6.62.0"},
		},
	})

	if !dp.done {
		t.Error("should be done")
	}
	view := dp.View()
	if !strings.Contains(view, "complete") {
		t.Error("should show complete")
	}
}

func TestDeployPageDoneError(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	dp := NewDeployPage(cfg)
	dp.building = false
	dp.confirmed = true
	dp.plan = &deploy.DeployPlan{Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Gate}}}}
	dp.versions = map[model.ServiceName]string{model.Gate: "6.62.0"}

	dp.Update(deployDoneMsg{err: fmt.Errorf("install failed")})

	view := dp.View()
	if !strings.Contains(view, "failed") {
		t.Error("should show failure")
	}
}

func TestDeployPageIgnoresKeysDuringBuild(t *testing.T) {
	cfg := config.NewDefault()
	dp := NewDeployPage(cfg)

	dp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if dp.confirmed {
		t.Error("should not confirm while building")
	}
}

func TestDeployPageIgnoresKeysDuringDeploy(t *testing.T) {
	cfg := config.NewDefault()
	dp := NewDeployPage(cfg)
	dp.building = false
	dp.deploying = true

	dp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if dp.cancelled {
		t.Error("should not cancel while deploying")
	}
}

func TestDeployPageStepMsg(t *testing.T) {
	cfg := config.NewDefault()
	dp := NewDeployPage(cfg)
	dp.building = false

	dp.Update(deployStepMsg{
		service: model.Gate,
		result:  deploy.DeployResult{Service: model.Gate, Version: "6.62.0"},
		stepIdx: 0,
	})

	if dp.statuses[model.Gate] != Done {
		t.Errorf("gate status = %v, want Done", dp.statuses[model.Gate])
	}
}

func TestDeployPageStepMsgFailed(t *testing.T) {
	cfg := config.NewDefault()
	dp := NewDeployPage(cfg)
	dp.building = false

	dp.Update(deployStepMsg{
		service: model.Gate,
		result:  deploy.DeployResult{Service: model.Gate, Err: fmt.Errorf("fail")},
		stepIdx: 0,
	})

	if dp.statuses[model.Gate] != Failed {
		t.Errorf("gate status = %v, want Failed", dp.statuses[model.Gate])
	}
}

func TestDeployPageWarnings(t *testing.T) {
	cfg := config.NewDefault()
	dp := NewDeployPage(cfg)

	dp.Update(deployBuildMsg{
		plan: &deploy.DeployPlan{
			Steps:    []deploy.DeployStep{{Services: []model.ServiceName{model.Gate}}},
			Warnings: []string{"gate depends on orca"},
		},
		versions: map[model.ServiceName]string{model.Gate: "1.0.0"},
	})

	view := dp.View()
	if !strings.Contains(view, "orca") {
		t.Error("should show warning about orca dependency")
	}
}

func TestDeployPageShowsVersion(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	dp := NewDeployPage(cfg)

	dp.Update(deployBuildMsg{
		plan:     &deploy.DeployPlan{Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Gate}}}},
		versions: map[model.ServiceName]string{model.Gate: "6.62.0"},
	})

	view := dp.View()
	if !strings.Contains(view, "1.35.0") {
		t.Error("should show spinnaker version")
	}
}

func TestDeployStatusStrings(t *testing.T) {
	tests := []struct {
		status DeployStatus
		str    string
		icon   string
	}{
		{Pending, "pending", "○"},
		{Installing, "installing", "◐"},
		{Restarting, "restarting", "◑"},
		{Done, "done", "●"},
		{Failed, "failed", "✗"},
		{DeployStatus(99), "unknown", "?"},
	}
	for _, tt := range tests {
		if tt.status.String() != tt.str {
			t.Errorf("%v.String() = %q, want %q", tt.status, tt.status.String(), tt.str)
		}
		if tt.status.icon() != tt.icon {
			t.Errorf("%v.icon() = %q, want %q", tt.status, tt.status.icon(), tt.icon)
		}
	}
}

func TestDeployPageDeployingView(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	dp := NewDeployPage(cfg)
	dp.building = false
	dp.confirmed = true
	dp.deploying = true
	dp.plan = &deploy.DeployPlan{Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Gate}}}}
	dp.versions = map[model.ServiceName]string{model.Gate: "6.62.0"}

	view := dp.View()
	if !strings.Contains(view, "Deploying") {
		t.Error("should show deploying message")
	}
}

func TestDeployPageServiceStatusIcons(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	dp := NewDeployPage(cfg)
	dp.building = false
	dp.plan = &deploy.DeployPlan{Steps: []deploy.DeployStep{
		{Services: []model.ServiceName{model.Gate, model.Orca}},
	}}
	dp.versions = map[model.ServiceName]string{model.Gate: "6.62.0", model.Orca: "8.47.0"}
	dp.statuses[model.Gate] = Done
	dp.statuses[model.Orca] = Failed

	view := dp.View()
	if !strings.Contains(view, "✓") {
		t.Error("should show done checkmark for gate")
	}
	if !strings.Contains(view, "✗") {
		t.Error("should show failed X for orca")
	}
}

func TestDeployPageInstallingStatus(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	dp := NewDeployPage(cfg)
	dp.building = false
	dp.plan = &deploy.DeployPlan{Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Gate}}}}
	dp.versions = map[model.ServiceName]string{model.Gate: "6.62.0"}
	dp.statuses[model.Gate] = Installing

	view := dp.View()
	if !strings.Contains(view, "◐") {
		t.Error("should show installing icon")
	}
}

func TestDeployPageUnknownVersion(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	dp := NewDeployPage(cfg)
	dp.building = false
	dp.plan = &deploy.DeployPlan{Steps: []deploy.DeployStep{{Services: []model.ServiceName{model.Gate}}}}
	dp.versions = map[model.ServiceName]string{}

	view := dp.View()
	if !strings.Contains(view, "unknown") {
		t.Error("should show 'unknown' for missing version")
	}
}

func TestDeployPageNonKeyMsg(t *testing.T) {
	cfg := config.NewDefault()
	dp := NewDeployPage(cfg)
	type customMsg struct{}
	result, cmd := dp.Update(customMsg{})
	if result != dp {
		t.Error("non-key msg should return same page")
	}
	if cmd != nil {
		t.Error("non-key msg should return nil cmd")
	}
}
