package validation

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func ValidateHealthProbes(health *config.ComponentHealthConfig) error {
	if err := health.Validate(); err != nil {
		return stderr.ErrUser{
			Err:         errors.New("invalid_health_probe"),
			Code:        "invalid_health_probe",
			Description: err.Error(),
		}
	}
	return nil
}

func ToAppHealthProbes(health *config.ComponentHealthConfig) app.ComponentHealthProbes {
	if health == nil {
		return nil
	}
	return ToAppHealthProbesFromList(health.Probes)
}

func ToAppHealthProbesFromList(probes []config.ComponentHealthProbeConfig) app.ComponentHealthProbes {
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

func ValidateHealthProbeList(probes []config.ComponentHealthProbeConfig) error {
	for _, probe := range probes {
		if err := probe.Validate(); err != nil {
			return stderr.ErrUser{
				Err:         errors.New("invalid_health_probe"),
				Code:        "invalid_health_probe",
				Description: err.Error(),
			}
		}
	}
	return nil
}

var healthCheckNameRe = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9._-]{0,98}[a-zA-Z0-9])?$`)

// why: ValidateRequiredChecks rejects names the push endpoint would refuse, so a
// deploy cannot be gated on a check that can never be reported.
func ValidateRequiredChecks(names []string) error {
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

func ToAppRequiredChecks(names []string) app.ComponentHealthRequiredChecks {
	if len(names) == 0 {
		return nil
	}
	return app.ComponentHealthRequiredChecks(names)
}
