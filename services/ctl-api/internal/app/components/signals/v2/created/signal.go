package created

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type Signal struct {
	ComponentID string `json:"component_id" validate:"required"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.ComponentID == "" {
		return errors.New("component_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	s.updateStatus(ctx, s.ComponentID, app.ComponentStatusActive, "component is active")
	return nil
}

// updateStatus is a helper method to update component status
func (s *Signal) updateStatus(ctx workflow.Context, compID string, status app.ComponentStatus, statusDescription string) {
	err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		ComponentID:       compID,
		Status:            status,
		StatusDescription: statusDescription,
	})

	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update component status",
		zap.String("component-id", compID),
		zap.Error(err))
}
