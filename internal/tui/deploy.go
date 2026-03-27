package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/deploy"
	"github.com/spinnaker/spinctl/internal/model"
)

// DeployStatus represents the status of a deploy step.
type DeployStatus int

const (
	Pending DeployStatus = iota
	Installing
	Restarting
	Done
	Failed
)

func (d DeployStatus) String() string {
	switch d {
	case Pending:
		return "pending"
	case Installing:
		return "installing"
	case Restarting:
		return "restarting"
	case Done:
		return "done"
	case Failed:
		return "failed"
	default:
		return "unknown"
	}
}

func (d DeployStatus) icon() string {
	switch d {
	case Pending:
		return "○"
	case Installing:
		return "◐"
	case Restarting:
		return "◑"
	case Done:
		return "●"
	case Failed:
		return "✗"
	default:
		return "?"
	}
}

// deployBuildMsg is sent after the deploy plan is built.
type deployBuildMsg struct {
	plan     *deploy.DeployPlan
	versions map[model.ServiceName]string
	err      error
}

// deployStepMsg is sent after each service deploy step completes.
type deployStepMsg struct {
	service model.ServiceName
	result  deploy.DeployResult
	stepIdx int
}

// deployDoneMsg is sent when the entire deploy finishes.
type deployDoneMsg struct {
	results []deploy.DeployResult
	err     error
}

// DeployPage shows the deploy confirmation and progress.
type DeployPage struct {
	cfg       *config.SpinctlConfig
	plan      *deploy.DeployPlan
	versions  map[model.ServiceName]string
	confirmed bool
	cancelled bool
	statuses  map[model.ServiceName]DeployStatus
	results   []deploy.DeployResult
	building  bool
	deploying bool
	done      bool
	err       error
	buildErr  error
}

// NewDeployPage creates a deploy page that builds a plan from the config.
func NewDeployPage(cfg *config.SpinctlConfig) *DeployPage {
	dp := &DeployPage{
		cfg:      cfg,
		statuses: make(map[model.ServiceName]DeployStatus),
		building: true,
	}
	return dp
}

// Init builds the deploy plan asynchronously.
func (d *DeployPage) Init() tea.Cmd {
	return d.buildPlan()
}

func (d *DeployPage) buildPlan() tea.Cmd {
	cfg := d.cfg
	return func() tea.Msg {
		// Build plan from enabled services.
		plan := deploy.BuildDeployPlan(nil)

		// Fetch BOM to resolve versions.
		cacheDir := filepath.Join(config.DefaultConfigDir(), "cache", "bom")
		fetcher := deploy.NewBOMFetcher(deploy.DefaultBOMURLPattern, cacheDir)
		bom, err := fetcher.Fetch(cfg.Version)
		if err != nil {
			return deployBuildMsg{err: fmt.Errorf("fetching BOM for %s: %w", cfg.Version, err)}
		}

		// Resolve versions for all services.
		versions := make(map[model.ServiceName]string)
		for _, step := range plan.Steps {
			for _, svc := range step.Services {
				ver, err := bom.ServiceVersion(svc)
				if err == nil {
					versions[svc] = ver
				}
			}
		}

		return deployBuildMsg{plan: plan, versions: versions}
	}
}

// Update handles input for the deploy page.
func (d *DeployPage) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case deployBuildMsg:
		d.building = false
		if msg.err != nil {
			d.buildErr = msg.err
		} else {
			d.plan = msg.plan
			d.versions = msg.versions
		}
		return d, nil

	case deployStepMsg:
		d.statuses[msg.result.Service] = Done
		if msg.result.Err != nil {
			d.statuses[msg.result.Service] = Failed
		}
		d.results = append(d.results, msg.result)
		return d, nil

	case deployDoneMsg:
		d.deploying = false
		d.done = true
		d.err = msg.err
		d.results = msg.results
		return d, nil

	case tea.KeyMsg:
		if d.building || d.deploying {
			return d, nil
		}
		if !d.confirmed && !d.cancelled && !d.done {
			switch msg.String() {
			case "y":
				d.confirmed = true
				d.deploying = true
				return d, d.runDeploy()
			case "n", "esc":
				d.cancelled = true
			}
		}
	}
	return d, nil
}

