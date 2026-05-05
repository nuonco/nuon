package driftchecksandbox

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

// SignalType is the queue signal type for triggering a v2 sandbox drift scan.
//
// Mirrors the per-component drift-check signal in signals/v2/driftcheck. This
// one is install-scoped: each install gets a single sandbox drift cron emitter
// (vs the per-component drift cron). It creates a drift_run_reprovision_sandbox
// workflow with PlanOnly:true and hands it to the install-workflows queue —
// the same plan-only execution path used by manual sandbox drift scans, so the
// drift-detected notification fires from a single hook point.
const SignalType signal.SignalType = "drift-check-sandbox"

type Signal struct {
	InstallID string `json:"install_id"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		InstallID: &s.InstallID,
		Operation: "drift-check-sandbox",
	}
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}

	if _, err := activities.AwaitGetByInstallID(ctx, s.InstallID); err != nil {
		return fmt.Errorf("install not found: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	wkflw, err := activities.AwaitCreateWorkflow(ctx, activities.CreateWorkflowRequest{
		InstallID:    s.InstallID,
		WorkflowType: app.WorkflowTypeDriftRunReprovisionSandbox,
		PlanOnly:     true,
		Metadata:     map[string]string{},
	})
	if err != nil {
		return fmt.Errorf("unable to create sandbox drift workflow: %w", err)
	}

	if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   s.InstallID,
		OwnerType: "installs",
		QueueName: helpers.InstallWorkflowsQueueName,
		Signal: &executeflow.Signal{
			WorkflowID: wkflw.ID,
		},
	}); err != nil {
		return fmt.Errorf("unable to enqueue flow execution signal: %w", err)
	}

	return nil
}
