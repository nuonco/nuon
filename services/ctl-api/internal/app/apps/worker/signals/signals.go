package signals

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Operation string

const (
	OperationCreated            Operation = "created"
	OperationProvision        Operation = "provision"
	OperationPollDependencies Operation = "poll_dependencies"
	OperationDeprovision      Operation = "deprovision"
	OperationReprovision      Operation = "reprovision"
	OperationUpdateSandbox    Operation = "update_sandbox"
	OperationConfigCreated    Operation = "config_created"
)

type Signal struct {
	Operation Operation `validate:"required"`

	// required for updated sandbox
	SandboxReleaseID string `validate:"required_if=Operation update_sandbox"`

	// required for new app config
	AppConfigID string `validate:"required_if=Operation config_created"`
}

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

