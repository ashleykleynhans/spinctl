package halimport

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spinnaker/spinctl/internal/config"
)

func TestImportFromHalDir(t *testing.T) {
	// Create a fake .hal directory with the basic config.
	tmpDir := t.TempDir()
	halDir := filepath.Join(tmpDir, ".hal")
	if err := os.MkdirAll(halDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Copy the basic test config as the hal config file.
	srcData, err := os.ReadFile(testdataPath("basic_hal_config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(halDir, "config"), srcData, 0600); err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(tmpDir, "spinctl", "config.yaml")

	result, err := Import(ImportOptions{
		HalDir:     halDir,
		OutputPath: outputPath,
		Backup:     true,
	})
	if err != nil {
		t.Fatalf("Import: %v", err)
	}

	// Check output file was created.
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output file not found: %v", err)
	}

	// Check backup was created.
	if result.BackupPath == "" {
		t.Fatal("expected backup path")
	}
	if _, err := os.Stat(result.BackupPath); err != nil {
		t.Fatalf("backup dir not found: %v", err)
	}

	// Check version.
	if result.Config.Version != "1.35.0" {
		t.Errorf("version = %q, want %q", result.Config.Version, "1.35.0")
	}

	// Check deployment name defaulted to currentDeployment.
	if result.DeploymentName != "default" {
		t.Errorf("deployment = %q, want %q", result.DeploymentName, "default")
	}

	// Check unmapped fields are reported.
	if len(result.UnmappedFields) == 0 {
		t.Error("expected unmapped fields")
	}

	// Verify the saved file can be loaded back.
	loaded, err := config.LoadFromFile(outputPath)
	if err != nil {
		t.Fatalf("LoadFromFile: %v", err)
	}
	if loaded.Version != "1.35.0" {
		t.Errorf("loaded version = %q, want %q", loaded.Version, "1.35.0")
	}
}

func TestImportNonexistentHalDir(t *testing.T) {
	_, err := Import(ImportOptions{
		HalDir:     "/nonexistent/path/.hal",
		OutputPath: "/tmp/spinctl-test-output.yaml",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent hal dir")
	}
}

func TestDetectHalDir(t *testing.T) {
	// Create a temporary home directory with .hal/config.
	tmpDir := t.TempDir()
	halDir := filepath.Join(tmpDir, ".hal")
	if err := os.MkdirAll(halDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(halDir, "config"), []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	// Override HOME for the test.
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	detected := DetectHalDir()
	if detected != halDir {
		t.Errorf("DetectHalDir() = %q, want %q", detected, halDir)
	}
}

func TestDetectHalDirNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	detected := DetectHalDir()
	if detected != "" {
		t.Errorf("DetectHalDir() = %q, want empty string", detected)
	}
}
