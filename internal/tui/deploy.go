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

// DeployStatus represents the status of a service deploy.
type DeployStatus int

const (
	Pending DeployStatus = iota
	Exporting
	Installing
	Restarting
	Done
	Failed
)

func (d DeployStatus) String() string {
	switch d {
	case Pending:
		return "pending"
	case Exporting:
		return "exporting"
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
	case Exporting:
		return "◐"
	case Installing:
		return "◑"
	case Restarting:
		return "◒"
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

// deployExportMsg is sent after config export completes.
type deployExportMsg struct {
	err error
}

// deployServiceStartMsg signals a service deploy is starting.
type deployServiceStartMsg struct {
	service model.ServiceName
	phase   string // "installing" or "restarting"
}

// deployServiceDoneMsg is sent after a single service deploy completes.
type deployServiceDoneMsg struct {
	service  model.ServiceName
	version  string
	err      error
	duration time.Duration
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
	durations map[model.ServiceName]time.Duration
	errors    map[model.ServiceName]error
	results   []deploy.DeployResult
	building  bool
	exporting bool
	deploying bool
	done      bool
	err       error
	buildErr  error
	current   model.ServiceName // currently deploying service

	// Ordered list of services to deploy.
	serviceOrder []model.ServiceName
	serviceIdx   int
}

// NewDeployPage creates a deploy page that builds a plan from the config.
func NewDeployPage(cfg *config.SpinctlConfig) *DeployPage {
	return &DeployPage{
		cfg:       cfg,
		statuses:  make(map[model.ServiceName]DeployStatus),
		durations: make(map[model.ServiceName]time.Duration),
		errors:    make(map[model.ServiceName]error),
		building:  true,
	}
}

// Init builds the deploy plan asynchronously.
func (d *DeployPage) Init() tea.Cmd {
	return d.buildPlan()
}

func (d *DeployPage) buildPlan() tea.Cmd {
	cfg := d.cfg
	return func() tea.Msg {
		plan := deploy.BuildDeployPlan(nil)

		cacheDir := filepath.Join(config.DefaultConfigDir(), "cache", "bom")
		fetcher := deploy.NewBOMFetcher(deploy.DefaultBOMURLPattern, cacheDir)
		bom, err := fetcher.Fetch(cfg.Version)
		if err != nil {
			return deployBuildMsg{err: fmt.Errorf("fetching BOM for %s: %w", cfg.Version, err)}
		}

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
			// Build ordered service list.
			for _, step := range d.plan.Steps {
				d.serviceOrder = append(d.serviceOrder, step.Services...)
			}
		}
		return d, nil

	case deployExportMsg:
		d.exporting = false
		if msg.err != nil {
			d.done = true
			d.err = msg.err
			return d, nil
		}
		// Start deploying first service.
		d.serviceIdx = 0
		return d, d.deployNext()

	case deployServiceStartMsg:
		switch msg.phase {
		case "installing":
			d.statuses[msg.service] = Installing
		case "restarting":
			d.statuses[msg.service] = Restarting
		}
		d.current = msg.service
		return d, nil

	case deployServiceDoneMsg:
		if msg.err != nil {
			d.statuses[msg.service] = Failed
			d.errors[msg.service] = msg.err
			d.durations[msg.service] = msg.duration
			d.done = true
			d.err = fmt.Errorf("%s failed: %s", msg.service, msg.err)
			return d, nil
		}
		d.statuses[msg.service] = Done
		d.durations[msg.service] = msg.duration
		d.serviceIdx++
		// Deploy next service.
		return d, d.deployNext()

	case deployDoneMsg:
		d.deploying = false
		d.done = true
		d.err = msg.err
		d.results = msg.results
		return d, nil

	case tea.KeyMsg:
		if d.building || d.deploying || d.exporting {
			return d, nil
		}
		if !d.confirmed && !d.cancelled && !d.done {
			switch msg.String() {
			case "y":
				d.confirmed = true
				d.exporting = true
				return d, d.exportConfigs()
			case "n", "esc":
				d.cancelled = true
			}
		}
	}
	return d, nil
}

func (d *DeployPage) exportConfigs() tea.Cmd {
	cfg := d.cfg
	return func() tea.Msg {
		err := deploy.ExportConfigs(cfg, "/opt/spinnaker/config")
		return deployExportMsg{err: err}
	}
}

func (d *DeployPage) deployNext() tea.Cmd {
	if d.serviceIdx >= len(d.serviceOrder) {
		// All done.
		d.deploying = false
		d.done = true
		return nil
	}

	svc := d.serviceOrder[d.serviceIdx]
	ver := d.versions[svc]
	if ver == "" {
		ver = "latest"
	}
	d.statuses[svc] = Installing
	d.current = svc
	d.deploying = true

	return func() tea.Msg {
		start := time.Now()

		exec := &deploy.RealExecutor{}
		pkg := fmt.Sprintf("%s=%s", svc.PackageName(), ver)

		// Install.
		if err := exec.Run(context.Background(), "sudo", "apt-get", "install", "-y", "-qq", pkg); err != nil {
			return deployServiceDoneMsg{service: svc, version: ver, err: err, duration: time.Since(start)}
		}

		// Restart (skip for deck).
		if svc != model.Deck {
			if err := exec.Run(context.Background(), "sudo", "systemctl", "daemon-reload"); err != nil {
				return deployServiceDoneMsg{service: svc, version: ver, err: err, duration: time.Since(start)}
			}
			if err := exec.Run(context.Background(), "sudo", "systemctl", "restart", svc.SystemdUnit()); err != nil {
				return deployServiceDoneMsg{service: svc, version: ver, err: err, duration: time.Since(start)}
			}
		}

		return deployServiceDoneMsg{service: svc, version: ver, duration: time.Since(start)}
	}
}

// progressBar renders a simple progress bar.
func progressBar(current, total int, width int) string {
	if total == 0 || width <= 2 {
		return ""
	}
	filled := (current * width) / total
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return menuDescStyle.Render(fmt.Sprintf("[%s] %d/%d", bar, current, total))
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

	// Version and count.
	enabledCount := len(d.serviceOrder)
	b.WriteString("  " + keyStyle.Render("Spinnaker version: ") + onStyle.Render(d.cfg.Version) + "\n")
	b.WriteString("  " + keyStyle.Render("Services to deploy: ") + valueStyle.Render(fmt.Sprintf("%d", enabledCount)) + "\n")

	// Progress bar when deploying.
	if d.deploying || d.done {
		completedCount := 0
		for _, s := range d.serviceOrder {
			if d.statuses[s] == Done || d.statuses[s] == Failed {
				completedCount++
			}
		}
		b.WriteString("  " + progressBar(completedCount, enabledCount, 30) + "\n")
	}
	b.WriteString("\n")

	// Show services with versions and status.
	for _, svc := range d.serviceOrder {
		ver := d.versions[svc]
		if ver == "" {
			ver = "unknown"
		}

		status := d.statuses[svc]
		var statusIcon string
		var svcStyle func(strs ...string) string

		switch status {
		case Done:
			statusIcon = onStyle.Render("●")
			svcStyle = keyStyle.Render
		case Failed:
			statusIcon = errorStyle.Render("✗")
			svcStyle = errorStyle.Render
		case Installing:
			statusIcon = editCursorStyle.Render("◐")
			svcStyle = keySelectedStyle.Render
		case Restarting:
			statusIcon = editCursorStyle.Render("◒")
			svcStyle = keySelectedStyle.Render
		case Exporting:
			statusIcon = editCursorStyle.Render("◑")
			svcStyle = keySelectedStyle.Render
		default:
			statusIcon = valueStyle.Render("○")
			svcStyle = valueStyle.Render
		}

		line := fmt.Sprintf("  %s %s  %s", statusIcon, svcStyle(fmt.Sprintf("%-15s", svc)), valueStyle.Render(ver))

		// Show duration for completed services.
		if dur, ok := d.durations[svc]; ok {
			line += "  " + menuDescStyle.Render(dur.Round(time.Millisecond).String())
		}

		// Show error for failed services.
		if err, ok := d.errors[svc]; ok {
			line += "\n    " + errorStyle.Render(err.Error())
		}

		b.WriteString(line + "\n")
	}

	b.WriteString("\n")

	if d.exporting {
		b.WriteString("  " + keyStyle.Render("Exporting config to /opt/spinnaker/config/...") + "\n")
	} else if d.cancelled {
		b.WriteString("  " + warnStyle.Render("Deploy cancelled.") + "\n")
	} else if d.done {
		if d.err != nil {
			errStr := d.err.Error()
			if strings.Contains(errStr, "permission denied") {
				b.WriteString("  " + errorStyle.Render("Deploy failed: permission denied. Run spinctl with sudo.") + "\n")
			} else {
				b.WriteString("  " + errorStyle.Render(fmt.Sprintf("Deploy failed: %s", errStr)) + "\n")
			}
		} else {
			b.WriteString("  " + successStyle.Render("Deploy complete.") + "\n")
		}
	} else if d.deploying {
		b.WriteString("  " + keyStyle.Render(fmt.Sprintf("Deploying %s...", d.current)) + "\n")
	} else {
		b.WriteString("  " + keyStyle.Render("y") + menuDescStyle.Render(": deploy  ") +
			keyStyle.Render("esc") + menuDescStyle.Render(": cancel") + "\n")
	}

	return b.String()
}
