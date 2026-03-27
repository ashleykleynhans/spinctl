// Package halimport provides tools for parsing Halyard configuration files
// and importing them into the spinctl configuration format.
package halimport

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/spinnaker/spinctl/internal/config"
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

	if dc.Providers != nil {
		cfg.Providers = mapProviders(dc.Providers)
	}

	if dc.Security != nil {
		cfg.Security = mapSecurity(dc.Security)
	}

	if dc.Features != nil {
		cfg.Features = dc.Features
	}

	// Store unmapped fields in Custom.
	if len(dc.Extra) > 0 {
		cfg.Custom = make(map[string]any, len(dc.Extra))
		for k, v := range dc.Extra {
			cfg.Custom[k] = v
		}
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
