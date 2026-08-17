package updated

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "updated"

type Signal struct {
	signal.LifecycleBase

	InstallID string `json:"install_id"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		InstallID:    &s.InstallID,
		Operation:    "install-updated",
		WorkflowID:   s.LifecycleWorkflowID,
		WorkflowType: s.LifecycleWorkflowType,
	}
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}

	// Validate install exists
	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "install not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Mark state as stale (copied from worker/updated.go)
	if err := activities.AwaitMarkStateStale(ctx, &activities.MarkStateStaleRequest{
		InstallID:       s.InstallID,
		TriggeredByID:   s.InstallID,
		TriggeredByType: "installs",
	}); err != nil {
		if !generics.IsGormErrRecordNotFound(err) {
			return errors.Wrap(err, "unable to mark state as stale")
		}
	}

	// Dynamic labels are materialized from install state; nothing else regenerates
	// state on this path (input/install updates only mark it stale), so render here
	// or templated values sit stale until an unrelated deploy. The render reads
	// state through GetInstallState, which regenerates the partials the enqueueing
	// endpoint marked stale. Best-effort: label rendering must not fail the update.
	if err := activities.AwaitRenderInstallLabels(ctx, &activities.RenderInstallLabelsRequest{
		InstallID: s.InstallID,
	}); err != nil {
		workflow.GetLogger(ctx).Warn("unable to render install label templates",
			"install_id", s.InstallID, "error", err)
	}
	return nil
}
