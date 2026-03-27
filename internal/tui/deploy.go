package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spinnaker/spinctl/internal/deploy"
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

// DeployPage shows the deploy confirmation and progress.
type DeployPage struct {
	plan      *deploy.DeployPlan
	confirmed bool
	cancelled bool
	statuses  []DeployStatus
	current   int
	done      bool
	err       error
}

// NewDeployPage creates a deploy page for the given plan.
func NewDeployPage(plan *deploy.DeployPlan) *DeployPage {
	dp := &DeployPage{
		plan: plan,
	}
	if plan != nil {
		dp.statuses = make([]DeployStatus, len(plan.Steps))
	}
	return dp
}

// Update handles input for the deploy page.
func (d *DeployPage) Update(msg tea.Msg) (page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !d.confirmed && !d.cancelled {
			switch msg.String() {
			case "y":
				d.confirmed = true
			case "n", "esc":
				d.cancelled = true
			}
		}
	}
	return d, nil
}

// View renders the deploy page.
func (d *DeployPage) View() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(headingStyle.Render("Deploy"))
	b.WriteString("\n\n")

	if d.plan == nil {
		b.WriteString("  " + valueStyle.Render("No deploy plan available.") + "\n")
		return b.String()
	}

	// Show warnings.
	if len(d.plan.Warnings) > 0 {
		b.WriteString("  Warnings:\n")
		for _, w := range d.plan.Warnings {
			b.WriteString(fmt.Sprintf("    ⚠ %s\n", w))
		}
		b.WriteString("\n")
	}

	// Show steps with status.
	for i, step := range d.plan.Steps {
		status := Pending
		if i < len(d.statuses) {
			status = d.statuses[i]
		}
		services := make([]string, len(step.Services))
		for j, svc := range step.Services {
			services[j] = svc.String()
		}
		b.WriteString(fmt.Sprintf("  %s Step %d: %s\n", status.icon(), i+1, strings.Join(services, ", ")))
	}

	b.WriteString("\n")

	if d.cancelled {
		b.WriteString("  Deploy cancelled.\n")
	} else if d.done {
		if d.err != nil {
			b.WriteString(fmt.Sprintf("  Deploy failed: %s\n", d.err))
		} else {
			b.WriteString("  Deploy complete.\n")
		}
	} else if !d.confirmed {
		b.WriteString("  Proceed with deploy? (y/n)\n")
	} else {
		b.WriteString("  Deploying...\n")
	}

	return b.String()
}
