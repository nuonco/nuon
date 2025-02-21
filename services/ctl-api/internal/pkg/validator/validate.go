package validator

import (
	"github.com/go-playground/validator/v10"
)

func New() *validator.Validate {
	v := validator.New()

	v.RegisterValidation("interpolated_name", interpolatedNameValidator)
	v.RegisterValidation("entity_name", entityNameValidator)
	v.RegisterValidation("cron_schedule", cronScheduleValidator)
	return v
}
