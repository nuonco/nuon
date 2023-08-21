package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Operation string

const (
	OperationProvision   Operation = "provision"
	OperationDeprovision Operation = "deprovision"
	OperationDeploy      Operation = "deploy"
)

type Signal struct {
	DryRun    bool
	Operation Operation `validate:"required"`

	DeployID string `validate:"required_if=Operation deploy"`
}

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return nil
}
