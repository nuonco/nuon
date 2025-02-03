package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/robfig/cron"
)

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

	_, err := cron.ParseStandard(cronExpr)
	return err == nil
}
