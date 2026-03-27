package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/spinnaker/spinctl/internal/model"
)

const CurrentSchemaVersion = 1

type SpinctlConfig struct {
	SchemaVersion    int                                 `yaml:"schema_version"`
	Version          string                              `yaml:"version"`
	ServiceOverrides map[model.ServiceName]string        `yaml:"service_overrides,omitempty"`
	Services         map[model.ServiceName]ServiceConfig `yaml:"services"`
	Providers        map[string]ProviderConfig           `yaml:"providers,omitempty"`
	Security         SecurityConfig                      `yaml:"security,omitempty"`
	Features         map[string]bool                     `yaml:"features,omitempty"`
	AptRepository    string                              `yaml:"apt_repository,omitempty"`
	Custom           map[string]any                      `yaml:"custom,omitempty"`
}

type ProviderConfig struct {
	Enabled  bool              `yaml:"enabled"`
	Accounts []ProviderAccount `yaml:"accounts,omitempty"`
	Extra    map[string]any    `yaml:",inline"`
}

type ProviderAccount struct {
	Name    string         `yaml:"name"`
	Context string         `yaml:"context,omitempty"`
	Extra   map[string]any `yaml:",inline"`
}

type SecurityConfig struct {
	Authn AuthnConfig `yaml:"authn,omitempty"`
	Authz AuthzConfig `yaml:"authz,omitempty"`
}

type AuthnConfig struct {
	Enabled bool `yaml:"enabled"`
}

type AuthzConfig struct {
	Enabled bool `yaml:"enabled"`
}

func LoadFromFile(path string) (*SpinctlConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	return LoadFromBytes(data)
}

func LoadFromBytes(data []byte) (*SpinctlConfig, error) {
	var cfg SpinctlConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.SchemaVersion == 0 {
		cfg.SchemaVersion = CurrentSchemaVersion
	}
	return &cfg, nil
}

func SaveToFile(cfg *SpinctlConfig, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

var defaultPorts = map[model.ServiceName]int{
	model.Clouddriver: 7002,
	model.Orca:        8083,
	model.Gate:        8084,
	model.Front50:     8080,
	model.Echo:        8089,
	model.Igor:        8088,
	model.Fiat:        7003,
	model.Rosco:       8087,
	model.Kayenta:     8090,
	model.Deck:        9000,
}

func NewDefault() *SpinctlConfig {
	services := make(map[model.ServiceName]ServiceConfig, 10)
	for _, name := range model.AllServiceNames() {
		services[name] = ServiceConfig{
			Enabled: false,
			Host:    "localhost",
			Port:    defaultPorts[name],
		}
	}
	return &SpinctlConfig{
		SchemaVersion: CurrentSchemaVersion,
		Services:      services,
		Features:      make(map[string]bool),
	}
}

func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".spinctl"
	}
	return filepath.Join(home, ".spinctl")
}

func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.yaml")
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0700)
}
