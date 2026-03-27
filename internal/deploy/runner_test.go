package deploy

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

func TestRunDeployAllServices(t *testing.T) {
	mock := NewMockExecutor()
	dir := t.TempDir()
	logFile := filepath.Join(dir, "deploy.log")
	stateFile := filepath.Join(dir, "state.json")

	cfg := config.NewDefault()
	cfg.Services[model.Front50] = config.ServiceConfig{
		Enabled: true,
		Host:    "localhost",
		Port:    8080,
	}

	bom := testBOM()
	runner := NewDeployRunner(mock, filepath.Join(dir, "config"), logFile, stateFile)

	// Deploy only front50 to keep test simple.
	results, err := runner.Run(context.Background(), cfg, bom, testServices(model.Front50))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Service != model.Front50 {
		t.Errorf("service = %v, want front50", results[0].Service)
	}

	if results[0].Err != nil {
		t.Errorf("unexpected error: %v", results[0].Err)
	}

	// Verify log file was written.
	logData, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("reading log: %v", err)
	}
	if len(logData) == 0 {
		t.Error("expected log file to have content")
	}

	// Verify state file was cleaned up.
	if _, err := os.Stat(stateFile); !os.IsNotExist(err) {
		t.Error("expected state file to be removed after success")
	}
}

func TestRunDeployStopsOnFailure(t *testing.T) {
	mock := NewMockExecutor()
	// Make apt-get install fail, which will cause the first service to fail.
	mock.SetFail("sudo", errSimulated)

	dir := t.TempDir()
	logFile := filepath.Join(dir, "deploy.log")
	stateFile := filepath.Join(dir, "state.json")

	cfg := config.NewDefault()
	bom := testBOM()
	runner := NewDeployRunner(mock, filepath.Join(dir, "config"), logFile, stateFile)

	filter := testServices(model.Orca, model.Gate)
	_, err := runner.Run(context.Background(), cfg, bom, filter)
	if err == nil {
		t.Fatal("expected error when service fails")
	}

	// Verify state file was saved.
	state, stateErr := LoadDeployState(stateFile)
	if stateErr != nil {
		t.Fatalf("LoadDeployState: %v", stateErr)
	}

	if len(state.Remaining) == 0 {
		t.Error("expected remaining services in state")
	}
}

func TestRunDeploySignalCancellation(t *testing.T) {
	mock := NewMockExecutor()
	dir := t.TempDir()
	logFile := filepath.Join(dir, "deploy.log")
	stateFile := filepath.Join(dir, "state.json")

	cfg := config.NewDefault()
	bom := testBOM()
	runner := NewDeployRunner(mock, filepath.Join(dir, "config"), logFile, stateFile)

	// Create an already-cancelled context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	filter := testServices(model.Front50, model.Orca)
	_, err := runner.Run(ctx, cfg, bom, filter)
	if err == nil {
		t.Fatal("expected cancellation error")
	}

	// Verify state file was saved.
	if _, stateErr := os.Stat(stateFile); os.IsNotExist(stateErr) {
		t.Error("expected state file to be saved on cancellation")
	}
}

func TestDeployStateResume(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "state.json")

	state := DeployState{
		Completed: []string{"front50", "fiat"},
		Remaining: []string{"clouddriver", "orca"},
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadDeployState(stateFile)
	if err != nil {
		t.Fatalf("LoadDeployState: %v", err)
	}

	if len(loaded.Completed) != 2 {
		t.Errorf("completed count = %d, want 2", len(loaded.Completed))
	}
	if len(loaded.Remaining) != 2 {
		t.Errorf("remaining count = %d, want 2", len(loaded.Remaining))
	}
	if loaded.Remaining[0] != "clouddriver" {
		t.Errorf("first remaining = %q, want %q", loaded.Remaining[0], "clouddriver")
	}
}

func TestRemoveDeployState(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "state.json")

	if err := os.WriteFile(stateFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	RemoveDeployState(stateFile)

	if _, err := os.Stat(stateFile); !os.IsNotExist(err) {
		t.Error("expected state file to be removed")
	}
}
