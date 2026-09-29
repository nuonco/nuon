package flow

import (
	"strconv"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstepgroup"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func DispatchGroupSignal(ctx workflow.Context, cfg StepConfig, group *app.WorkflowStepGroup, flw *app.Workflow) (string, error) {
	groupStart := workflow.Now(ctx)
	logger := workflow.GetLogger(ctx)

	sig := &executeworkflowstepgroup.Signal{
		WorkflowID:      flw.ID,
		StepGroupID:     group.ID,
		GroupIdx:        group.GroupIdx,
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		QueueName:       cfg.QueueName,
		TargetQueueName: cfg.TargetQueueName,
		Parallel:        group.Parallel,
		DerivedTimeout:  group.Timeout,
	}

	signalOwnerID := group.ID
	signalOwnerType := "workflow_step_groups"
	if signalOwnerID == "" {
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

	cb := callback.New(ctx, signalOwnerID)
	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		QueueName:       cfg.QueueName,
		Signal:          sig,
		SignalOwnerID:   signalOwnerID,
		SignalOwnerType: signalOwnerType,
		Callback:        cb,
	})
	if err != nil {
		return "", errors.Wrapf(err, "unable to enqueue group signal for group %d", group.GroupIdx)
	}

	timeoutOpts := signal.TimeoutActivityOpts(signal.DeriveTimeout(sig))

	groupTags := []string{"group_idx", strconv.Itoa(group.GroupIdx), "workflow_type", string(flw.Type)}

	if group.ID != "" && group.Timeout > 0 {
		resp, err := workflowactivities.AwaitForwardGroupFinished(ctx, workflowactivities.ForwardGroupFinishedRequest{
			StepGroupID: group.ID,
		}, timeoutOpts)
		if err != nil {
			if ctx.Err() != nil {
				cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
				defer cancelCtxCancel()
				client.AwaitCancelSignal(cancelCtx, enqueueResp.QueueSignalID)
			}
			if cfg.MW != nil {
				cfg.MW.Timing(ctx, "workflow.step_group.latency", workflow.Now(ctx).Sub(groupStart), append(groupTags, "status", "error")...)
			}
			return "", errors.Wrapf(err, "group signal failed for group %d", group.GroupIdx)
		}
		if cfg.MW != nil {
			cfg.MW.Timing(ctx, "workflow.step_group.latency", workflow.Now(ctx).Sub(groupStart), append(groupTags, "status", "success")...)
		}
		return resp.Directive, nil
	}

	groupTimeout := group.Timeout
	if groupTimeout == 0 {
		groupTimeout = callback.FallbackAwaitTimeout
	}
	_, err = callback.AwaitWithTimeout(ctx, cb, groupTimeout)
	if err != nil {
		if ctx.Err() != nil {
			cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer cancelCtxCancel()
			client.AwaitCancelSignal(cancelCtx, enqueueResp.QueueSignalID)
		}
		if cfg.MW != nil {
			cfg.MW.Timing(ctx, "workflow.step_group.latency", workflow.Now(ctx).Sub(groupStart), append(groupTags, "status", "error")...)
		}
		return "", errors.Wrapf(err, "group signal failed for group %d", group.GroupIdx)
	}

	if cfg.MW != nil {
		cfg.MW.Timing(ctx, "workflow.step_group.latency", workflow.Now(ctx).Sub(groupStart), append(groupTags, "status", "success")...)
	}

	return "", nil
}

func DispatchGroupSignalByIdx(ctx workflow.Context, cfg StepConfig, groupIdx int, flw *app.Workflow, parallel bool) (string, error) {
	return DispatchGroupSignal(ctx, cfg, &app.WorkflowStepGroup{
		GroupIdx: groupIdx,
		Parallel: parallel,
	}, flw)
}
