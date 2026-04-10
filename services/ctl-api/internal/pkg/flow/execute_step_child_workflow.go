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

// dispatchFlowStepSignal enqueues the execute-workflow-step signal to the step queue.
// The signal runs the full executeFlowStep lifecycle in its own handler workflow,
// including inner signal dispatch, approval handling, policy checks, etc.
func (c *WorkflowConductor[DomainSignal]) dispatchFlowStepSignal(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
	logger := workflow.GetLogger(ctx)

	sig := &executeworkflowstep.Signal{
		StepID:          step.ID,
		WorkflowID:      flw.ID,
		OwnerID:         c.StepOwnerID,
		OwnerType:       c.StepOwnerType,
		TargetQueueName: c.StepTargetQueueName,
	}

	logger.Info("enqueuing execute-workflow-step signal to step queue",
		"step_id", step.ID,
		"step_name", step.Name,
		"step_queue", c.StepQueueName,
		"target_queue", c.StepTargetQueueName,
		"owner_id", c.StepOwnerID,
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
		OwnerID:         c.StepOwnerID,
		OwnerType:       c.StepOwnerType,
		QueueName:       c.StepQueueName,
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
