package delete

import (
	"fmt"
	"slices"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
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
	s.updateStatus(ctx, s.ComponentID, app.ComponentStatusDeleteQueued, "delete: polling for component being unused by active installs and latest app config")
	if err := s.pollComponentBeingUnused(ctx, s.ComponentID); err != nil {
		return err
	}

	// update status
	s.updateStatus(ctx, s.ComponentID, app.ComponentStatusDeprovisioning, "deleting component")
	if err := activities.AwaitDelete(ctx, activities.DeleteRequest{
		ComponentID: s.ComponentID,
	}); err != nil {
		s.updateStatus(ctx, s.ComponentID, app.ComponentStatusError, "unable to delete component from database")
		return fmt.Errorf("unable to delete component: %w", err)
	}

	return nil
}

// pollComponentBeingUnused polls until the component is no longer used by any installs or app configs
func (s *Signal) pollComponentBeingUnused(ctx workflow.Context, compID string) error {
	deadline := workflow.Now(ctx).Add(time.Minute * 60)

	inFlight := true
	for inFlight {
		appCfg, err := activities.AwaitGetComponentAppConfigByComponentID(ctx, compID)
		if err != nil {
			s.updateStatus(ctx, compID, app.ComponentStatusError, "delete: unable to get component app config")
			return fmt.Errorf("unable to get component app config: %w", err)
		}

		inFlight = false
		if slices.Contains(appCfg.ComponentIDs, compID) {
			inFlight = true
		}

		depsInstalls, err := activities.AwaitGetComponentInstalls(ctx, activities.GetComponentInstallsRequest{
			ComponentID: compID,
			AppID:       appCfg.AppID,
		})
		if err != nil {
			s.updateStatus(ctx, compID, app.ComponentStatusError, "delete: unable to get component dependent installs")
			return fmt.Errorf("unable to get component dependent installs: %w", err)
		}

		if len(depsInstalls) > 0 {
			inFlight = true
		}

		if workflow.Now(ctx).After(deadline) {
			s.updateStatus(ctx, compID, app.ComponentStatusError, "delete: timed out waiting for dependent installs to remove the component")
			return fmt.Errorf("timeout waiting for dependent installs to remove the component")
		}

		workflow.Sleep(ctx, defaultPollTimeout)
	}

	return nil
}

// updateStatus is a helper method to update component status
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
