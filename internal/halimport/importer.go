package halimport

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spinnaker/spinctl/internal/config"
)

// ImportOptions configures the import process.
type ImportOptions struct {
	// HalDir is the path to the .hal directory.
	HalDir string

	// DeploymentName is the Halyard deployment to import.
	// If empty, the currentDeployment from the Halyard config is used.
	DeploymentName string

	// OutputPath is the path to write the spinctl config file.
	OutputPath string

	// Backup controls whether to create a backup of the .hal directory.
	Backup bool
}

// ImportResult contains the outcome of an import operation.
type ImportResult struct {
	// Config is the resulting spinctl configuration.
	Config *config.SpinctlConfig

	// BackupPath is the path to the backup directory, if created.
	BackupPath string

	// DeploymentName is the name of the imported deployment.
	DeploymentName string

	// UnmappedFields lists the top-level Halyard fields that were placed
	// into the Custom catch-all.
	UnmappedFields []string
}

// Import parses a Halyard configuration directory, optionally backs it up,
// maps the configuration to spinctl format, and saves the result.
func Import(opts ImportOptions) (*ImportResult, error) {
	halConfigPath := filepath.Join(opts.HalDir, "config")

	// Parse the halyard config.
	hal, err := parseHalFile(halConfigPath)
	if err != nil {
		return nil, fmt.Errorf("parsing halyard config: %w", err)
	}

	// Determine deployment name.
	deploymentName := opts.DeploymentName
	if deploymentName == "" {
		deploymentName = hal.CurrentDeployment
	}

	// Backup if requested.
	var backupPath string
	if opts.Backup {
		backupPath, err = backupHalDir(opts.HalDir)
		if err != nil {
			return nil, fmt.Errorf("backing up hal dir: %w", err)
		}
	}

	// Map to spinctl config.
	cfg, err := mapHalToSpinctl(hal, deploymentName)
	if err != nil {
		return nil, fmt.Errorf("mapping config: %w", err)
	}

	// Save the config.
	if err := config.SaveToFile(cfg, opts.OutputPath); err != nil {
		return nil, fmt.Errorf("saving config: %w", err)
	}

	// Collect unmapped field names.
	var unmapped []string
	for k := range cfg.Custom {
		unmapped = append(unmapped, k)
	}

	return &ImportResult{
		Config:         cfg,
		BackupPath:     backupPath,
		DeploymentName: deploymentName,
		UnmappedFields: unmapped,
	}, nil
}

// DetectHalDir looks for a .hal/config file in the user's home directory.
// Returns the .hal directory path if found, or an empty string if not.
func DetectHalDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	halDir := filepath.Join(home, ".hal")
	configPath := filepath.Join(halDir, "config")
	if _, err := os.Stat(configPath); err != nil {
		return ""
	}
	return halDir
}

// backupHalDir copies the .hal directory to a timestamped backup location.
func backupHalDir(halDir string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	backupPath := halDir + ".backup." + timestamp

	if err := copyDir(halDir, backupPath); err != nil {
		return "", fmt.Errorf("copying hal dir: %w", err)
	}

	return backupPath, nil
}

// copyDir recursively copies a directory tree.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// countEnabled returns the number of enabled providers in a config.
func countEnabled(cfg *config.SpinctlConfig) int {
	count := 0
	for _, p := range cfg.Providers {
		if p.Enabled {
			count++
		}
	}
	return count
}
