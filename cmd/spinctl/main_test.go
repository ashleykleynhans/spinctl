package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

func TestRootCommandHelp(t *testing.T) {
	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !containsStr(output, "spinctl") {
		t.Error("help should mention spinctl")
	}
}

func TestDeployCommand(t *testing.T) {
	cmd := rootCmd()
	deploy := findSubcommand(cmd, "deploy")
	if deploy == nil {
		t.Fatal("deploy subcommand not found")
	}
}

func TestImportCommand(t *testing.T) {
	cmd := rootCmd()
	imp := findSubcommand(cmd, "import")
	if imp == nil {
		t.Fatal("import subcommand not found")
	}
}

func TestShowCommand(t *testing.T) {
	cmd := rootCmd()
	show := findSubcommand(cmd, "show")
	if show == nil {
		t.Fatal("show subcommand not found")
	}
}

func TestDeployCommandFlags(t *testing.T) {
	cmd := rootCmd()
	deploy := findSubcommand(cmd, "deploy")
	if deploy == nil {
		t.Fatal("deploy subcommand not found")
	}
	servicesFlag := deploy.Flags().Lookup("services")
	if servicesFlag == nil {
		t.Error("deploy should have --services flag")
	}
	dryRunFlag := deploy.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("deploy should have --dry-run flag")
	}
	if dryRunFlag.DefValue != "false" {
		t.Errorf("--dry-run default = %q, want 'false'", dryRunFlag.DefValue)
	}
}

func TestDeployCommandHelp(t *testing.T) {
	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"deploy", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !containsStr(output, "Deploy") {
		t.Error("deploy help should mention 'Deploy'")
	}
}

func TestImportCommandFlags(t *testing.T) {
	cmd := rootCmd()
	imp := findSubcommand(cmd, "import")
	if imp == nil {
		t.Fatal("import subcommand not found")
	}
	halDirFlag := imp.Flags().Lookup("hal-dir")
	if halDirFlag == nil {
		t.Error("import should have --hal-dir flag")
	}
}

func TestImportCommandHelp(t *testing.T) {
	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"import", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !containsStr(output, "Import") || !containsStr(output, "Halyard") {
		t.Error("import help should mention 'Import' and 'Halyard'")
	}
}

func TestShowCommandHelp(t *testing.T) {
	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"show", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !containsStr(output, "Show") {
		t.Error("show help should mention 'Show'")
	}
}

func TestDeployCommandInvalidService(t *testing.T) {
	// Override HOME so loadOrCreateConfig doesn't touch real config.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := rootCmd()
	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)
	cmd.SetOut(&errBuf)
	cmd.SetArgs([]string{"deploy", "--services", "nonexistentservice"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid service name")
	}
	if err != nil && !strings.Contains(err.Error(), "unknown service") {
		t.Errorf("expected 'unknown service' in error, got: %v", err)
	}
}

func TestDeployCommandDryRunValidationFails(t *testing.T) {
	// Override HOME so loadOrCreateConfig creates a default config (no version).
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := rootCmd()
	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)
	cmd.SetOut(&errBuf)
	cmd.SetArgs([]string{"deploy", "--dry-run"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected validation failure for default config with no version")
	}
}

func TestShowCommandNoConfig(t *testing.T) {
	// Override HOME to a dir with no config.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := rootCmd()
	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)
	cmd.SetOut(&errBuf)
	cmd.SetArgs([]string{"show"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no config file exists")
	}
}

func TestShowCommandWithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create a spinctl config file.
	configDir := filepath.Join(tmpDir, ".spinctl")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	configPath := filepath.Join(configDir, "config.yaml")
	if err := config.SaveToFile(cfg, configPath); err != nil {
		t.Fatal(err)
	}

	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"show"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("show command error: %v", err)
	}
}

func TestImportCommandNoHalDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := rootCmd()
	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)
	cmd.SetOut(&errBuf)
	cmd.SetArgs([]string{"import"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no .hal directory exists")
	}
	if err != nil && !strings.Contains(err.Error(), ".hal") {
		t.Errorf("expected error mentioning '.hal', got: %v", err)
	}
}

func TestImportCommandWithHalDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create a fake .hal directory with config.
	halDir := filepath.Join(tmpDir, "myhal")
	if err := os.MkdirAll(halDir, 0700); err != nil {
		t.Fatal(err)
	}
	halConfig := `currentDeployment: default
deploymentConfigurations:
  - name: default
    version: "1.35.0"
`
	if err := os.WriteFile(filepath.Join(halDir, "config"), []byte(halConfig), 0600); err != nil {
		t.Fatal(err)
	}

	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"import", "--hal-dir", halDir})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("import command error: %v", err)
	}
}

