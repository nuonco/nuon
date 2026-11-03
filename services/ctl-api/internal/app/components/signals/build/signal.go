package build

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	createdsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/created"
	componentsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	orgprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/provision"
	orgreprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/reprovision"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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
	err := s.preflight(ctx)
	if err != nil {
		return activities.FailBuild(ctx, s.BuildID, err)
	}

	err = componentsworker.AwaitBuild(ctx, componentsworker.BuildRequest{
		ID:          s.ComponentID,
		BuildID:     s.BuildID,
		SandboxMode: s.SandboxMode,
	})
	if err != nil && (temporal.IsTimeoutError(err) || temporal.IsTerminatedError(err)) {
		return activities.FailBuild(ctx, s.BuildID, err)
	}
	return err
}

func (s *Signal) preflight(ctx workflow.Context) error {
	if err := componentsworker.UpdateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusPlanning, "creating build plan"); err != nil {
		return err
	}

	l := workflow.GetLogger(ctx)
	l.Info("executing build")
	preflightOptions := workflow.ActivityOptions{
		RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 1},
	}
	if _, err := activities.AwaitGetBuildGitSourceByBuildID(ctx, s.BuildID, &preflightOptions); err != nil {
		l.Error(err.Error())
		return err
	}

	comp, err := activities.AwaitGetComponentByComponentID(ctx, s.ComponentID)
	if err != nil {
		return fmt.Errorf("unable to get component: %w", err)
	}

	// Ensure org is provisioned before building.
	if err := queueclient.EnsureQueueSignal(ctx, comp.OrgID, "orgs", orgprovision.SignalType, orgreprovision.SignalType); err != nil {
		return fmt.Errorf("org provision not ready: %w", err)
	}
	if comp.Status != app.ComponentStatusActive {
		createdSignal := &createdsignal.Signal{ComponentID: s.ComponentID}
		if err := createdSignal.Execute(ctx); err != nil {
			return fmt.Errorf("unable to activate component: %w", err)
		}
	}

	// A component is activated asynchronously by its `created` signal. During a
	// sync burst the build can be enqueued before that signal commits the active
	// status, so wait for activation before the worker validates the status.
	if err := queueclient.EnsureQueueSignal(ctx, comp.ID, "components", createdsignal.SignalType); err != nil {
		return fmt.Errorf("component activation not ready: %w", err)
	}

	return nil
}

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	return componentsworker.UpdateBuildStatus(cancelCtx, s.BuildID, app.ComponentBuildStatus(app.StatusCancelled), "build cancelled")
}
