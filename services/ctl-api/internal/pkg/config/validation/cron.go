package validation

import (
	"fmt"

	"github.com/robfig/cron"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

func ValidateCronSchedule(cronExpr string) error {
	if cronExpr == "" {
		return nil
	}

	sched, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid cron expression: %w", err),
			Description: fmt.Sprintf("Cron expression '%s' is invalid. Must be a valid cron format (e.g., '0 0 * * *' for daily at midnight)", cronExpr),
		}
	}

	if minInterval := validatorPkg.MinScheduleInterval(sched); minInterval < validatorPkg.MinCronInterval {
		return stderr.ErrUser{
			Err:         fmt.Errorf("cron schedule fires too frequently: %s", minInterval),
			Description: fmt.Sprintf("Cron expression '%s' fires every %s, which is more often than the allowed minimum of %s.", cronExpr, minInterval, validatorPkg.MinCronInterval),
		}
	}

	return nil
}
