package deploy

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/spinnaker/spinctl/internal/model"
)

// BOMService holds the version info for a single Spinnaker service in the BOM.
type BOMService struct {
	Version string `yaml:"version"`
}

// BOM represents a Spinnaker Bill of Materials file.
type BOM struct {
	Version      string                 `yaml:"version"`
	Timestamp    string                 `yaml:"timestamp"`
	Services     map[string]BOMService  `yaml:"services"`
	Dependencies map[string]BOMService  `yaml:"dependencies,omitempty"`
}

// parseBOMFile reads and parses a BOM from a YAML file on disk.
func parseBOMFile(path string) (*BOM, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading BOM file: %w", err)
	}
	return parseBOMBytes(data)
}

// parseBOMBytes parses BOM YAML from raw bytes.
func parseBOMBytes(data []byte) (*BOM, error) {
	var bom BOM
	if err := yaml.Unmarshal(data, &bom); err != nil {
		return nil, fmt.Errorf("parsing BOM: %w", err)
	}
	if bom.Version == "" {
		return nil, fmt.Errorf("BOM missing required version field")
	}
	return &bom, nil
}

// ServiceVersion returns the version string for a given service from the BOM.
func (b *BOM) ServiceVersion(name model.ServiceName) (string, error) {
	svc, ok := b.Services[name.String()]
	if !ok {
		return "", fmt.Errorf("service %q not found in BOM %s", name, b.Version)
	}
	return svc.Version, nil
}

// ResolveVersions builds a map of service names to their resolved versions.
// Config-level overrides take precedence over BOM versions.
func ResolveVersions(bom *BOM, overrides map[model.ServiceName]string, services []model.ServiceName) (map[model.ServiceName]string, error) {
	resolved := make(map[model.ServiceName]string, len(services))
	for _, svc := range services {
		if v, ok := overrides[svc]; ok {
			resolved[svc] = v
			continue
		}
		v, err := bom.ServiceVersion(svc)
		if err != nil {
			return nil, err
		}
		resolved[svc] = v
	}
	return resolved, nil
}
