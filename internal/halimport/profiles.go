package halimport

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

// importServiceSettings reads .hal/<deployment>/service-settings/<service>.yml
// files and merges them into the config. These files contain host, port,
// enabled overrides, and other service-level settings.
func importServiceSettings(cfg *config.SpinctlConfig, dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // Directory doesn't exist, nothing to import.
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		serviceName := strings.TrimSuffix(entry.Name(), ".yml")
		svcName, err := model.ServiceNameFromString(serviceName)
		if err != nil {
			continue // Unknown service, skip.
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}

		// Parse as generic map to extract known fields.
		var settings map[string]any
		if err := yaml.Unmarshal(data, &settings); err != nil {
			continue
		}

		svc := cfg.Services[svcName]

		// Extract known fields.
		if host, ok := settings["host"].(string); ok && host != "" {
			svc.Host = host
			delete(settings, "host")
		}
		if port, ok := settings["port"].(int); ok && port != 0 {
			svc.Port = port
			delete(settings, "port")
		}
		if enabled, ok := settings["enabled"].(bool); ok {
			svc.Enabled = enabled
			delete(settings, "enabled")
		}

		// Remaining fields go into Settings as a yaml.Node.
		if len(settings) > 0 {
			mergeIntoSettings(&svc, settings)
		}

		cfg.Services[svcName] = svc
	}
}

// importProfiles reads .hal/<deployment>/profiles/<service>-local.yml files
// and merges them into the service's Settings. These are Spring Boot config
// overrides that get placed in /opt/spinnaker/config/<service>-local.yml.
func importProfiles(cfg *config.SpinctlConfig, dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // Directory doesn't exist, nothing to import.
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		// Profiles can be named <service>-local.yml or <service>.yml
		name := strings.TrimSuffix(entry.Name(), ".yml")
		name = strings.TrimSuffix(name, "-local")

		svcName, err := model.ServiceNameFromString(name)
		if err != nil {
			continue // Unknown service, skip.
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}

		var profileSettings map[string]any
		if err := yaml.Unmarshal(data, &profileSettings); err != nil {
			continue
		}

		if len(profileSettings) > 0 {
			svc := cfg.Services[svcName]
			mergeIntoSettings(&svc, profileSettings)
			cfg.Services[svcName] = svc
		}
	}
}

// mergeIntoSettings merges a map into a ServiceConfig's Settings yaml.Node.
func mergeIntoSettings(svc *config.ServiceConfig, settings map[string]any) {
	// Marshal the new settings to yaml.Node.
	newData, err := yaml.Marshal(settings)
	if err != nil {
		return
	}
	var newNode yaml.Node
	if err := yaml.Unmarshal(newData, &newNode); err != nil {
		return
	}
	// Unwrap document node.
	if newNode.Kind == yaml.DocumentNode && len(newNode.Content) > 0 {
		newNode = *newNode.Content[0]
	}

	if svc.Settings.Kind == 0 {
		// No existing settings, use the new node directly.
		svc.Settings = newNode
		return
	}

	// Merge: append new key/value pairs from newNode into existing Settings.
	if svc.Settings.Kind == yaml.MappingNode && newNode.Kind == yaml.MappingNode {
		existing := make(map[string]bool)
		for i := 0; i+1 < len(svc.Settings.Content); i += 2 {
			existing[svc.Settings.Content[i].Value] = true
		}
		for i := 0; i+1 < len(newNode.Content); i += 2 {
			key := newNode.Content[i].Value
			if !existing[key] {
				svc.Settings.Content = append(svc.Settings.Content, newNode.Content[i], newNode.Content[i+1])
			}
		}
	}
}
