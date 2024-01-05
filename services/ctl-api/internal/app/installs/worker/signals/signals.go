package signals

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Operation string

const (
	OperationProvision          Operation = "provision"
	OperationDeprovision        Operation = "deprovision"
	OperationDelete             Operation = "delete"
	OperationReprovision        Operation = "reprovision"
	OperationDeploy             Operation = "deploy"
	OperationForgotten          Operation = "forgotten"
	OperationPollDependencies   Operation = "poll_dependencies"
	OperationDeployComponents   Operation = "deploy_components"
	OperationTeardownComponents Operation = "teardown_components"
)

type Signal struct {
	Operation Operation `validate:"required" json:"operation"`

	DeployID string `validate:"required_if=Operation deploy" json:"deploy_id"`

	Async bool `json:"async"`
}

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return nil
}
