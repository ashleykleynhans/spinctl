package deploy

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

func TestDebianDeployerCheckSudo(t *testing.T) {
	mock := NewMockExecutor()
	d := NewDebianDeployer(mock, t.TempDir())

	if err := d.CheckSudo(context.Background()); err != nil {
		t.Fatalf("CheckSudo: %v", err)
	}

	if !mock.HasCommand("sudo -n true") {
		t.Error("expected sudo -n true command")
	}
}

func TestDebianDeployerUpdateApt(t *testing.T) {
	mock := NewMockExecutor()
	d := NewDebianDeployer(mock, t.TempDir())

	if err := d.UpdateApt(context.Background()); err != nil {
		t.Fatalf("UpdateApt: %v", err)
	}

	if !mock.HasCommand("sudo apt-get update -qq") {
		t.Error("expected apt-get update command")
	}
}

func TestDebianDeployerDeployService(t *testing.T) {
	mock := NewMockExecutor()
	d := NewDebianDeployer(mock, t.TempDir())

	err := d.DeployService(context.Background(), model.Orca, "8.47.0")
	if err != nil {
		t.Fatalf("DeployService: %v", err)
	}

	if !mock.HasCommand("sudo apt-get install -y -qq spinnaker-orca=8.47.0") {
		t.Error("expected apt-get install command for orca")
	}
	if !mock.HasCommand("sudo systemctl restart orca.service") {
		t.Error("expected systemctl restart for orca")
	}
}

func TestDebianDeployerDeployDeckSkipsSystemd(t *testing.T) {
	mock := NewMockExecutor()
	d := NewDebianDeployer(mock, t.TempDir())

	err := d.DeployService(context.Background(), model.Deck, "3.16.0")
	if err != nil {
		t.Fatalf("DeployService(deck): %v", err)
	}

	if mock.HasCommand("systemctl") {
		t.Error("deck should not trigger systemctl commands")
	}
}

func TestDebianDeployerWriteServiceConfig(t *testing.T) {
	dir := t.TempDir()
	mock := NewMockExecutor()
	d := NewDebianDeployer(mock, dir)

	cfg := config.ServiceConfig{
		Enabled: true,
		Host:    "localhost",
		Port:    8083,
	}

	if err := d.WriteServiceConfig(model.Orca, cfg); err != nil {
		t.Fatalf("WriteServiceConfig: %v", err)
	}

	path := filepath.Join(dir, "orca", "orca.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading config file: %v", err)
	}

	if len(data) == 0 {
		t.Error("config file is empty")
	}
}

func TestBuildDeployPlanOrdering(t *testing.T) {
	plan := BuildDeployPlan(nil)

	if len(plan.Steps) == 0 {
		t.Fatal("expected non-empty plan")
	}

	// First step should be front50.
	first := plan.Steps[0].Services
	if len(first) != 1 || first[0] != model.Front50 {
		t.Errorf("first step = %v, want [front50]", first)
	}

	// Last step should be deck.
	last := plan.Steps[len(plan.Steps)-1].Services
	if len(last) != 1 || last[0] != model.Deck {
		t.Errorf("last step = %v, want [deck]", last)
	}
}

func TestDebianDeployerDeployServiceInstallFails(t *testing.T) {
	mock := NewMockExecutor()
	mock.SetFail("sudo", errSimulated)
	d := NewDebianDeployer(mock, t.TempDir())

	err := d.DeployService(context.Background(), model.Orca, "8.47.0")
	if err == nil {
		t.Error("expected error when install fails")
	}
}

func TestBuildDeployPlanFilterWithWarnings(t *testing.T) {
	filter := []model.ServiceName{model.Gate}
	plan := BuildDeployPlan(filter)

	if len(plan.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(plan.Steps))
	}

	if plan.Steps[0].Services[0] != model.Gate {
		t.Errorf("expected gate in plan, got %v", plan.Steps[0].Services)
	}

	if len(plan.Warnings) == 0 {
		t.Error("expected warnings about missing dependencies for gate")
	}

	// Verify at least one warning mentions a dependency.
	found := false
	for _, w := range plan.Warnings {
		if findSubstring(w, "depends on") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("no dependency warning found in: %v", plan.Warnings)
	}
}
