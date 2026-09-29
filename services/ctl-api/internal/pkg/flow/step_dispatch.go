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

type StepConfig struct {
	GroupQueueName         string
	QueueName              string
	TargetQueueName        string
	GenerateStepsQueueName string
	OwnerID                string
	OwnerType              string
	MW                     tmetrics.Writer
}

func DispatchStepSignal(ctx workflow.Context, cfg StepConfig, step *app.WorkflowStep, flw *app.Workflow) error {
	stepStart := workflow.Now(ctx)
	logger := workflow.GetLogger(ctx)

	sig := &executeworkflowstep.Signal{
		StepID:          step.ID,
		StepName:        step.Name,
		StepIdx:         step.Idx,
		StepGroupID:     step.WorkflowStepGroupID,
		GroupIdx:        step.GroupIdx,
		GroupRetryIdx:   step.GroupRetryIdx,
		RetryIndex:      step.RetryIndex,
		WorkflowID:      flw.ID,
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		TargetQueueName: cfg.TargetQueueName,
	}

	logger.Info("enqueuing execute-workflow-step signal to step queue",
		"step_id", step.ID,
		"step_name", step.Name,
		"step_queue", cfg.QueueName,
		"target_queue", cfg.TargetQueueName,
		"owner_id", cfg.OwnerID,
	)

	cb := callback.New(ctx, step.ID)
	enqueueReq := &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		QueueName:       cfg.QueueName,
		Signal:          sig,
		SignalOwnerID:   step.ID,
		SignalOwnerType: "install_workflow_steps",
		Callback:        cb,
	}

	markQueued := func() error {
		return statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusQueued,
			},
		})
	}

	var enqueueResp *sharedactivities.EnqueueSignalToOwnerResponse
	if executeworkflowstep.EnqueueBeforeQueued(ctx) {
		var err error
		enqueueResp, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, enqueueReq)
		if err != nil {
			_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID:     step.ID,
				Status: executeworkflowstep.DispatchFailedStatus(err),
			})
			return errors.Wrapf(err, "unable to enqueue execute-workflow-step signal for step %s", step.Name)
		}
		if err := markQueued(); err != nil {
			logger.Warn("step execute signal enqueued but queued status write failed",
				"step_id", step.ID,
				"error", err)
		}
	} else {
		if err := markQueued(); err != nil {
			return errors.Wrapf(err, "unable to mark step %s as queued", step.Name)
		}
		var err error
		enqueueResp, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, enqueueReq)
		if err != nil {
			return errors.Wrapf(err, "unable to enqueue execute-workflow-step signal for step %s", step.Name)
		}
	}

	stepTimeout := step.Timeout
	if stepTimeout == 0 {
		stepTimeout = callback.FallbackAwaitTimeout
	}
	_, err := callback.AwaitWithTimeout(ctx, cb, stepTimeout)
	if err != nil {
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
