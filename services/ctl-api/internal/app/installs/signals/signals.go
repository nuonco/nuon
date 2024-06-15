package signals

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

const (
	TemporalNamespace string = "installs"

	OperationCreated            eventloop.SignalType = "created"
	OperationProvision          eventloop.SignalType = "provision"
	OperationDeprovision        eventloop.SignalType = "deprovision"
	OperationDelete             eventloop.SignalType = "delete"
	OperationReprovision        eventloop.SignalType = "reprovision"
	OperationRestart            eventloop.SignalType = "restart"
	OperationDeploy             eventloop.SignalType = "deploy"
	OperationForgotten          eventloop.SignalType = "forgotten"
	OperationPollDependencies   eventloop.SignalType = "poll_dependencies"
	OperationDeployComponents   eventloop.SignalType = "deploy_components"
	OperationTeardownComponents eventloop.SignalType = "teardown_components"
)

type Signal struct {
	Type eventloop.SignalType

	DeployID string `validate:"required_if=Operation deploy" json:"deploy_id"`
	Async    bool   `json:"async"`

	eventloop.BaseSignal
}

var _ eventloop.Signal = (*Signal)(nil)

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return nil
}

func (s *Signal) SignalType() eventloop.SignalType {
	return s.Type
}

func (s *Signal) Namespace() string {
	return TemporalNamespace
}

func (s *Signal) Name() string {
	return string(s.Type)
}

func (s *Signal) Start() bool {
	switch s.Type {
	case OperationCreated:
		return true
	case OperationRestart:
		return true
	default:
	}

	return false
}

func (s *Signal) GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error) {
	org, err := middlewares.OrgFromContext(ctx)
	if err == nil {
		return org, nil
	}

	install := app.Install{}
	res := db.WithContext(ctx).
		Preload("Org").
		First(&install, "id = ?", id)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install.Org, nil
}
