package flow

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstep"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// StepConfig holds the queue/owner configuration needed to dispatch and manage
// workflow steps. Used by both the execute-flow signal (directly) and the
// WorkflowConductor (for backward compatibility).
type StepConfig struct {
	GroupQueueName  string
	QueueName       string
	TargetQueueName string
	OwnerID         string
	OwnerType       string
}

// DispatchStepSignal enqueues the execute-workflow-step signal to the step queue.
// The signal runs the full step lifecycle in its own handler workflow.
func DispatchStepSignal(ctx workflow.Context, cfg StepConfig, step *app.WorkflowStep, flw *app.Workflow) error {
	logger := workflow.GetLogger(ctx)

	sig := &executeworkflowstep.Signal{
		StepID:          step.ID,
		WorkflowID:      flw.ID,
		WorkflowType:    string(flw.Type),
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		TargetQueueName: cfg.TargetQueueName,
		// Forward stamped names so workflow_step lifecycle webhook events
		// carry human-readable identifiers without a per-event DB lookup.
		// These come from GetFlow's preload (Org) + polymorphic owner-name
		// lookup (installs/apps/app_branches).
		OrgID:     flw.OrgID,
		OrgName:   flw.Org.Name,
		OwnerName: flw.OwnerName,
	}

	logger.Info("enqueuing execute-workflow-step signal to step queue",
		"step_id", step.ID,
		"step_name", step.Name,
		"step_queue", cfg.QueueName,
		"target_queue", cfg.TargetQueueName,
		"owner_id", cfg.OwnerID,
	)

	// Mark step as queued so it's visible to users while waiting in the queue
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status: app.StatusQueued,
		},
	}); err != nil {
		return errors.Wrapf(err, "unable to mark step %s as queued", step.Name)
	}

	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		QueueName:       cfg.QueueName,
		Signal:          sig,
		SignalOwnerID:   step.ID,
		SignalOwnerType: "install_workflow_steps",
	})
	if err != nil {
		return errors.Wrapf(err, "unable to enqueue execute-workflow-step signal for step %s", step.Name)
	}

	_, err = client.AwaitAwaitSignal(ctx, enqueueResp.QueueSignalID)
	if err != nil {
		// If the parent workflow was cancelled, propagate cancellation to the step signal
		if ctx.Err() != nil {
			cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer cancelCtxCancel()
			client.AwaitCancelSignal(cancelCtx, enqueueResp.QueueSignalID)
		}
		return errors.Wrapf(err, "execute-workflow-step signal failed for step %s", step.Name)
	}

	return nil
}
