package validator

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var interpolatedNameRegex *regexp.Regexp = regexp.MustCompile(`^[a-z0-9_{}\.]*$`)

type interpolatedNameString struct {
	Val string `validate:"interpolated_name"`
}

func InterpolatedName(v *validator.Validate, val string) error {
	obj := interpolatedNameString{
		Val: val,
	}

	if err := v.Struct(obj); err != nil {
		return fmt.Errorf("validation failed: value '%s' does not match the required pattern '%s'", val, interpolatedNameRegex.String())
	}

	return nil
}

func interpolatedNameValidator(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return true
	}

	return interpolatedNameRegex.MatchString(fl.Field().String())
}
