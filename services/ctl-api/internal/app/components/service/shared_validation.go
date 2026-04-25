package service

import (
	"errors"
	"fmt"
	"time"

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
