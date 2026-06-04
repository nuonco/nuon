package validator

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/robfig/cron"
)

// MinCronInterval is the minimum allowed interval between consecutive fires of
// any user-defined cron schedule (e.g. action triggers, drift scans).
const MinCronInterval = 5 * time.Minute

// minIntervalProbeFires is the number of consecutive scheduled fires we walk
// when computing the smallest interval a schedule produces. Crons can have
// variable gaps (e.g. weekday-only or "0,15,30,45,46 * * * *"), so we sample
// enough fires to surface the shortest gap in the cycle.
const minIntervalProbeFires = 200

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
// of the given schedule, sampled over the next minIntervalProbeFires fires.
//
// Exposed so callers outside the struct-tag flow (e.g. the syncer-side
// `ValidateCronSchedule` helper) can apply the same rule without re-parsing.
func MinScheduleInterval(sched cron.Schedule) time.Duration {
	now := time.Now().UTC()
	prev := sched.Next(now)
	if prev.IsZero() {
		// schedule never fires (e.g. impossible date) — treat as infinite gap.
		return time.Duration(1<<63 - 1)
	}
	min := time.Duration(1<<63 - 1)
	for i := 0; i < minIntervalProbeFires; i++ {
		next := sched.Next(prev)
		if next.IsZero() {
			// no further fires within the underlying parser's search horizon
			// (robfig/cron caps at ~5 years). The gaps we already sampled are
			// representative; stop probing rather than computing a bogus delta.
			break
		}
		if d := next.Sub(prev); d < min {
			min = d
		}
		prev = next
	}
	return min
}
