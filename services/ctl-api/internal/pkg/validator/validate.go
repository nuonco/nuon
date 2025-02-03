package validator

import "github.com/go-playground/validator/v10"

func New() *validator.Validate {
	v := validator.New()

	v.RegisterValidation("interpolatedName", interpolatedNameValidator)
	v.RegisterValidation("entityName", entityNameValidator)
	v.RegisterValidation("cron_schedule", cronScheduleValidator)
	return v
}
