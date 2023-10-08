package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Operation string

const (
	OperationProvision   Operation = "provision"
	OperationDeprovision Operation = "deprovision"
	OperationReprovision Operation = "reprovision"
)

type Signal struct {
	DryRun    bool
	Operation Operation `validate:"required"`
}

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}
