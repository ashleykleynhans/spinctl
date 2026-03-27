package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spinnaker/spinctl/internal/model"
)

func TestLoadValidConfig(t *testing.T) {
	cfg, err := LoadFromFile("testdata/valid_config.yaml")
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}
	if cfg.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", cfg.SchemaVersion)
	}
	if cfg.Version != "1.35.0" {
		t.Errorf("Version = %q, want %q", cfg.Version, "1.35.0")
	}
	if cfg.AptRepository != "https://us-apt.pkg.dev/projects/spinnaker-community" {
		t.Errorf("AptRepository = %q", cfg.AptRepository)
	}
	if v, ok := cfg.ServiceOverrides[model.Clouddriver]; !ok || v != "5.82.1" {
		t.Errorf("ServiceOverrides[clouddriver] = %q, want %q", v, "5.82.1")
	}
	gate, ok := cfg.Services[model.Gate]
	if !ok {
		t.Fatal("gate service not found")
	}
	if !gate.Enabled || gate.Port != 8084 {
		t.Errorf("gate = enabled:%v port:%d", gate.Enabled, gate.Port)
	}
	if !cfg.Features["artifacts"] {
		t.Error("features.artifacts should be true")
	}
	if cfg.Custom == nil {
		t.Error("Custom should not be nil")
	}
}

func TestLoadMinimalConfig(t *testing.T) {
	cfg, err := LoadFromFile("testdata/minimal_config.yaml")
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}
	if cfg.Version != "1.35.0" {
		t.Errorf("Version = %q", cfg.Version)
	}
	if len(cfg.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(cfg.Services))
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := LoadFromFile("testdata/nonexistent.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestSaveAndReload(t *testing.T) {
	original, err := LoadFromFile("testdata/valid_config.yaml")
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	if err := SaveToFile(original, tmpFile); err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
	reloaded, err := LoadFromFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadFromFile(reloaded) error = %v", err)
	}
	if reloaded.Version != original.Version {
		t.Errorf("Version mismatch: %q vs %q", reloaded.Version, original.Version)
	}
	if len(reloaded.Services) != len(original.Services) {
		t.Errorf("Services count mismatch: %d vs %d", len(reloaded.Services), len(original.Services))
	}
}

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefault()
	if cfg.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", cfg.SchemaVersion)
	}
	if len(cfg.Services) != 10 {
		t.Errorf("expected 10 services, got %d", len(cfg.Services))
	}
	gate := cfg.Services[model.Gate]
	if gate.Port != 8084 {
		t.Errorf("gate default port = %d, want 8084", gate.Port)
	}
}

func TestLoadFromBytesInvalidYAML(t *testing.T) {
	_, err := LoadFromBytes([]byte("{{{{invalid yaml"))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadFromBytesSchemaVersionZeroDefaultsToOne(t *testing.T) {
	yamlData := []byte("version: '1.35.0'\nservices: {}\n")
	cfg, err := LoadFromBytes(yamlData)
	if err != nil {
		t.Fatalf("LoadFromBytes() error = %v", err)
	}
	if cfg.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", cfg.SchemaVersion, CurrentSchemaVersion)
	}
}

func TestLoadFromBytesPreservesSchemaVersion(t *testing.T) {
	yamlData := []byte("schema_version: 1\nversion: '1.35.0'\n")
	cfg, err := LoadFromBytes(yamlData)
	if err != nil {
		t.Fatalf("LoadFromBytes() error = %v", err)
	}
	if cfg.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", cfg.SchemaVersion)
	}
}

func TestLoadFromBytesEmptyData(t *testing.T) {
	cfg, err := LoadFromBytes([]byte(""))
	if err != nil {
		t.Fatalf("LoadFromBytes() error = %v", err)
	}
	// Empty YAML should produce a zero-value config with schema defaulted to 1.
	if cfg.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", cfg.SchemaVersion, CurrentSchemaVersion)
	}
}

func TestDefaultConfigDir(t *testing.T) {
	dir := DefaultConfigDir()
	if dir == "" {
		t.Error("DefaultConfigDir() returned empty string")
	}
	// Should end with .spinctl.
	base := filepath.Base(dir)
	if base != ".spinctl" {
		t.Errorf("DefaultConfigDir() base = %q, want '.spinctl'", base)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path == "" {
		t.Error("DefaultConfigPath() returned empty string")
	}
	base := filepath.Base(path)
	if base != "config.yaml" {
		t.Errorf("DefaultConfigPath() base = %q, want 'config.yaml'", base)
	}
	// Should be within DefaultConfigDir.
	dir := filepath.Dir(path)
	expected := DefaultConfigDir()
	if dir != expected {
		t.Errorf("DefaultConfigPath() dir = %q, want %q", dir, expected)
	}
}

func TestDefaultPort(t *testing.T) {
	tests := []struct {
		name model.ServiceName
		want int
	}{
		{model.Gate, 8084},
		{model.Clouddriver, 7002},
		{model.Deck, 9000},
		{model.Orca, 8083},
		{model.Front50, 8080},
		{model.Echo, 8089},
		{model.Igor, 8088},
		{model.Fiat, 7003},
		{model.Rosco, 8087},
		{model.Kayenta, 8090},
	}
	for _, tt := range tests {
		got := DefaultPort(tt.name)
		if got != tt.want {
			t.Errorf("DefaultPort(%v) = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestDefaultPortUnknown(t *testing.T) {
	got := DefaultPort(model.ServiceName(999))
	if got != 0 {
		t.Errorf("DefaultPort(unknown) = %d, want 0", got)
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "a", "b", "c")
	err := ensureDir(newDir)
	if err != nil {
		t.Fatalf("ensureDir() error = %v", err)
	}
	info, err := os.Stat(newDir)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if !info.IsDir() {
		t.Error("ensureDir() did not create directory")
	}
}

func TestEnsureDirAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	// Should not error when directory already exists.
	err := ensureDir(tmpDir)
	if err != nil {
		t.Fatalf("ensureDir() error = %v", err)
	}
}

func TestSaveToFileCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nested := filepath.Join(tmpDir, "sub", "dir", "config.yaml")
	cfg := NewDefault()
	cfg.Version = "1.0.0"
	if err := SaveToFile(cfg, nested); err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}
	if _, err := os.Stat(nested); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestSaveToFileReadOnlyDir(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a read-only file where MkdirAll needs to create a child dir.
	blocker := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0400); err != nil {
		t.Fatal(err)
	}
	cfg := NewDefault()
	// Try to save inside a path that goes through the blocker file (not a dir).
	err := SaveToFile(cfg, filepath.Join(blocker, "sub", "config.yaml"))
	if err == nil {
		t.Error("expected error when directory cannot be created")
	}
}

func TestDefaultConfigDirWithHome(t *testing.T) {
	// Ensure HOME is set so we exercise the success path.
	home := os.Getenv("HOME")
	if home == "" {
		t.Setenv("HOME", "/tmp")
	}
	dir := DefaultConfigDir()
	if !filepath.IsAbs(dir) {
		t.Errorf("DefaultConfigDir() = %q, expected absolute path", dir)
	}
	if filepath.Base(dir) != ".spinctl" {
		t.Errorf("DefaultConfigDir() base = %q, want '.spinctl'", filepath.Base(dir))
	}
}

func TestAcquireLockEnsureDirFailure(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a file where a directory is expected.
	blocker := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0400); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(blocker, "sub", ".lock")
	_, err := AcquireLock(lockPath)
	if err == nil {
		t.Error("expected error when ensureDir fails")
	}
}
