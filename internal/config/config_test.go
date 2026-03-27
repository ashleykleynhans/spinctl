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
