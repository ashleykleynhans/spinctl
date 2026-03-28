package deploy

import (
	"context"
	"fmt"

	"github.com/spinnaker/spinctl/internal/model"
)

// DebianDeployer handles Debian-based Spinnaker deployments using apt.
type DebianDeployer struct {
	exec      Executor
	configDir string
}

// NewDebianDeployer creates a new DebianDeployer.
func NewDebianDeployer(exec Executor, configDir string) *DebianDeployer {
	return &DebianDeployer{exec: exec, configDir: configDir}
}

// CheckSudo verifies the current user has sudo privileges.
func (d *DebianDeployer) CheckSudo(ctx context.Context) error {
	return d.exec.Run(ctx, "sudo", "-n", "true")
}

// UpdateApt runs apt-get update.
func (d *DebianDeployer) UpdateApt(ctx context.Context) error {
	return d.exec.Run(ctx, "sudo", "apt-get", "update", "-qq")
}

// DeployService installs a specific version of a Spinnaker service package
// and restarts its systemd unit.
func (d *DebianDeployer) DeployService(ctx context.Context, name model.ServiceName, version string) error {
	pkg := fmt.Sprintf("%s=%s", name.PackageName(), version)
	if err := d.exec.Run(ctx, "sudo", "apt-get", "install", "-y", "-qq", pkg); err != nil {
		return fmt.Errorf("installing %s: %w", name, err)
	}

	if name == model.Deck {
		// Deck doesn't have a systemd service.
		return nil
	}

	if err := d.exec.Run(ctx, "sudo", "systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("daemon-reload for %s: %w", name, err)
	}

	if err := d.exec.Run(ctx, "sudo", "systemctl", "restart", name.SystemdUnit()); err != nil {
		return fmt.Errorf("restarting %s: %w", name, err)
	}

	return nil
}

