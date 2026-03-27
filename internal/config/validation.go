package config

import "fmt"

// Validate checks a SpinctlConfig for common configuration errors and returns
// a slice of all problems found. An empty slice means the config is valid.
func Validate(cfg *SpinctlConfig) []error {
	var errs []error

	if cfg.Version == "" {
		errs = append(errs, fmt.Errorf("version is required"))
	}

	hasEnabled := false
	for name, svc := range cfg.Services {
		if svc.Enabled {
			hasEnabled = true
			if svc.Port == 0 {
				errs = append(errs, fmt.Errorf("service %s: port is required for enabled services", name))
			}
			if svc.Port < 0 || svc.Port > 65535 {
				errs = append(errs, fmt.Errorf("service %s: port %d is out of range (1-65535)", name, svc.Port))
			}
			if svc.Host == "" {
				errs = append(errs, fmt.Errorf("service %s: host is required for enabled services", name))
			}
		} else if svc.Port != 0 && (svc.Port < 1 || svc.Port > 65535) {
			errs = append(errs, fmt.Errorf("service %s: port %d is out of range (1-65535)", name, svc.Port))
		}
	}

	if !hasEnabled {
		errs = append(errs, fmt.Errorf("at least one service must be enabled"))
	}

	return errs
}
