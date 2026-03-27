package halimport

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spinnaker/spinctl/internal/model"
)

func testdataPath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func TestMapBasicHalConfig(t *testing.T) {
	hal, err := parseHalFile(testdataPath("basic_hal_config.yaml"))
	if err != nil {
		t.Fatalf("parseHalFile: %v", err)
	}

	cfg, err := mapHalToSpinctl(hal, "default")
	if err != nil {
		t.Fatalf("mapHalToSpinctl: %v", err)
	}

	// Check version.
	if cfg.Version != "1.35.0" {
		t.Errorf("version = %q, want %q", cfg.Version, "1.35.0")
	}

	// Check kubernetes provider enabled with account "prod".
	k8s, ok := cfg.Providers["kubernetes"]
	if !ok {
		t.Fatal("kubernetes provider not found")
	}
	if !k8s.Enabled {
		t.Error("kubernetes provider should be enabled")
	}
	if len(k8s.Accounts) != 1 || k8s.Accounts[0].Name != "prod" {
		t.Errorf("kubernetes accounts = %v, want [prod]", k8s.Accounts)
	}

	// Check aws provider disabled.
	aws, ok := cfg.Providers["aws"]
	if !ok {
		t.Fatal("aws provider not found")
	}
	if aws.Enabled {
		t.Error("aws provider should be disabled")
	}

	// Check features.
	if !cfg.Features["artifacts"] {
		t.Error("features.artifacts should be true")
	}

	// Check unmapped globals in Custom (notifications, metricStores).
	if cfg.Custom == nil {
		t.Fatal("Custom map should not be nil")
	}
	if _, ok := cfg.Custom["notifications"]; !ok {
		t.Error("Custom should contain notifications")
	}
	if _, ok := cfg.Custom["metricStores"]; !ok {
		t.Error("Custom should contain metricStores")
	}
	if _, ok := cfg.Custom["deploymentEnvironment"]; !ok {
		t.Error("Custom should contain deploymentEnvironment")
	}
}

func TestMapMultiDeployment(t *testing.T) {
	hal, err := parseHalFile(testdataPath("multi_deployment_hal_config.yaml"))
	if err != nil {
		t.Fatalf("parseHalFile: %v", err)
	}

	deployments := listDeployments(hal)
	if len(deployments) != 2 {
		t.Fatalf("listDeployments = %d, want 2", len(deployments))
	}

	cfg, err := mapHalToSpinctl(hal, "staging")
	if err != nil {
		t.Fatalf("mapHalToSpinctl: %v", err)
	}

	if cfg.Version != "1.35.0" {
		t.Errorf("version = %q, want %q", cfg.Version, "1.35.0")
	}

	k8s, ok := cfg.Providers["kubernetes"]
	if !ok {
		t.Fatal("kubernetes provider not found")
	}
	if len(k8s.Accounts) != 1 || k8s.Accounts[0].Name != "staging" {
		t.Errorf("kubernetes accounts = %v, want [staging]", k8s.Accounts)
	}
}

func TestMapNonexistentDeployment(t *testing.T) {
	hal, err := parseHalFile(testdataPath("basic_hal_config.yaml"))
	if err != nil {
		t.Fatalf("parseHalFile: %v", err)
	}

	_, err = mapHalToSpinctl(hal, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent deployment")
	}
}

func TestMapMinimalConfig(t *testing.T) {
	hal, err := parseHalFile(testdataPath("minimal_hal_config.yaml"))
	if err != nil {
		t.Fatalf("parseHalFile: %v", err)
	}

	cfg, err := mapHalToSpinctl(hal, "default")
	if err != nil {
		t.Fatalf("mapHalToSpinctl: %v", err)
	}

	if cfg.Version != "1.35.0" {
		t.Errorf("version = %q, want %q", cfg.Version, "1.35.0")
	}

	// Should have 10 default services.
	if len(cfg.Services) != 10 {
		t.Errorf("services count = %d, want 10", len(cfg.Services))
	}

	// Gate should have port 8084.
	gate, ok := cfg.Services[model.Gate]
	if !ok {
		t.Fatal("gate service not found")
	}
	if gate.Port != 8084 {
		t.Errorf("gate port = %d, want 8084", gate.Port)
	}
}