func (d *DeployPage) runDeploy() tea.Cmd {
	cfg := d.cfg
	return func() tea.Msg {
		cacheDir := filepath.Join(config.DefaultConfigDir(), "cache", "bom")
		fetcher := deploy.NewBOMFetcher(deploy.DefaultBOMURLPattern, cacheDir)
		bom, err := fetcher.Fetch(cfg.Version)
		if err != nil {
			return deployDoneMsg{err: err}
		}

		spinctlDir := config.DefaultConfigDir()
		exec := &deploy.RealExecutor{}
		runner := deploy.NewDeployRunner(exec,
			"/opt/spinnaker/config",
			filepath.Join(spinctlDir, "deploy.log"),
			filepath.Join(spinctlDir, "deploy-state.json"),
		)

		ctx := context.Background()
		results, err := runner.Run(ctx, cfg, bom, nil)
		return deployDoneMsg{results: results, err: err}
	}
}

// View renders the deploy page.
func (d *DeployPage) View() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(headingStyle.Render("Deploy"))
	b.WriteString("\n\n")

	if d.building {
		b.WriteString("  " + keyStyle.Render("Building deploy plan...") + "\n")
		return b.String()
	}

	if d.buildErr != nil {
		b.WriteString("  " + errorStyle.Render(fmt.Sprintf("Error: %s", d.buildErr)) + "\n\n")
		b.WriteString("  " + menuDescStyle.Render("esc: back") + "\n")
		return b.String()
	}

	if d.plan == nil {
		b.WriteString("  " + valueStyle.Render("No deploy plan available.") + "\n")
		return b.String()
	}

	// Show warnings.
	if len(d.plan.Warnings) > 0 {
		for _, w := range d.plan.Warnings {
			b.WriteString("  " + warnStyle.Render("⚠ "+w) + "\n")
		}
		b.WriteString("\n")
	}

	// Count enabled services.
	enabledCount := 0
	for _, svc := range d.cfg.Services {
		if svc.Enabled {
			enabledCount++
		}
	}
	b.WriteString("  " + keyStyle.Render("Spinnaker version: ") + onStyle.Render(d.cfg.Version) + "\n")
	b.WriteString("  " + keyStyle.Render("Services to deploy: ") + valueStyle.Render(fmt.Sprintf("%d", enabledCount)) + "\n\n")

	// Show services with versions and status.
	for _, step := range d.plan.Steps {
		for _, svc := range step.Services {
			ver := d.versions[svc]
			if ver == "" {
				ver = "unknown"
			}

			status := d.statuses[svc]
			var statusStr string
			switch status {
			case Done:
				statusStr = onStyle.Render("✓")
			case Failed:
				statusStr = warnStyle.Render("✗")
			case Installing, Restarting:
				statusStr = editCursorStyle.Render("◐")
			default:
				statusStr = valueStyle.Render("○")
			}

			label := keyStyle.Render(fmt.Sprintf("%-15s", svc))
			version := valueStyle.Render(ver)
			b.WriteString(fmt.Sprintf("  %s %s  %s\n", statusStr, label, version))
		}
	}

	b.WriteString("\n")

	if d.cancelled {
		b.WriteString("  " + warnStyle.Render("Deploy cancelled.") + "\n")
	} else if d.done {
		if d.err != nil {
			b.WriteString("  " + errorStyle.Render(fmt.Sprintf("Deploy failed: %s", d.err)) + "\n")
		} else {
			b.WriteString("  " + successStyle.Render("Deploy complete.") + "\n")
		}
		// Show timing.
		for _, r := range d.results {
			status := onStyle.Render("OK")
			if r.Err != nil {
				status = errorStyle.Render("FAIL")
			}
			b.WriteString(fmt.Sprintf("  %s %-15s %s\n", status,
				keyStyle.Render(r.Service.String()),
				valueStyle.Render(r.Duration.Round(time.Millisecond).String())))
		}
	} else if d.deploying {
		b.WriteString("  " + keyStyle.Render("Deploying...") + "\n")
	} else {
		b.WriteString("  " + keyStyle.Render("y") + menuDescStyle.Render(": deploy  ") +
			keyStyle.Render("esc") + menuDescStyle.Render(": cancel") + "\n")
	}

	return b.String()
}
