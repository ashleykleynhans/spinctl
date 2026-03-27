package deploy

import (
	"context"
	"fmt"
	"os/exec"
)

// RealExecutor runs commands via os/exec.
type RealExecutor struct{}

// Run executes a command with the given name and arguments.
func (r *RealExecutor) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command %q failed: %w\noutput: %s", name, err, string(output))
	}
	return nil
}
