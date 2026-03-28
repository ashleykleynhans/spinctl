package deploy

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/spinnaker/spinctl/internal/model"
)

// countingExecutor wraps an Executor and fails after a certain number of calls.
type countingExecutor struct {
	inner     Executor
	failAfter int
	failErr   error
	mu        sync.Mutex
	count     int
}

func (c *countingExecutor) Run(ctx context.Context, name string, args ...string) error {
	c.mu.Lock()
	c.count++
	n := c.count
	c.mu.Unlock()
	if n > c.failAfter {
		return fmt.Errorf("call %d: %w", n, c.failErr)
	}
	return c.inner.Run(ctx, name, args...)
}

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

func TestDebianDeployerDeployServiceRestartFails(t *testing.T) {
	mock := NewMockExecutor()
	d := NewDebianDeployer(mock, t.TempDir())

	// Make systemctl restart fail by succeeding the first two sudo calls
	// (install and daemon-reload), then failing the third (restart).
	customExec := &countingExecutor{
		inner:     mock,
		failAfter: 2, // fail on 3rd call
		failErr:   errSimulated,
	}
	d2 := NewDebianDeployer(customExec, t.TempDir())

	err := d2.DeployService(context.Background(), model.Orca, "8.47.0")
	if err == nil {
		t.Error("expected error when restart fails")
	}
	// The mock executor with no failures verifies basic command routing.
	_ = d.CheckSudo(context.Background())
}

func TestDebianDeployerDeployServiceDaemonReloadFails(t *testing.T) {
	// Succeed on install (1st call), fail on daemon-reload (2nd call).
	customExec := &countingExecutor{
		inner:     NewMockExecutor(),
		failAfter: 1,
		failErr:   errSimulated,
	}
	d := NewDebianDeployer(customExec, t.TempDir())
	err := d.DeployService(context.Background(), model.Orca, "8.47.0")
	if err == nil {
		t.Error("expected error when daemon-reload fails")
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
