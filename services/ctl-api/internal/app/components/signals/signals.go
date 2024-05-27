package signals

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

const (
	TemporalNamespace string = "components"
)

const (
	OperationCreated          eventloop.SignalType = "created"
	OperationRestart          eventloop.SignalType = "restart"
	OperationBuild            eventloop.SignalType = "build"
	OperationQueueBuild       eventloop.SignalType = "queue_build"
	OperationProvision        eventloop.SignalType = "provision"
	OperationDelete           eventloop.SignalType = "delete"
	OperationPollDependencies eventloop.SignalType = "poll_dependencies"
	OperationConfigCreated    eventloop.SignalType = "config_created"
)

type Signal struct {
	Type eventloop.SignalType `validate:"required"`

	BuildID string `validate:"required_if=Operation build"`

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
	switch s.SignalType() {
	case OperationCreated:
		return true
	case OperationRestart:
		return true
	default:
	}

	return false
}

func (s *Signal) GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error) {
	org, err := org.FromContext(ctx)
	if err == nil {
		return org, nil
	}

	comp := app.Component{}
	res := db.WithContext(ctx).
		Preload("Org").
		First(&comp, "id = ?", id)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	return &comp.Org, nil
}
