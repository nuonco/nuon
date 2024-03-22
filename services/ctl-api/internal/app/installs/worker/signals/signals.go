package signals

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

func (o Operation) DisplayName() string {
	str := strings.ReplaceAll(string(o), "_", " ")

	caser := cases.Title(language.English)
	str = caser.String(str)
	return str
}

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
