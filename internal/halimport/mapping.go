// Package halimport provides tools for parsing Halyard configuration files
// and importing them into the spinctl configuration format.
package halimport

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

// halConfig represents the top-level Halyard configuration file.
type halConfig struct {
	HalyardVersion           string             `yaml:"halyardVersion"`
	CurrentDeployment        string             `yaml:"currentDeployment"`
	DeploymentConfigurations []deploymentConfig `yaml:"deploymentConfigurations"`
}

// deploymentConfig represents a single Halyard deployment configuration.
// Known fields are mapped explicitly; everything else is captured via the
// Extra field using yaml:",inline".
type deploymentConfig struct {
	Name     string              `yaml:"name"`
	Version  string              `yaml:"version"`
	Providers map[string]halProvider `yaml:"providers,omitempty"`
	Security  *halSecurity        `yaml:"security,omitempty"`
	Features  map[string]bool     `yaml:"features,omitempty"`

	// Extra captures all unmapped fields (notifications, metricStores,
	// deploymentEnvironment, etc.).
	Extra map[string]any `yaml:",inline"`
}

// halProvider represents a cloud provider block in Halyard config.
type halProvider struct {
	Enabled  bool          `yaml:"enabled"`
	Accounts []halAccount  `yaml:"accounts,omitempty"`
	Extra    map[string]any `yaml:",inline"`
}

// halAccount represents a provider account in Halyard config.
type halAccount struct {
	Name    string         `yaml:"name"`
	Context string         `yaml:"context,omitempty"`
	Extra   map[string]any `yaml:",inline"`
}

// halSecurity represents the security section of a Halyard deployment.
type halSecurity struct {
	Authn *halAuthToggle `yaml:"authn,omitempty"`
	Authz *halAuthToggle `yaml:"authz,omitempty"`
}

// halAuthToggle represents a simple enabled/disabled auth toggle.
type halAuthToggle struct {
	Enabled bool `yaml:"enabled"`
}

// parseHalFile reads and parses a Halyard YAML configuration file.
func parseHalFile(path string) (*halConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading halyard config: %w", err)
	}

	var hal halConfig
	if err := yaml.Unmarshal(data, &hal); err != nil {
		return nil, fmt.Errorf("parsing halyard config: %w", err)
	}

	return &hal, nil
}

// listDeployments returns the names of all deployment configurations.
func listDeployments(hal *halConfig) []string {
	names := make([]string, len(hal.DeploymentConfigurations))
	for i, dc := range hal.DeploymentConfigurations {
		names[i] = dc.Name
	}
	return names
}

// findDeployment returns the deployment configuration with the given name.
func findDeployment(hal *halConfig, name string) (*deploymentConfig, error) {
	for i := range hal.DeploymentConfigurations {
		if hal.DeploymentConfigurations[i].Name == name {
			return &hal.DeploymentConfigurations[i], nil
		}
	}
	return nil, fmt.Errorf("deployment %q not found", name)
}

// mapHalToSpinctl converts a Halyard deployment configuration into a SpinctlConfig.
func mapHalToSpinctl(hal *halConfig, deploymentName string) (*config.SpinctlConfig, error) {
	dc, err := findDeployment(hal, deploymentName)
	if err != nil {
		return nil, err
	}

	cfg := config.NewDefault()
	cfg.Version = dc.Version

	// Halyard enables all core services by default. Enable them all
	// in the imported config to match the running state.
	for _, name := range model.AllServiceNames() {
		svc := cfg.Services[name]
		svc.Enabled = true
		cfg.Services[name] = svc
	}

	if dc.Providers != nil {
		cfg.Providers = mapProviders(dc.Providers)
	}

	if dc.Security != nil {
		cfg.Security = mapSecurity(dc.Security)
	}

	if dc.Features != nil {
		cfg.Features = dc.Features
	}

	// Map known halyard sections to dedicated config fields.
	knownSections := map[string]func(any){
		"artifacts":             func(v any) { cfg.Artifacts = toMapStringAny(v) },
		"persistentStorage":     func(v any) { cfg.PersistentStorage = toMapStringAny(v) },
		"notifications":         func(v any) { cfg.Notifications = toMapStringAny(v) },
		"ci":                    func(v any) { cfg.CI = toMapStringAny(v) },
		"repository":            func(v any) { cfg.Repository = toMapStringAny(v) },
		"pubsub":                func(v any) { cfg.Pubsub = toMapStringAny(v) },
		"canary":                func(v any) { cfg.Canary = toMapStringAny(v) },
		"webhook":               func(v any) { cfg.Webhook = toMapStringAny(v) },
		"metricStores":          func(v any) { cfg.MetricStores = toMapStringAny(v) },
		"stats":                 func(v any) { cfg.Stats = toMapStringAny(v) },
		"deploymentEnvironment": func(v any) { cfg.DeploymentEnvironment = toMapStringAny(v) },
		"spinnaker":             func(v any) { cfg.Spinnaker = toMapStringAny(v) },
		"timezone": func(v any) {
			if s, ok := v.(string); ok {
				cfg.Timezone = s
			}
		},
	}

	custom := make(map[string]any)
	for k, v := range dc.Extra {
		if handler, ok := knownSections[k]; ok {
			handler(v)
		} else {
			custom[k] = v
		}
	}
	if len(custom) > 0 {
		cfg.Custom = custom
	}

	return cfg, nil
}

// mapProviders converts Halyard provider configurations to spinctl ProviderConfigs.
func mapProviders(halProviders map[string]halProvider) map[string]config.ProviderConfig {
	providers := make(map[string]config.ProviderConfig, len(halProviders))
	for name, hp := range halProviders {
		pc := config.ProviderConfig{
			Enabled: hp.Enabled,
		}
		for _, ha := range hp.Accounts {
			acct := config.ProviderAccount{
				Name:    ha.Name,
				Context: ha.Context,
			}
			if len(ha.Extra) > 0 {
				acct.Extra = ha.Extra
			}
			pc.Accounts = append(pc.Accounts, acct)
		}
		providers[name] = pc
	}
	return providers
}

// toMapStringAny converts an any value to map[string]any, or returns nil.
func toMapStringAny(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

// mapSecurity converts Halyard security configuration to spinctl SecurityConfig.
func mapSecurity(hs *halSecurity) config.SecurityConfig {
	sc := config.SecurityConfig{}
	if hs.Authn != nil {
		sc.Authn = config.AuthnConfig{Enabled: hs.Authn.Enabled}
	}
	if hs.Authz != nil {
		sc.Authz = config.AuthzConfig{Enabled: hs.Authz.Enabled}
	}
	return sc
}
