package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	entityNameRegex *regexp.Regexp = regexp.MustCompile("^[a-z0-9_-]*$")
)

type entityNameString struct {
	Val string `validate:"entityName"`
}

func entityName(v *validator.Validate, val string) error {
	obj := entityNameString{
		Val: val,
	}

	return v.Struct(obj)
}

func entityNameValidator(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return false
	}

	return entityNameRegex.MatchString(fl.Field().String())
}