func TestRootCommandSubcommands(t *testing.T) {
	cmd := rootCmd()
	expected := map[string]bool{"deploy": false, "import": false, "show": false}
	for _, c := range cmd.Commands() {
		if _, ok := expected[c.Name()]; ok {
			expected[c.Name()] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

func TestRootCommandDescription(t *testing.T) {
	cmd := rootCmd()
	if cmd.Short == "" {
		t.Error("root command should have a short description")
	}
	if cmd.Long == "" {
		t.Error("root command should have a long description")
	}
}

func TestDeployCommandMultipleServices(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := rootCmd()
	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)
	cmd.SetOut(&errBuf)
	// Valid service names but no BOM version set, so validation will fail.
	cmd.SetArgs([]string{"deploy", "--services", "gate,orca", "--dry-run"})
	err := cmd.Execute()
	// Should fail validation (no version in default config).
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestDeployCommandValidationErrorOutput(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"deploy"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected validation error for default config")
	}
	if err != nil && !strings.Contains(err.Error(), "validation") {
		t.Errorf("expected 'validation' in error, got: %v", err)
	}
}

func TestDeployCommandDryRunWithValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create config with version and an enabled service.
	configDir := filepath.Join(tmpDir, ".spinctl")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	svc := cfg.Services[model.Gate]
	svc.Enabled = true
	cfg.Services[model.Gate] = svc
	configPath := filepath.Join(configDir, "config.yaml")
	if err := config.SaveToFile(cfg, configPath); err != nil {
		t.Fatal(err)
	}

	// Also create cached BOM so we don't hit the network.
	cacheDir := filepath.Join(configDir, "cache", "bom")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}
	bomYAML := `version: "1.35.0"
timestamp: "2025-01-15"
services:
  clouddriver:
    version: "5.82.1"
  orca:
    version: "8.47.0"
  gate:
    version: "6.62.0"
  front50:
    version: "2.33.0"
  echo:
    version: "2.40.0"
  igor:
    version: "4.18.0"
  fiat:
    version: "1.43.0"
  rosco:
    version: "1.20.0"
  kayenta:
    version: "2.40.0"
  deck:
    version: "3.16.0"
`
	if err := os.WriteFile(filepath.Join(cacheDir, "1.35.0.yml"), []byte(bomYAML), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"deploy", "--dry-run"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("deploy --dry-run error: %v", err)
	}
}

func TestDeployCommandDryRunFilteredServices(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create config with version and an enabled service.
	configDir := filepath.Join(tmpDir, ".spinctl")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	svc := cfg.Services[model.Gate]
	svc.Enabled = true
	cfg.Services[model.Gate] = svc
	configPath := filepath.Join(configDir, "config.yaml")
	if err := config.SaveToFile(cfg, configPath); err != nil {
		t.Fatal(err)
	}

	// Create cached BOM.
	cacheDir := filepath.Join(configDir, "cache", "bom")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}
	bomYAML := `version: "1.35.0"
timestamp: "2025-01-15"
services:
  clouddriver:
    version: "5.82.1"
  orca:
    version: "8.47.0"
  gate:
    version: "6.62.0"
  front50:
    version: "2.33.0"
  echo:
    version: "2.40.0"
  igor:
    version: "4.18.0"
  fiat:
    version: "1.43.0"
  rosco:
    version: "1.20.0"
  kayenta:
    version: "2.40.0"
  deck:
    version: "3.16.0"
`
	if err := os.WriteFile(filepath.Join(cacheDir, "1.35.0.yml"), []byte(bomYAML), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"deploy", "--services", "gate", "--dry-run"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("deploy --dry-run --services gate error: %v", err)
	}
}

func TestLoadOrCreateConfigDefault(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg, path := loadOrCreateConfig()
	if cfg == nil {
		t.Fatal("loadOrCreateConfig() returned nil config")
	}
	if path == "" {
		t.Error("loadOrCreateConfig() returned empty path")
	}
	// Default config should have 10 services.
	if len(cfg.Services) != 10 {
		t.Errorf("expected 10 services, got %d", len(cfg.Services))
	}
}

func TestDeployCommandDryRunWithWarnings(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".spinctl")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	svc := cfg.Services[model.Gate]
	svc.Enabled = true
	cfg.Services[model.Gate] = svc
	configPath := filepath.Join(configDir, "config.yaml")
	if err := config.SaveToFile(cfg, configPath); err != nil {
		t.Fatal(err)
	}

	cacheDir := filepath.Join(configDir, "cache", "bom")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}
	bomYAML := `version: "1.35.0"
timestamp: "2025-01-15"
services:
  clouddriver:
    version: "5.82.1"
  orca:
    version: "8.47.0"
  gate:
    version: "6.62.0"
  front50:
    version: "2.33.0"
  echo:
    version: "2.40.0"
  igor:
    version: "4.18.0"
  fiat:
    version: "1.43.0"
  rosco:
    version: "1.20.0"
  kayenta:
    version: "2.40.0"
  deck:
    version: "3.16.0"
`
	if err := os.WriteFile(filepath.Join(cacheDir, "1.35.0.yml"), []byte(bomYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Use gate only, which has dependencies, triggering warnings.
	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"deploy", "--services", "gate", "--dry-run"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("deploy --dry-run --services gate error: %v", err)
	}
}

func TestImportCommandWithUnmappedFields(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	halDir := filepath.Join(tmpDir, "myhal")
	if err := os.MkdirAll(halDir, 0700); err != nil {
		t.Fatal(err)
	}
	halConfig := `currentDeployment: default
deploymentConfigurations:
  - name: default
    version: "1.35.0"
    notifications:
      slack:
        enabled: true
    metricStores:
      datadog:
        enabled: false
`
	if err := os.WriteFile(filepath.Join(halDir, "config"), []byte(halConfig), 0600); err != nil {
		t.Fatal(err)
	}

	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"import", "--hal-dir", halDir})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("import command error: %v", err)
	}
}

func TestLoadOrCreateConfigExisting(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".spinctl")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	cfg := config.NewDefault()
	cfg.Version = "1.35.0"
	configPath := filepath.Join(configDir, "config.yaml")
	if err := config.SaveToFile(cfg, configPath); err != nil {
		t.Fatal(err)
	}

	loaded, path := loadOrCreateConfig()
	if loaded.Version != "1.35.0" {
		t.Errorf("version = %q, want '1.35.0'", loaded.Version)
	}
	if path == "" {
		t.Error("path should not be empty")
	}
}

func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, c := range cmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
