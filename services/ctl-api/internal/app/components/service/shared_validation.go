package service

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// validateBuildTimeout validates a build timeout duration string.
// Returns an error if the format is invalid or the value is out of range.
func validateBuildTimeout(timeout string) error {
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return stderr.ErrUser{
			Err:         errors.New("invalid_timeout"),
			Code:        "invalid_timeout",
			Description: "timeout must be a valid duration (e.g., '30m', '45m')",
		}
	}

	if d < app.MinBuildTimeout {
		return stderr.ErrUser{
			Err:         errors.New("timeout_too_short"),
			Code:        "timeout_too_short",
			Description: fmt.Sprintf("build timeout must be at least %s", app.MinBuildTimeout),
		}
	}
	if d > app.MaxBuildTimeout {
		return stderr.ErrUser{
			Err:         errors.New("timeout_too_long"),
			Code:        "timeout_too_long",
			Description: fmt.Sprintf("build timeout cannot exceed %s", app.MaxBuildTimeout),
		}
	}
	return nil
}

// validateDeployTimeout validates a deploy timeout duration string.
// Returns an error if the format is invalid or the value is out of range.
func validateDeployTimeout(timeout string) error {
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return stderr.ErrUser{
			Err:         errors.New("invalid_timeout"),
			Code:        "invalid_timeout",
			Description: "timeout must be a valid duration (e.g., '30m', '45m')",
		}
	}

	if d < app.MinDeployTimeout {
		return stderr.ErrUser{
			Err:         errors.New("timeout_too_short"),
			Code:        "timeout_too_short",
			Description: fmt.Sprintf("deploy timeout must be at least %s", app.MinDeployTimeout),
		}
	}
	if d > app.MaxDeployTimeout {
		return stderr.ErrUser{
			Err:         errors.New("timeout_too_long"),
			Code:        "timeout_too_long",
			Description: fmt.Sprintf("deploy timeout cannot exceed %s", app.MaxDeployTimeout),
		}
	}
	return nil
}

// validateHealthStabilizationWindow validates a health stabilization window duration string.
// Returns an error if the format is invalid or the value is out of range.
func validateHealthStabilizationWindow(window string) error {
	d, err := time.ParseDuration(window)
	if err != nil {
		return stderr.ErrUser{
			Err:         errors.New("invalid_health_stabilization_window"),
			Code:        "invalid_health_stabilization_window",
			Description: "health stabilization_window must be a valid duration (e.g., '3m', '10m')",
		}
	}

	if d < app.MinHealthStabilizationWindow {
		return stderr.ErrUser{
			Err:         errors.New("health_stabilization_window_too_short"),
			Code:        "health_stabilization_window_too_short",
			Description: fmt.Sprintf("health stabilization_window must be at least %s", app.MinHealthStabilizationWindow),
		}
	}
	if d > app.MaxHealthStabilizationWindow {
		return stderr.ErrUser{
			Err:         errors.New("health_stabilization_window_too_long"),
			Code:        "health_stabilization_window_too_long",
			Description: fmt.Sprintf("health stabilization_window cannot exceed %s", app.MaxHealthStabilizationWindow),
		}
	}
	return nil
}

// HealthProbeRequest is one synthetic health check declared on a component.
// Command is an argv array, never a shell string.
type HealthProbeRequest struct {
	Type     string   `json:"type,omitempty"`
	Name     string   `json:"name,omitempty"`
	URL      string   `json:"url,omitempty"`
	Command  []string `json:"command,omitempty"`
	Interval string   `json:"interval,omitempty"`
}

func validateHealthProbes(probes []HealthProbeRequest) error {
	for _, probe := range probes {
		cfg := config.ComponentHealthProbeConfig{
			Type:    probe.Type,
			Name:    probe.Name,
			URL:     probe.URL,
			Command: probe.Command,
		}
		if err := cfg.Validate(); err != nil {
			return stderr.ErrUser{
				Err:         errors.New("invalid_health_probe"),
				Code:        "invalid_health_probe",
				Description: err.Error(),
			}
		}
	}
	return nil
}

// validateRequiredChecks rejects names the push endpoint would refuse, so a
// deploy cannot be gated on a check that can never be reported.
func validateRequiredChecks(names []string) error {
	seen := map[string]bool{}
	for _, name := range names {
		if !healthCheckNameRe.MatchString(name) {
			return stderr.ErrUser{
				Err:         errors.New("invalid_required_check"),
				Code:        "invalid_required_check",
				Description: fmt.Sprintf("required check name %q must be 1-100 characters of letters, digits, dots, dashes, or underscores, starting and ending with a letter or digit", name),
			}
		}
		if seen[name] {
			return stderr.ErrUser{
				Err:         errors.New("duplicate_required_check"),
				Code:        "duplicate_required_check",
				Description: fmt.Sprintf("required check %q is listed more than once", name),
			}
		}
		seen[name] = true
	}
	return nil
}

// healthCheckNameRe mirrors the push endpoint's name rule.
var healthCheckNameRe = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9._-]{0,98}[a-zA-Z0-9])?$`)

func toAppRequiredChecks(names []string) app.ComponentHealthRequiredChecks {
	if len(names) == 0 {
		return nil
	}
	return app.ComponentHealthRequiredChecks(names)
}

func toAppHealthProbes(probes []HealthProbeRequest) app.ComponentHealthProbes {
	if len(probes) == 0 {
		return nil
	}

	out := make(app.ComponentHealthProbes, 0, len(probes))
	for _, probe := range probes {
		out = append(out, app.ComponentHealthProbe{
			Type:    probe.Type,
			Name:    probe.Name,
			URL:     probe.URL,
			Command: probe.Command,
		})
	}
	return out
}

func validateMaxAutoRetries(maxAutoRetries int) error {
	if maxAutoRetries < 0 {
		return stderr.ErrUser{
			Err:         errors.New("max_auto_retries_negative"),
			Code:        "max_auto_retries_negative",
			Description: "max_auto_retries cannot be negative",
		}
	}
	if maxAutoRetries > app.MaxAutoRetries {
		return stderr.ErrUser{
			Err:         errors.New("max_auto_retries_too_high"),
			Code:        "max_auto_retries_too_high",
			Description: fmt.Sprintf("max_auto_retries cannot exceed %d", app.MaxAutoRetries),
		}
	}
	return nil
}
