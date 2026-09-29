package executeflow

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstepgroup"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func (s *Signal) executeGroup(ctx workflow.Context, group *app.WorkflowStepGroup, flw *app.Workflow) (string, error) {
	logger := workflow.GetLogger(ctx)
	cfg := s.stepConfig()

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
		ResidentFlow:    s.Resident,
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

	s.groupDispatchSeq++
	cb := callback.NewAttempt(ctx, signalOwnerID, s.groupDispatchSeq)
	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		QueueName:       cfg.GroupQueueName,
		Signal:          sig,
		SignalOwnerID:   signalOwnerID,
		SignalOwnerType: signalOwnerType,
		Callback:        cb,
	})
	if err != nil {
		return "", errors.Wrapf(err, "unable to enqueue group signal for group %d", group.GroupIdx)
	}

	s.activeGroupQueueSignalID = enqueueResp.QueueSignalID
	defer func() { s.activeGroupQueueSignalID = "" }()

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
		return "", errors.Wrapf(err, "group signal failed for group %d", group.GroupIdx)
	}

	if group.ID != "" {
		updatedGroup, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepGroupByID(ctx, group.ID)
		if err != nil {
			return "", errors.Wrap(err, "unable to re-fetch step group after group")
		}
		return updatedGroup.ResultDirective, nil
	}

	updatedFlw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, flw.ID)
	if err != nil {
		return "", errors.Wrap(err, "unable to re-fetch workflow after group")
	}
	return updatedFlw.ResultDirective, nil
}
