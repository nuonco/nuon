package flow

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstep"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// StepConfig holds the queue/owner configuration needed to dispatch and manage
// workflow steps. Used by both the execute-flow signal (directly) and the
// WorkflowConductor (for backward compatibility).
type StepConfig struct {
	GroupQueueName         string
	QueueName              string
	TargetQueueName        string
	GenerateStepsQueueName string
	OwnerID                string
	OwnerType              string
	MW                     tmetrics.Writer
}

// DispatchStepSignal enqueues the execute-workflow-step signal to the step queue.
// The signal runs the full step lifecycle in its own handler workflow.
func DispatchStepSignal(ctx workflow.Context, cfg StepConfig, step *app.WorkflowStep, flw *app.Workflow) error {
	stepStart := workflow.Now(ctx)
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

	// Create a callback so the handler signals us on completion.
	cb := callback.New(ctx, step.ID)

	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		QueueName:       cfg.QueueName,
		Signal:          sig,
		SignalOwnerID:   step.ID,
		SignalOwnerType: "install_workflow_steps",
		Callback:        cb,
	})
	if err != nil {
		return errors.Wrapf(err, "unable to enqueue execute-workflow-step signal for step %s", step.Name)
	}

	// Wait for completion via signal channel — zero activity overhead, zero heartbeats.
	_, err = callback.Await(ctx, cb)
	if err != nil {
		// If the parent workflow was cancelled, propagate cancellation to the step signal
		if ctx.Err() != nil {
			cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer cancelCtxCancel()
			client.AwaitCancelSignal(cancelCtx, enqueueResp.QueueSignalID)
		}
		if cfg.MW != nil {
			cfg.MW.Timing(ctx, "workflow.step.latency", workflow.Now(ctx).Sub(stepStart),
				"step_name", step.Name, "workflow_type", string(flw.Type), "status", "error")
		}
		return errors.Wrapf(err, "execute-workflow-step signal failed for step %s", step.Name)
	}

	if cfg.MW != nil {
		cfg.MW.Timing(ctx, "workflow.step.latency", workflow.Now(ctx).Sub(stepStart),
			"step_name", step.Name, "workflow_type", string(flw.Type), "status", "success")
	}

	return nil
}
