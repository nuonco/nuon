package build

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	createdsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/created"
	componentsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	orgprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/provision"
	orgreprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/reprovision"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type Signal struct {
	ComponentID string `json:"component_id" validate:"required"`
	BuildID     string `json:"build_id" validate:"required"`
	SandboxMode bool   `json:"sandbox_mode"`
}

var (
	_ signal.Signal           = (*Signal)(nil)
	_ signal.SignalWithCancel = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.ComponentID == "" {
		return errors.New("component_id is required")
	}
	if s.BuildID == "" {
		return errors.New("build_id is required")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusPlanning, "creating build plan")

	l := workflow.GetLogger(ctx)
	l.Info("executing build")
	preflightOptions := workflow.ActivityOptions{
		RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 1},
	}
	if _, err := activities.AwaitGetBuildGitSourceByBuildID(ctx, s.BuildID, &preflightOptions); err != nil {
		l.Error(err.Error())
		s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, err.Error())
		return err
	}

	comp, err := activities.AwaitGetComponentByComponentID(ctx, s.ComponentID)
	if err != nil {
		s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, "unable to get component")
		return fmt.Errorf("unable to get component: %w", err)
	}

	// Ensure org is provisioned before building.
	if err := queueclient.EnsureQueueSignal(ctx, comp.OrgID, "orgs", orgprovision.SignalType, orgreprovision.SignalType); err != nil {
		s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, "org provision not ready")
		return fmt.Errorf("org provision not ready: %w", err)
	}
	if comp.Status != app.ComponentStatusActive {
		createdSignal := &createdsignal.Signal{ComponentID: s.ComponentID}
		if err := createdSignal.Execute(ctx); err != nil {
			s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, "unable to activate component")
			return fmt.Errorf("unable to activate component: %w", err)
		}
	}

	// A component is activated asynchronously by its `created` signal. During a
	// sync burst the build can be enqueued before that signal commits the active
	// status, so wait for activation before the worker validates the status.
	if err := queueclient.EnsureQueueSignal(ctx, comp.ID, "components", createdsignal.SignalType); err != nil {
		s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, "component activation not ready")
		return fmt.Errorf("component activation not ready: %w", err)
	}

	return componentsworker.AwaitBuild(ctx, componentsworker.BuildRequest{
		ID:          s.ComponentID,
		BuildID:     s.BuildID,
		SandboxMode: s.SandboxMode,
	})
}

// updateBuildStatus updates the build status
func (s *Signal) updateBuildStatus(ctx workflow.Context, bldID string, status app.ComponentBuildStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)
	err := activities.AwaitUpdateBuildStatus(ctx, activities.UpdateBuildStatus{
		BuildID:           bldID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err != nil {
		l.Error("unable to update build status",
			zap.String("build-id", bldID),
			zap.Error(err))
		return
	}

	err = statusactivities.AwaitUpdateBuildStatusV2(ctx, statusactivities.UpdateBuildStatusV2{
		BuildID:           bldID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err != nil {
		l.Error("unable to update build status v2",
			zap.String("build-id", bldID),
			zap.Error(err))
		return
	}
}

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	s.updateBuildStatus(cancelCtx, s.BuildID, app.ComponentBuildStatus(app.StatusCancelled), "build cancelled")
	return nil
}
