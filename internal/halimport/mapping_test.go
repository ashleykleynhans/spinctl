package halimport

import (
	"os"
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

	// Check known halyard sections mapped to dedicated fields.
	if cfg.Notifications == nil {
		t.Error("Notifications should not be nil")
	}
	if cfg.MetricStores == nil {
		t.Error("MetricStores should not be nil")
	}
	if cfg.DeploymentEnvironment == nil {
		t.Error("DeploymentEnvironment should not be nil")
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

func TestMapSecurityNilFields(t *testing.T) {
	sec := mapSecurity(&halSecurity{})
	if sec.Authn.Enabled {
		t.Error("authn should default to disabled when nil")
	}
	if sec.Authz.Enabled {
		t.Error("authz should default to disabled when nil")
	}
}

func TestMapSecurityAuthnOnly(t *testing.T) {
	sec := mapSecurity(&halSecurity{
		Authn: &halAuthToggle{Enabled: true},
	})
	if !sec.Authn.Enabled {
		t.Error("authn should be enabled")
	}
	if sec.Authz.Enabled {
		t.Error("authz should be disabled")
	}
}

func TestMapSecurityAuthzOnly(t *testing.T) {
	sec := mapSecurity(&halSecurity{
		Authz: &halAuthToggle{Enabled: true},
	})
	if sec.Authn.Enabled {
		t.Error("authn should be disabled")
	}
	if !sec.Authz.Enabled {
		t.Error("authz should be enabled")
	}
}

func TestMapProvidersEmpty(t *testing.T) {
	result := mapProviders(map[string]halProvider{})
	if len(result) != 0 {
		t.Errorf("expected empty providers map, got %d", len(result))
	}
}

func TestMapProvidersWithAccountExtras(t *testing.T) {
	providers := map[string]halProvider{
		"kubernetes": {
			Enabled: true,
			Accounts: []halAccount{
				{
					Name:    "test",
					Context: "ctx",
					Extra:   map[string]any{"requiredGroupMembership": []any{"group1"}},
				},
			},
		},
	}
	result := mapProviders(providers)
	k8s := result["kubernetes"]
	if len(k8s.Accounts) != 1 {
		t.Fatalf("accounts count = %d, want 1", len(k8s.Accounts))
	}
	if k8s.Accounts[0].Extra == nil {
		t.Error("account extra should not be nil")
	}
}

func TestMapProvidersWithNoAccounts(t *testing.T) {
	providers := map[string]halProvider{
		"aws": {
			Enabled:  false,
			Accounts: nil,
		},
	}
	result := mapProviders(providers)
	aws := result["aws"]
	if aws.Enabled {
		t.Error("aws should be disabled")
	}
	if len(aws.Accounts) != 0 {
		t.Errorf("accounts count = %d, want 0", len(aws.Accounts))
	}
}

func TestListDeployments(t *testing.T) {
	hal := &halConfig{
		DeploymentConfigurations: []deploymentConfig{
			{Name: "default"},
			{Name: "staging"},
		},
	}
	names := listDeployments(hal)
	if len(names) != 2 {
		t.Fatalf("expected 2 deployments, got %d", len(names))
	}
	if names[0] != "default" || names[1] != "staging" {
		t.Errorf("deployments = %v, want [default, staging]", names)
	}
}

func TestFindDeploymentNotFound(t *testing.T) {
	hal := &halConfig{
		DeploymentConfigurations: []deploymentConfig{
			{Name: "default"},
		},
	}
	_, err := findDeployment(hal, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent deployment")
	}
}

func TestMapHalToSpinctlNoProviders(t *testing.T) {
	hal := &halConfig{
		DeploymentConfigurations: []deploymentConfig{
			{
				Name:    "default",
				Version: "1.35.0",
			},
		},
	}
	cfg, err := mapHalToSpinctl(hal, "default")
	if err != nil {
		t.Fatalf("mapHalToSpinctl: %v", err)
	}
	if cfg.Providers != nil {
		t.Error("providers should be nil when none configured")
	}
}

func TestMapHalToSpinctlNoSecurity(t *testing.T) {
	hal := &halConfig{
		DeploymentConfigurations: []deploymentConfig{
			{
				Name:    "default",
				Version: "1.35.0",
			},
		},
	}
	cfg, err := mapHalToSpinctl(hal, "default")
	if err != nil {
		t.Fatalf("mapHalToSpinctl: %v", err)
	}
	if cfg.Security.Authn.Enabled || cfg.Security.Authz.Enabled {
		t.Error("security should be disabled when not configured")
	}
}

func TestMapHalToSpinctlNoFeatures(t *testing.T) {
	hal := &halConfig{
		DeploymentConfigurations: []deploymentConfig{
			{
				Name:    "default",
				Version: "1.35.0",
			},
		},
	}
	cfg, err := mapHalToSpinctl(hal, "default")
	if err != nil {
		t.Fatalf("mapHalToSpinctl: %v", err)
	}
	// Features should remain the default (empty map from NewDefault).
	if len(cfg.Features) != 0 {
		t.Errorf("features should be empty, got %v", cfg.Features)
	}
}

func TestMapHalToSpinctlNoExtras(t *testing.T) {
	hal := &halConfig{
		DeploymentConfigurations: []deploymentConfig{
			{
				Name:    "default",
				Version: "1.35.0",
			},
		},
	}
	cfg, err := mapHalToSpinctl(hal, "default")
	if err != nil {
		t.Fatalf("mapHalToSpinctl: %v", err)
	}
	if cfg.Custom != nil {
		t.Error("custom should be nil when no extra fields")
	}
}

func TestParseHalFileInvalid(t *testing.T) {
	_, err := parseHalFile("/nonexistent/path/config")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseHalFileInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config")
	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := parseHalFile(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestMapHalToSpinctlWithAllSections(t *testing.T) {
	hal := &halConfig{
		DeploymentConfigurations: []deploymentConfig{
			{
				Name:    "default",
				Version: "1.35.0",
				Extra: map[string]any{
					"artifacts":             map[string]any{"github": map[string]any{"enabled": true}},
					"persistentStorage":     map[string]any{"s3": map[string]any{"bucket": "spin"}},
					"notifications":         map[string]any{"slack": map[string]any{"enabled": true}},
					"ci":                    map[string]any{"jenkins": map[string]any{"enabled": true}},
					"canary":                map[string]any{"enabled": true, "serviceIntegrations": []any{}},
					"metricStores":          map[string]any{"prometheus": map[string]any{"enabled": false}},
					"timezone":              "America/Los_Angeles",
					"deploymentEnvironment": map[string]any{"type": "Distributed"},
					"spinnaker":             map[string]any{"extensibility": map[string]any{}},
					"repository":            map[string]any{"artifactory": map[string]any{"enabled": false}},
					"pubsub":                map[string]any{"google": map[string]any{"enabled": false}},
					"webhook":               map[string]any{"trust": map[string]any{"enabled": false}},
					"stats":                 map[string]any{"enabled": true},
					"customField":           "customValue",
				},
			},
		},
	}

	cfg, err := mapHalToSpinctl(hal, "default")
	if err != nil {
		t.Fatalf("mapHalToSpinctl: %v", err)
	}

	if cfg.Artifacts == nil {
		t.Error("Artifacts should not be nil")
	}
	if cfg.PersistentStorage == nil {
		t.Error("PersistentStorage should not be nil")
	}
	if cfg.Notifications == nil {
		t.Error("Notifications should not be nil")
	}
	if cfg.CI == nil {
		t.Error("CI should not be nil")
	}
	if cfg.Canary == nil {
		t.Error("Canary should not be nil")
	}
	if cfg.MetricStores == nil {
		t.Error("MetricStores should not be nil")
	}
	if cfg.DeploymentEnvironment == nil {
		t.Error("DeploymentEnvironment should not be nil")
	}
	if cfg.Spinnaker == nil {
		t.Error("Spinnaker should not be nil")
	}
	if cfg.Repository == nil {
		t.Error("Repository should not be nil")
	}
	if cfg.Pubsub == nil {
		t.Error("Pubsub should not be nil")
	}
	if cfg.Webhook == nil {
		t.Error("Webhook should not be nil")
	}
	if cfg.Stats == nil {
		t.Error("Stats should not be nil")
	}
	if cfg.Timezone != "America/Los_Angeles" {
		t.Errorf("Timezone = %q, want %q", cfg.Timezone, "America/Los_Angeles")
	}
	if cfg.Custom == nil || cfg.Custom["customField"] != "customValue" {
		t.Error("Custom should contain unmapped fields")
	}
}

func TestMapHalToSpinctlTimezone(t *testing.T) {
	hal := &halConfig{
		DeploymentConfigurations: []deploymentConfig{
			{
				Name:    "default",
				Version: "1.35.0",
				Extra: map[string]any{
					"timezone": "Europe/London",
				},
			},
		},
	}

	cfg, err := mapHalToSpinctl(hal, "default")
	if err != nil {
		t.Fatalf("mapHalToSpinctl: %v", err)
	}

	if cfg.Timezone != "Europe/London" {
		t.Errorf("Timezone = %q, want %q", cfg.Timezone, "Europe/London")
	}
}

func TestToMapStringAnyNonMap(t *testing.T) {
	// Pass a string -- should return nil.
	result := toMapStringAny("not a map")
	if result != nil {
		t.Errorf("toMapStringAny(string) = %v, want nil", result)
	}

	// Pass an int -- should return nil.
	result = toMapStringAny(42)
	if result != nil {
		t.Errorf("toMapStringAny(int) = %v, want nil", result)
	}

	// Pass nil -- should return nil.
	result = toMapStringAny(nil)
	if result != nil {
		t.Errorf("toMapStringAny(nil) = %v, want nil", result)
	}
}

func TestMinimalConfig(t *testing.T) {
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
