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

func TestMergeIntoSettingsEmpty(t *testing.T) {
	svc := config.ServiceConfig{
		Enabled: true,
		Host:    "localhost",
		Port:    8080,
	}

	settings := map[string]any{
		"server": map[string]any{"port": 9090},
	}
	mergeIntoSettings(&svc, settings)

	if svc.Settings.Kind == 0 {
		t.Error("Settings should not be empty after merge")
	}
}

func TestMergeIntoSettingsExisting(t *testing.T) {
	svc := config.ServiceConfig{
		Enabled: true,
		Host:    "localhost",
		Port:    8080,
	}

	// First merge creates settings.
	mergeIntoSettings(&svc, map[string]any{"server": map[string]any{"port": 9090}})

	// Second merge should not overwrite existing key.
	mergeIntoSettings(&svc, map[string]any{
		"server": map[string]any{"port": 7070}, // should not overwrite
		"cors":   map[string]any{"enabled": true},
	})

	// Should have both "server" and "cors" keys.
	keyCount := len(svc.Settings.Content) / 2
	if keyCount != 2 {
		t.Errorf("expected 2 keys in settings, got %d", keyCount)
	}
}

func TestImportServiceSettingsInvalidYAML(t *testing.T) {
	cfg := config.NewDefault()
	dir := t.TempDir()
	// Write invalid YAML.
	os.WriteFile(filepath.Join(dir, "gate.yml"), []byte("{{invalid yaml"), 0644)

	// Should not panic, just skip invalid file.
	importServiceSettings(cfg, dir)
	gate := cfg.Services[model.Gate]
	if gate.Host != "localhost" {
		t.Error("gate host should remain default after invalid YAML")
	}
}

func TestImportProfilesSkipsDirectories(t *testing.T) {
	cfg := config.NewDefault()
	dir := t.TempDir()

	// Create a subdirectory that looks like a service.
	os.MkdirAll(filepath.Join(dir, "gate.yml"), 0755)

	// Should not panic.
	importProfiles(cfg, dir)
}

func TestImportProfilesInvalidYAML(t *testing.T) {
	cfg := config.NewDefault()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "gate-local.yml"), []byte("{{not valid yaml"), 0644)

	// Should not panic, just skip invalid file.
	importProfiles(cfg, dir)
	gate := cfg.Services[model.Gate]
	if gate.Settings.Kind != 0 {
		t.Error("gate Settings should be empty after invalid YAML")
	}
}

func TestImportServiceSettingsSkipsNonYML(t *testing.T) {
	cfg := config.NewDefault()
	dir := t.TempDir()
	// Write a .txt file that should be skipped.
	os.WriteFile(filepath.Join(dir, "gate.txt"), []byte("host: 0.0.0.0"), 0644)

	importServiceSettings(cfg, dir)
	gate := cfg.Services[model.Gate]
	if gate.Host != "localhost" {
		t.Error("non-.yml files should be skipped")
	}
}

func TestImportServiceSettingsSkipsDirectories(t *testing.T) {
	cfg := config.NewDefault()
	dir := t.TempDir()
	// Create a directory named gate.yml.
	os.MkdirAll(filepath.Join(dir, "gate.yml"), 0755)

	importServiceSettings(cfg, dir)
	gate := cfg.Services[model.Gate]
	if gate.Host != "localhost" {
		t.Error("directories should be skipped even with .yml name")
	}
}

func TestImportProfilesUnknownService(t *testing.T) {
	cfg := config.NewDefault()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "unknown-local.yml"), []byte("key: val"), 0644)

	importProfiles(cfg, dir)
	// Should not crash or add anything.
	if len(cfg.Services) != 10 {
		t.Errorf("expected 10 services, got %d", len(cfg.Services))
	}
}

func TestImportProfilesEmptyFile(t *testing.T) {
	cfg := config.NewDefault()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "gate-local.yml"), []byte(""), 0644)

	importProfiles(cfg, dir)
	gate := cfg.Services[model.Gate]
	if gate.Settings.Kind != 0 {
		t.Error("empty profile should not create settings")
	}
}

func TestImportProfilesNonYMLSkipped(t *testing.T) {
	cfg := config.NewDefault()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "gate.txt"), []byte("key: val"), 0644)

	importProfiles(cfg, dir)
	gate := cfg.Services[model.Gate]
	if gate.Settings.Kind != 0 {
		t.Error("non-.yml profiles should be skipped")
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

func TestImportProfilesOverridesFile(t *testing.T) {
	cfg := config.NewDefault()
	for _, name := range model.AllServiceNames() {
		svc := cfg.Services[name]
		svc.Enabled = true
		cfg.Services[name] = svc
	}

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "orca-overrides.yml"), []byte("tasks:\n  timeout: 300"), 0644)

	importProfiles(cfg, dir)

	orca := cfg.Services[model.Orca]
	if orca.Settings.Kind == 0 {
		t.Error("orca Settings should contain overrides data")
	}
}

func TestImportProfilesBaseYml(t *testing.T) {
	cfg := config.NewDefault()
	for _, name := range model.AllServiceNames() {
		svc := cfg.Services[name]
		svc.Enabled = true
		cfg.Services[name] = svc
	}

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "orca.yml"), []byte("server:\n  port: 8083"), 0644)

	importProfiles(cfg, dir)

	orca := cfg.Services[model.Orca]
	if orca.Settings.Kind == 0 {
		t.Error("orca Settings should contain profile data from orca.yml")
	}
}

func TestImportProfilesSettingsLocalJS(t *testing.T) {
	cfg := config.NewDefault()

	dir := t.TempDir()
	jsContent := "window.spinnakerSettings.feature.canary = true;"
	os.WriteFile(filepath.Join(dir, "settings-local.js"), []byte(jsContent), 0644)

	importProfiles(cfg, dir)

	if cfg.ProfileFiles == nil {
		t.Fatal("ProfileFiles should not be nil")
	}
	if cfg.ProfileFiles["settings-local.js"] != jsContent {
		t.Errorf("ProfileFiles[settings-local.js] = %q, want %q", cfg.ProfileFiles["settings-local.js"], jsContent)
	}
}

func TestImportProfilesNonYAMLFile(t *testing.T) {
	cfg := config.NewDefault()

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "custom-script.sh"), []byte("#!/bin/bash\necho hello"), 0644)

	importProfiles(cfg, dir)

	if cfg.ProfileFiles == nil {
		t.Fatal("ProfileFiles should not be nil")
	}
	if _, ok := cfg.ProfileFiles["custom-script.sh"]; !ok {
		t.Error("should store non-YAML files in ProfileFiles")
	}
}

func TestExtractServiceName(t *testing.T) {
	tests := []struct {
		filename string
		want     string
		ok       bool
	}{
		{"gate.yml", "gate", true},
		{"gate-local.yml", "gate", true},
		{"orca-overrides.yml", "orca", true},
		{"orca-override.yml", "orca", true},
		{"clouddriver-custom.yml", "clouddriver", true},
		{"unknown-local.yml", "", false},
		{"settings-local.js", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			svc, ok := extractServiceName(tt.filename)
			if ok != tt.ok {
				t.Errorf("extractServiceName(%q) ok = %v, want %v", tt.filename, ok, tt.ok)
				return
			}
			if ok && svc.String() != tt.want {
				t.Errorf("extractServiceName(%q) = %q, want %q", tt.filename, svc, tt.want)
			}
		})
	}
}
