package updatecomponenttype

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type Signal struct {
	ComponentID   string            `json:"component_id" validate:"required"`
	ComponentType app.ComponentType `json:"component_type" validate:"required"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.ComponentID == "" {
		return errors.New("component_id is required")
	}
	if s.ComponentType == "" {
		return errors.New("component_type is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	return activities.AwaitUpdateComponentType(ctx, activities.UpdateComponentTypeRequest{
		ComponentID: s.ComponentID,
		Type:        s.ComponentType,
	})
}
