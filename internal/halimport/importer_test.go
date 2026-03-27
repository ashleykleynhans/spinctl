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

	// All fields in the basic fixture should be mapped to dedicated config fields.
	// No unmapped fields expected.
	if len(result.UnmappedFields) > 0 {
		t.Errorf("unexpected unmapped fields: %v", result.UnmappedFields)
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

func TestCountEnabled(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Providers = map[string]config.ProviderConfig{
		"kubernetes": {Enabled: true},
		"aws":        {Enabled: false},
		"gce":        {Enabled: true},
	}
	count := countEnabled(cfg)
	if count != 2 {
		t.Errorf("countEnabled() = %d, want 2", count)
	}
}

func TestCountEnabledNoProviders(t *testing.T) {
	cfg := config.NewDefault()
	count := countEnabled(cfg)
	if count != 0 {
		t.Errorf("countEnabled() = %d, want 0", count)
	}
}

func TestCopyDirRecursive(t *testing.T) {
	src := t.TempDir()
	// Create nested structure.
	subDir := filepath.Join(src, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "file1.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(t.TempDir(), "copy")
	if err := copyDir(src, dst); err != nil {
		t.Fatalf("copyDir: %v", err)
	}

	// Verify files exist.
	data, err := os.ReadFile(filepath.Join(dst, "file1.txt"))
	if err != nil {
		t.Fatalf("reading copied file1: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("file1 content = %q, want 'hello'", string(data))
	}

	data, err = os.ReadFile(filepath.Join(dst, "sub", "file2.txt"))
	if err != nil {
		t.Fatalf("reading copied file2: %v", err)
	}
	if string(data) != "world" {
		t.Errorf("file2 content = %q, want 'world'", string(data))
	}
}

func TestCopyDirNonexistentSrc(t *testing.T) {
	err := copyDir("/nonexistent/path", t.TempDir())
	if err == nil {
		t.Error("expected error for nonexistent source")
	}
}

func TestCopyFileNonexistent(t *testing.T) {
	err := copyFile("/nonexistent/file", filepath.Join(t.TempDir(), "out"))
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestImportWithoutBackup(t *testing.T) {
	tmpDir := t.TempDir()
	halDir := filepath.Join(tmpDir, ".hal")
	if err := os.MkdirAll(halDir, 0700); err != nil {
		t.Fatal(err)
	}

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
		Backup:     false,
	})
	if err != nil {
		t.Fatalf("Import: %v", err)
	}

	if result.BackupPath != "" {
		t.Errorf("expected no backup path, got %q", result.BackupPath)
	}
}

func TestImportUsesCurrentDeployment(t *testing.T) {
	tmpDir := t.TempDir()
	halDir := filepath.Join(tmpDir, ".hal")
	if err := os.MkdirAll(halDir, 0700); err != nil {
		t.Fatal(err)
	}

	srcData, err := os.ReadFile(testdataPath("basic_hal_config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(halDir, "config"), srcData, 0600); err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(tmpDir, "spinctl", "config.yaml")
	// Empty deployment name should use currentDeployment from hal config.
	result, err := Import(ImportOptions{
		HalDir:         halDir,
		DeploymentName: "",
		OutputPath:     outputPath,
	})
	if err != nil {
		t.Fatalf("Import: %v", err)
	}

	if result.DeploymentName != "default" {
		t.Errorf("deployment = %q, want 'default'", result.DeploymentName)
	}
}
