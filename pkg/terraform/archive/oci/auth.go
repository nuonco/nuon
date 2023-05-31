package oci

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Auth struct {
	Username string `validate:"required"`
	Token    string `validate:"required"`
}

func (a *Auth) Validate(v *validator.Validate) error {
	if err := v.Struct(a); err != nil {
		return fmt.Errorf("unable to validate auth: %w", err)
	}

	return nil
}
