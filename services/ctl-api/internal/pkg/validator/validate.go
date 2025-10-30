package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// FormatValidationError formats validation errors with custom messages for specific tags
func FormatValidationError(err error) error {
	if err == nil {
		return nil
	}

	if errs, ok := err.(validator.ValidationErrors); ok {
		var errMsg string
		for _, e := range errs {
			var fieldErr string
			switch e.Tag() {
			case "interpolated_name":
				fieldErr = fmt.Sprintf("%s: Name must contain only lowercase letters, numbers, underscores, dots, and curly braces (for interpolation)", e.Field())
			case "entity_name":
				fieldErr = fmt.Sprintf("%s: Name must contain only lowercase letters, numbers, underscores, and hyphens", e.Field())
			default:
				fieldErr = fmt.Sprintf("%s: %s", e.Field(), e.Error())
			}
			errMsg += fieldErr + "\n"
		}
		return fmt.Errorf("invalid request:\n%s", errMsg)
	}

	return err
}

func New() *validator.Validate {
	v := validator.New()

	v.RegisterValidation("interpolated_name", interpolatedNameValidator)
	v.RegisterValidation("entity_name", entityNameValidator)
	v.RegisterValidation("cron_schedule", cronScheduleValidator)
	v.RegisterValidation("optional_json", optionalJSONValidator)

	return v
}
