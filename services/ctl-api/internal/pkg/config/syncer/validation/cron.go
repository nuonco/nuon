package validation

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/robfig/cron"
)

// ValidateCronSchedule validates a cron expression string.
// Returns an error if the cron expression is invalid.
// Duplicates logic from services/ctl-api/internal/pkg/validator/cron_schedule.go
func ValidateCronSchedule(cronExpr string) error {
	if cronExpr == "" {
		return nil // Empty is allowed (optional field)
	}

	_, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid cron expression: %w", err),
			Description: fmt.Sprintf("Cron expression '%s' is invalid. Must be a valid cron format (e.g., '0 0 * * *' for daily at midnight)", cronExpr),
		}
	}

	return nil
}
