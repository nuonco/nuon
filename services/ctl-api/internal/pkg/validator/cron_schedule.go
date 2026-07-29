package validator

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cronutil"
	"github.com/robfig/cron"
)

// MinCronInterval is the minimum allowed interval between consecutive fires of
// any user-defined cron schedule (e.g. action triggers, drift scans).
const MinCronInterval = 5 * time.Minute

type cronScheduleString struct {
	Val string `validate:"cron_schedule"`
}

func CronSchedule(v *validator.Validate, val string) error {
	obj := cronScheduleString{
		Val: val,
	}

	return v.Struct(obj)
}

func cronScheduleValidator(fl validator.FieldLevel) bool {
	cronExpr := fl.Field().String()
	if cronExpr == "" {
		return true
	}

	sched, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return false
	}

	return MinScheduleInterval(sched) >= MinCronInterval
}

// MinScheduleInterval returns the smallest gap between two consecutive fires
// of the given schedule.
func MinScheduleInterval(sched cron.Schedule) time.Duration {
	return cronutil.MinScheduleInterval(sched)
}
