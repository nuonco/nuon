package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	interpolatedNameRegex *regexp.Regexp = regexp.MustCompile("^[a-zA-Z0-9_]*$")
)

func interpolatedNameValidator(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return false
	}

	return interpolatedNameRegex.MatchString(fl.Field().String())
}
