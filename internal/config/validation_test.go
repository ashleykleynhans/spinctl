package config

import (
	"strings"
	"testing"

	"github.com/spinnaker/spinctl/internal/model"
)

func TestValidateValidConfig(t *testing.T) {
	cfg, err := LoadFromFile("testdata/valid_config.yaml")
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}
	errs := Validate(cfg)
	if len(errs) > 0 {
		t.Errorf("Validate() returned errors for valid config: %v", errs)
	}
}

func TestValidateMissingVersion(t *testing.T) {
	cfg := NewDefault()
	cfg.Version = ""
	cfg.Services[model.Gate] = ServiceConfig{Enabled: true, Host: "localhost", Port: 8084}
	errs := Validate(cfg)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "version") {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for missing version")
	}
}

func TestValidateNoEnabledServices(t *testing.T) {
	cfg := NewDefault()
	cfg.Version = "1.35.0"
	errs := Validate(cfg)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "enabled") {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for no enabled services")
	}
}

func TestValidateInvalidPort(t *testing.T) {
	cfg := NewDefault()
	cfg.Version = "1.35.0"
	cfg.Services[model.Gate] = ServiceConfig{Enabled: true, Host: "localhost", Port: 99999}
	errs := Validate(cfg)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "port") {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for invalid port")
	}
}

func TestValidateEnabledServiceMissingPort(t *testing.T) {
	cfg := NewDefault()
	cfg.Version = "1.35.0"
	cfg.Services[model.Gate] = ServiceConfig{Enabled: true, Host: "localhost", Port: 0}
	errs := Validate(cfg)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "port") {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for enabled service missing port")
	}
}

func TestValidateEmptyHost(t *testing.T) {
	cfg := NewDefault()
	cfg.Version = "1.35.0"
	cfg.Services[model.Gate] = ServiceConfig{Enabled: true, Host: "", Port: 8084}
	errs := Validate(cfg)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "host") {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for empty host")
	}
}

func TestValidateDisabledServiceInvalidPort(t *testing.T) {
	cfg := NewDefault()
	cfg.Version = "1.35.0"
	cfg.Services[model.Gate] = ServiceConfig{Enabled: true, Host: "localhost", Port: 8084}
	cfg.Services[model.Igor] = ServiceConfig{Enabled: false, Port: -5}
	errs := Validate(cfg)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "igor") && strings.Contains(e.Error(), "port") {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for disabled service with invalid port")
	}
}

func TestValidateNegativePort(t *testing.T) {
	cfg := NewDefault()
	cfg.Version = "1.35.0"
	cfg.Services[model.Gate] = ServiceConfig{Enabled: true, Host: "localhost", Port: -1}
	errs := Validate(cfg)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "port") {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for negative port")
	}
}

func TestValidateMultipleErrors(t *testing.T) {
	cfg := NewDefault()
	// No version, no enabled services, and invalid port on a disabled service.
	cfg.Services[model.Gate] = ServiceConfig{Enabled: false, Port: 99999}
	errs := Validate(cfg)
	if len(errs) < 2 {
		t.Errorf("expected at least 2 errors, got %d: %v", len(errs), errs)
	}
}
