package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	interpolatedNameRegex *regexp.Regexp = regexp.MustCompile("^[a-z0-9_]*$")
)

type interpolatedNameString struct {
	Val string `validate:"interpolatedName"`
}

func InterpolatedName(v *validator.Validate, val string) error {
	obj := interpolatedNameString{
		Val: val,
	}

	return v.Struct(obj)
}

func interpolatedNameValidator(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return false
	}

	return interpolatedNameRegex.MatchString(fl.Field().String())
}
