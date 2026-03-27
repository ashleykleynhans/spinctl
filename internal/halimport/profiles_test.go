package halimport

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

func TestImportServiceSettings(t *testing.T) {
	cfg := config.NewDefault()
	for _, name := range model.AllServiceNames() {
		svc := cfg.Services[name]
		svc.Enabled = true
		cfg.Services[name] = svc
	}

	dir := t.TempDir()
	// Write a gate.yml service-settings file.
	gateSettings := `host: 0.0.0.0
port: 9090
enabled: true
healthEndpoint: /health
env:
  JAVA_OPTS: "-Xmx512m"
`
	os.WriteFile(filepath.Join(dir, "gate.yml"), []byte(gateSettings), 0644)

	importServiceSettings(cfg, dir)

	gate := cfg.Services[model.Gate]
	if gate.Host != "0.0.0.0" {
		t.Errorf("gate host = %q, want %q", gate.Host, "0.0.0.0")
	}
	if gate.Port != 9090 {
		t.Errorf("gate port = %d, want 9090", gate.Port)
	}
	// healthEndpoint and env should be in Settings.
	if gate.Settings.Kind == 0 {
		t.Error("gate Settings should not be empty")
	}
}

func TestImportServiceSettingsDisablesService(t *testing.T) {
	cfg := config.NewDefault()
	for _, name := range model.AllServiceNames() {
		svc := cfg.Services[name]
		svc.Enabled = true
		cfg.Services[name] = svc
	}

	dir := t.TempDir()
	kayentaSettings := `enabled: false
`
	os.WriteFile(filepath.Join(dir, "kayenta.yml"), []byte(kayentaSettings), 0644)

	importServiceSettings(cfg, dir)

	kayenta := cfg.Services[model.Kayenta]
	if kayenta.Enabled {
		t.Error("kayenta should be disabled after importing service-settings")
	}
}

func TestImportServiceSettingsMissingDir(t *testing.T) {
	cfg := config.NewDefault()
	// Should not panic on missing directory.
	importServiceSettings(cfg, "/nonexistent/path")
}

func TestImportServiceSettingsUnknownService(t *testing.T) {
	cfg := config.NewDefault()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "unknown-service.yml"), []byte("port: 1234"), 0644)

	importServiceSettings(cfg, dir)
	// Should not add unknown services.
	if len(cfg.Services) != 10 {
		t.Errorf("expected 10 services, got %d", len(cfg.Services))
	}
}

func TestImportProfiles(t *testing.T) {
	cfg := config.NewDefault()
	for _, name := range model.AllServiceNames() {
		svc := cfg.Services[name]
		svc.Enabled = true
		cfg.Services[name] = svc
	}

	dir := t.TempDir()
	// Write a gate-local.yml profile.
	gateProfile := `server:
  ssl:
    enabled: true
    keyStore: /path/to/keystore
cors:
  allowedOrigins: "*"
`
	os.WriteFile(filepath.Join(dir, "gate-local.yml"), []byte(gateProfile), 0644)

	importProfiles(cfg, dir)

	gate := cfg.Services[model.Gate]
	if gate.Settings.Kind == 0 {
		t.Error("gate Settings should contain profile data")
	}
}

func TestImportProfilesWithoutLocalSuffix(t *testing.T) {
	cfg := config.NewDefault()
	for _, name := range model.AllServiceNames() {
		svc := cfg.Services[name]
		svc.Enabled = true
		cfg.Services[name] = svc
	}

	dir := t.TempDir()
	// Write orca.yml (without -local suffix).
	orcaProfile := `tasks:
  executionWindow:
    timezone: America/Los_Angeles
`
	os.WriteFile(filepath.Join(dir, "orca.yml"), []byte(orcaProfile), 0644)

	importProfiles(cfg, dir)

	orca := cfg.Services[model.Orca]
	if orca.Settings.Kind == 0 {
		t.Error("orca Settings should contain profile data")
	}
}

func TestImportProfilesMissingDir(t *testing.T) {
	cfg := config.NewDefault()
	importProfiles(cfg, "/nonexistent/path")
}

func TestImportProfilesMergesWithExisting(t *testing.T) {
	cfg := config.NewDefault()
	for _, name := range model.AllServiceNames() {
		svc := cfg.Services[name]
		svc.Enabled = true
		cfg.Services[name] = svc
	}

	dir := t.TempDir()

	// First import creates settings.
	os.WriteFile(filepath.Join(dir, "gate-local.yml"), []byte("server:\n  port: 8084"), 0644)
	importProfiles(cfg, dir)

	// Second import with different keys should merge.
	dir2 := t.TempDir()
	os.WriteFile(filepath.Join(dir2, "gate-local.yml"), []byte("cors:\n  enabled: true"), 0644)
	importProfiles(cfg, dir2)

	gate := cfg.Services[model.Gate]
	if gate.Settings.Kind == 0 {
		t.Error("gate Settings should have merged data")
	}
}

func TestFullImportWithProfilesAndServiceSettings(t *testing.T) {
	// Create a fake .hal directory with all components.
	halDir := t.TempDir()

	// Main config.
	halConfig := `halyardVersion: "1.56.0"
currentDeployment: default
deploymentConfigurations:
  - name: default
    version: "1.35.0"
    providers:
      kubernetes:
        enabled: true
        accounts:
          - name: prod
            context: prod-context
`
	os.WriteFile(filepath.Join(halDir, "config"), []byte(halConfig), 0644)

	// Service settings.
	ssDir := filepath.Join(halDir, "default", "service-settings")
	os.MkdirAll(ssDir, 0755)
	os.WriteFile(filepath.Join(ssDir, "gate.yml"), []byte("host: 0.0.0.0\nport: 9090"), 0644)
	os.WriteFile(filepath.Join(ssDir, "kayenta.yml"), []byte("enabled: false"), 0644)

	// Profiles.
	profDir := filepath.Join(halDir, "default", "profiles")
	os.MkdirAll(profDir, 0755)
	os.WriteFile(filepath.Join(profDir, "gate-local.yml"), []byte("server:\n  ssl:\n    enabled: true"), 0644)

	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "config.yaml")

	result, err := Import(ImportOptions{
		HalDir:         halDir,
		DeploymentName: "default",
		OutputPath:     outPath,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	cfg := result.Config

	// Check gate got service-settings overrides.
	gate := cfg.Services[model.Gate]
	if gate.Host != "0.0.0.0" {
		t.Errorf("gate host = %q, want %q", gate.Host, "0.0.0.0")
	}
	if gate.Port != 9090 {
		t.Errorf("gate port = %d, want 9090", gate.Port)
	}

	// Check gate got profile settings.
	if gate.Settings.Kind == 0 {
		t.Error("gate Settings should contain profile data")
	}

	// Check kayenta is disabled.
	kayenta := cfg.Services[model.Kayenta]
	if kayenta.Enabled {
		t.Error("kayenta should be disabled")
	}
}
