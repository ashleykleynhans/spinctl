package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

// DeployResult captures the outcome of deploying a single service.
type DeployResult struct {
	Service   model.ServiceName
	Version   string
	Err       error
	StartedAt time.Time
	Duration  time.Duration
}

// DeployState tracks progress so a deploy can be resumed after failure.
type DeployState struct {
	Completed []string `json:"completed"`
	Remaining []string `json:"remaining"`
}

// DeployRunner orchestrates a full Spinnaker deployment.
type DeployRunner struct {
	deployer  *DebianDeployer
	configDir string
	logFile   string
	stateFile string
}

// NewDeployRunner creates a new DeployRunner.
func NewDeployRunner(exec Executor, configDir, logFile, stateFile string) *DeployRunner {
	return &DeployRunner{
		deployer:  NewDebianDeployer(exec, configDir),
		configDir: configDir,
		logFile:   logFile,
		stateFile: stateFile,
	}
}

// Run executes the deploy plan for the given config, BOM, and optional filter.
// It writes logs, checks for context cancellation before each service, saves
// state on failure, and cleans up state on success.
func (r *DeployRunner) Run(ctx context.Context, cfg *config.SpinctlConfig, bom *BOM, filter []model.ServiceName) ([]DeployResult, error) {
	// Open log file.
	lf, err := os.OpenFile(r.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}
	defer lf.Close()
	logger := log.New(lf, "", log.LstdFlags)

	// Export all config files before deploying.
	logger.Printf("Exporting config to %s", r.configDir)
	if err := ExportConfigs(cfg, r.configDir); err != nil {
		return nil, fmt.Errorf("exporting configs: %w", err)
	}
	logger.Printf("Config export complete")

	plan := BuildDeployPlan(filter)
	for _, w := range plan.Warnings {
		logger.Printf("WARNING: %s", w)
	}

	// Collect all services in order.
	var allServices []model.ServiceName
	for _, step := range plan.Steps {
		allServices = append(allServices, step.Services...)
	}

	var results []DeployResult
	var completed []string

	for _, svc := range allServices {
		// Check for cancellation.
		select {
		case <-ctx.Done():
			logger.Printf("deploy cancelled before %s", svc)
			remaining := serviceNamesToStrings(allServices[len(completed):])
			r.saveState(completed, remaining)
			return results, fmt.Errorf("deploy cancelled: %w", ctx.Err())
		default:
		}

		version, err := bom.ServiceVersion(svc)
		if err != nil {
			return results, fmt.Errorf("resolving version for %s: %w", svc, err)
		}

		// Apply config override if present.
		if v, ok := cfg.ServiceOverrides[svc]; ok {
			version = v
		}

		logger.Printf("deploying %s version %s", svc, version)
		start := time.Now()

		// Write config for service.
		if svcCfg, ok := cfg.Services[svc]; ok {
			if writeErr := r.deployer.WriteServiceConfig(svc, svcCfg); writeErr != nil {
				logger.Printf("ERROR writing config for %s: %v", svc, writeErr)
			}
		}

		// Install and restart.
		deployErr := r.deployer.DeployService(ctx, svc, version)
		duration := time.Since(start)

		result := DeployResult{
			Service:   svc,
			Version:   version,
			Err:       deployErr,
			StartedAt: start,
			Duration:  duration,
		}
		results = append(results, result)

		if deployErr != nil {
			logger.Printf("ERROR deploying %s: %v", svc, deployErr)
			remaining := serviceNamesToStrings(allServices[len(completed):])
			r.saveState(completed, remaining)
			return results, fmt.Errorf("deploying %s: %w", svc, deployErr)
		}

		completed = append(completed, svc.String())
		logger.Printf("deployed %s in %s", svc, duration)
	}

	// Success: clean up state file.
	RemoveDeployState(r.stateFile)
	logger.Printf("deploy complete: %d services", len(results))

	return results, nil
}

func (r *DeployRunner) saveState(completed, remaining []string) {
	state := DeployState{
		Completed: completed,
		Remaining: remaining,
	}
	data, err := json.Marshal(state)
	if err != nil {
		return
	}
	_ = os.WriteFile(r.stateFile, data, 0644)
}

// LoadDeployState reads a deploy state file from disk.
func LoadDeployState(path string) (*DeployState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading deploy state: %w", err)
	}
	var state DeployState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing deploy state: %w", err)
	}
	return &state, nil
}

// RemoveDeployState removes a deploy state file if it exists.
func RemoveDeployState(path string) {
	_ = os.Remove(path)
}

func serviceNamesToStrings(names []model.ServiceName) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = n.String()
	}
	return out
}
