package validation

import (
	"errors"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// ValidateTeardown validates a component teardown block, using the same rules
// the CLI applies before sync.
func ValidateTeardown(teardown *config.ComponentTeardownConfig) error {
	if err := teardown.Validate(); err != nil {
		return stderr.ErrUser{
			Err:         errors.New("invalid_teardown"),
			Code:        "invalid_teardown",
			Description: err.Error(),
		}
	}
	return nil
}

// ToAppTeardown converts a teardown block into the JSONB payload persisted on
// the component config connection.
func ToAppTeardown(teardown *config.ComponentTeardownConfig) *app.ComponentTeardown {
	if teardown == nil || len(teardown.Retain) == 0 {
		return nil
	}
	return &app.ComponentTeardown{
		Retain: append([]string(nil), teardown.Retain...),
	}
}
