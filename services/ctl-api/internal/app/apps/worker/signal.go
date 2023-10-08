package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Operation string

const (
	OperationProvision        Operation = "provision"
	OperationPollDependencies Operation = "poll_dependencies"
	OperationDeprovision      Operation = "deprovision"
	OperationReprovision      Operation = "reprovision"
	OperationUpdateSandbox    Operation = "update_sandbox"
)

type Signal struct {
	DryRun    bool
	Operation Operation `validate:"required"`

	SandboxReleaseID string `validate:"required_if=Operation update_sandbox"`
}

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}
