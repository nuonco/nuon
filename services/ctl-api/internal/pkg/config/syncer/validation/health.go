package validation

import (
	"errors"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// ValidateHealthProbes validates the probes declared on a component health
// block, using the same rules the CLI applies before sync.
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

// ToAppHealthProbes converts declared probes into the JSONB payload persisted on
// the component config connection.
func ToAppHealthProbes(health *config.ComponentHealthConfig) app.ComponentHealthProbes {
	if health == nil || len(health.Probes) == 0 {
		return nil
	}

	out := make(app.ComponentHealthProbes, 0, len(health.Probes))
	for _, probe := range health.Probes {
		out = append(out, app.ComponentHealthProbe{
			Type:    probe.Type,
			Name:    probe.Name,
			URL:     probe.URL,
			Command: probe.Command,
		})
	}
	return out
}
