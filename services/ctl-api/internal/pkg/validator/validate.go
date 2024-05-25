package validator

import "github.com/go-playground/validator/v10"

func New() *validator.Validate {
	v := validator.New()

	v.RegisterValidation("interpolatedName", interpolatedNameValidator)
	v.RegisterValidation("entityName", entityNameValidator)
	return v
}
