package deploy

import (
	"context"
	"fmt"

	"github.com/spinnaker/spinctl/internal/model"
)

// Executor abstracts running shell commands so implementations can be swapped
// for testing.
type Executor interface {
	Run(ctx context.Context, name string, args ...string) error
}

// DeployStep represents one step in a deploy plan, containing services that
// can be deployed in parallel.
type DeployStep struct {
	Services []model.ServiceName
}

// DeployPlan is an ordered list of deploy steps.
type DeployPlan struct {
	Steps    []DeployStep
	Warnings []string
}

// serviceDependencies maps each service to the services it depends on.
var serviceDependencies = map[model.ServiceName][]model.ServiceName{
	model.Fiat:        {model.Front50},
	model.Clouddriver: {model.Front50, model.Fiat},
	model.Orca:        {model.Clouddriver, model.Front50, model.Fiat},
	model.Echo:        {model.Orca, model.Front50},
	model.Igor:        {model.Echo},
	model.Rosco:       {},
	model.Kayenta:     {},
	model.Gate:        {model.Orca, model.Clouddriver, model.Front50, model.Fiat},
	model.Deck:        {model.Gate},
	model.Front50:     {},
}

// BuildDeployPlan creates an ordered deployment plan. If filter is non-empty,
// only the listed services are included but warnings are emitted for any
// missing dependencies.
func BuildDeployPlan(filter []model.ServiceName) *DeployPlan {
	plan := &DeployPlan{}
	order := model.DeploymentOrder()

	if len(filter) == 0 {
		for _, group := range order {
			plan.Steps = append(plan.Steps, DeployStep{Services: group})
		}
		return plan
	}

	filterSet := make(map[model.ServiceName]bool, len(filter))
	for _, s := range filter {
		filterSet[s] = true
	}

	// Check for missing dependencies.
	for _, s := range filter {
		for _, dep := range serviceDependencies[s] {
			if !filterSet[dep] {
				plan.Warnings = append(plan.Warnings,
					fmt.Sprintf("%s depends on %s which is not in the deploy filter", s, dep))
			}
		}
	}

	for _, group := range order {
		var filtered []model.ServiceName
		for _, s := range group {
			if filterSet[s] {
				filtered = append(filtered, s)
			}
		}
		if len(filtered) > 0 {
			plan.Steps = append(plan.Steps, DeployStep{Services: filtered})
		}
	}

	return plan
}
