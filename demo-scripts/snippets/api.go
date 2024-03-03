package api

import "github.com/go-playground/validator/v10"

type service struct {
	v *validator.Validate
}
