package validation

import (
	"errors"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// ValidateBuildTimeout validates a build timeout duration string.
// Returns an error if the format is invalid or the value is out of range.
// Duplicates logic from services/ctl-api/internal/app/components/service/shared_validation.go
func ValidateBuildTimeout(timeout string) error {
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

// ValidateDeployTimeout validates a deploy timeout duration string.
// Returns an error if the format is invalid or the value is out of range.
// Duplicates logic from services/ctl-api/internal/app/components/service/shared_validation.go
func ValidateDeployTimeout(timeout string) error {
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

// ValidateHealthStabilizationWindow validates a health stabilization window duration string.
// Returns an error if the format is invalid or the value is out of range.
// Duplicates logic from services/ctl-api/internal/app/components/service/shared_validation.go
func ValidateHealthStabilizationWindow(window string) error {
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
