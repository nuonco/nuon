package delete

import (
	"fmt"
	"slices"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/poll"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type Signal struct {
	ComponentID string `json:"component_id" validate:"required"`

	v *validator.Validate
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithParams = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) WithParams(p *signal.Params) {
	s.v = p.V
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.ComponentID == "" {
		return errors.New("component_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	s.updateStatus(ctx, s.ComponentID, app.ComponentStatusDeleteQueued, "delete: polling for component being unused by active installs and latest app config")
	if err := s.pollComponentBeingUnused(ctx, s.ComponentID); err != nil {
		return err
	}

	s.updateStatus(ctx, s.ComponentID, app.ComponentStatusDeprovisioning, "deleting component")
	if err := activities.AwaitDelete(ctx, activities.DeleteRequest{
		ComponentID: s.ComponentID,
	}); err != nil {
		s.updateStatus(ctx, s.ComponentID, app.ComponentStatusError, "unable to delete component from database")
		return fmt.Errorf("unable to delete component: %w", err)
	}

	return nil
}

func (s *Signal) pollComponentBeingUnused(ctx workflow.Context, compID string) error {
	deadline := workflow.Now(ctx).Add(time.Minute * 60)

	return poll.Poll(ctx, s.v, poll.PollOpts{
		MaxTS:           deadline,
		InitialInterval: 10 * time.Second,
		MaxInterval:     5 * time.Minute,
		BackoffFactor:   2,
		Fn: func(ctx workflow.Context) error {
			appCfg, err := activities.AwaitGetComponentAppConfigByComponentID(ctx, compID)
			if err != nil {
				s.updateStatus(ctx, compID, app.ComponentStatusError, "delete: unable to get component app config")
				return errors.Wrap(poll.NonRetryableError, "unable to get component app config")
			}

			if slices.Contains(appCfg.ComponentIDs, compID) {
				return errors.New("component still in latest app config")
			}

			depsInstalls, err := activities.AwaitGetComponentInstalls(ctx, activities.GetComponentInstallsRequest{
				ComponentID: compID,
				AppID:       appCfg.AppID,
			})
			if err != nil {
				s.updateStatus(ctx, compID, app.ComponentStatusError, "delete: unable to get component dependent installs")
				return errors.Wrap(poll.NonRetryableError, "unable to get component dependent installs")
			}

			if len(depsInstalls) > 0 {
				return errors.New("component still used by installs")
			}

			return nil
		},
	})
}

func (s *Signal) updateStatus(ctx workflow.Context, compID string, status app.ComponentStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)
	err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		ComponentID:       compID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err != nil {
		l.Error("unable to update component status",
			zap.String("component-id", compID),
			zap.Error(err))
		return
	}

	err = statusactivities.AwaitUpdateComponentStatusV2(ctx, statusactivities.UpdateComponentStatusV2Request{
		ComponentID:       compID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err != nil {
		l.Error("unable to update component status v2",
			zap.String("component-id", compID),
			zap.Error(err))
		return
	}
}
