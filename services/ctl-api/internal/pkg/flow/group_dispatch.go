package flow

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstepgroup"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// DispatchGroupSignal enqueues an execute-workflow-step-group signal to the step
// queue and awaits its completion via the group-finished update handler. Returns
// the group's directive directly so the caller doesn't need to re-fetch from DB.
func DispatchGroupSignal(ctx workflow.Context, cfg StepConfig, group *app.WorkflowStepGroup, flw *app.Workflow) (string, error) {
	logger := workflow.GetLogger(ctx)

	sig := &executeworkflowstepgroup.Signal{
		WorkflowID:      flw.ID,
		WorkflowType:    string(flw.Type),
		StepGroupID:     group.ID,
		GroupIdx:        group.GroupIdx,
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		QueueName:       cfg.QueueName,
		TargetQueueName: cfg.TargetQueueName,
		Parallel:        group.Parallel,
	}

	// Use the step group ID as the signal owner for direct group lookup during retries.
	signalOwnerID := group.ID
	signalOwnerType := "workflow_step_groups"
	if signalOwnerID == "" {
		// Backward compat: if no group ID, use the workflow ID.
		signalOwnerID = flw.ID
		signalOwnerType = "install_workflows"
	}

	logger.Info("dispatching group signal",
		"group_idx", group.GroupIdx,
		"step_group_id", group.ID,
		"workflow_id", flw.ID,
		"parallel", group.Parallel,
		"queue", cfg.QueueName,
	)

	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		QueueName:       cfg.QueueName,
		Signal:          sig,
		SignalOwnerID:   signalOwnerID,
		SignalOwnerType: signalOwnerType,
	})
	if err != nil {
		return "", errors.Wrapf(err, "unable to enqueue group signal for group %d", group.GroupIdx)
	}

	var awaitOpts []*workflow.ActivityOptions
	if t, ok := signal.Signal(sig).(signal.SignalWithTimeout); ok && t.Timeout() > 0 {
		awaitOpts = append(awaitOpts, &workflow.ActivityOptions{
			ScheduleToCloseTimeout: t.Timeout(),
		})
	}

	// Wait for the group to finish via the group-finished update handler.
	// If the group has a StepGroupID, use ForwardGroupFinished for resilient
	// completion tracking. Fall back to AwaitAwaitSignal for backward compat
	// when no StepGroupID is available (legacy in-flight workflows).
	if group.ID != "" {
		resp, err := workflowactivities.AwaitForwardGroupFinished(ctx, workflowactivities.ForwardGroupFinishedRequest{
			StepGroupID: group.ID,
		}, awaitOpts...)
		if err != nil {
			if ctx.Err() != nil {
				cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
				defer cancelCtxCancel()
				client.AwaitCancelSignal(cancelCtx, enqueueResp.QueueSignalID)
			}
			return "", errors.Wrapf(err, "group signal failed for group %d", group.GroupIdx)
		}
		return resp.Directive, nil
	}

	// Legacy fallback: no StepGroupID, use framework-level finished handler.
	_, err = client.AwaitAwaitSignal(ctx, enqueueResp.QueueSignalID, awaitOpts...)
	if err != nil {
		if ctx.Err() != nil {
			cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer cancelCtxCancel()
			client.AwaitCancelSignal(cancelCtx, enqueueResp.QueueSignalID)
		}
		return "", errors.Wrapf(err, "group signal failed for group %d", group.GroupIdx)
	}

	return "", nil
}

// DispatchGroupSignalByIdx is a backward-compatible helper that dispatches a group
// signal using a GroupIdx and parallel flag directly, without a WorkflowStepGroup object.
func DispatchGroupSignalByIdx(ctx workflow.Context, cfg StepConfig, groupIdx int, flw *app.Workflow, parallel bool) (string, error) {
	return DispatchGroupSignal(ctx, cfg, &app.WorkflowStepGroup{
		GroupIdx: groupIdx,
		Parallel: parallel,
	}, flw)
}
