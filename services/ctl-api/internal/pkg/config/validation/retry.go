package validation

import (
	"errors"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// ValidateMaxAutoRetries validates a max auto retries value.
func ValidateMaxAutoRetries(maxAutoRetries int) error {
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
