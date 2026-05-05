package executeworkflowstepgroup

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstep"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// Execute runs all steps in this group, either sequentially or in parallel.
func (s *Signal) Execute(ctx workflow.Context) error {
	defer func() { s.finished = true }()

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow logger")
	}

	// Update group status to in-progress.
	s.updateGroupStatus(ctx, app.CompositeStatus{
		Status:                 app.StatusInProgress,
		StatusHumanDescription: "executing group steps",
	})

	var execErr error
	if s.Parallel {
		execErr = s.executeParallel(ctx, l)
	} else {
		execErr = s.executeSequential(ctx, l)
	}

	// Update group status based on outcome.
	if execErr != nil {
		if s.cancelRequested {
			s.updateGroupStatus(ctx, app.CompositeStatus{
				Status:                 app.StatusCancelled,
				StatusHumanDescription: "group cancelled",
			})
		} else {
			s.updateGroupStatus(ctx, app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: "group execution failed",
				Metadata: map[string]any{
					"error_message": execErr.Error(),
				},
			})
		}
	} else if s.lastDirective == DirectiveStop {
		s.updateGroupStatus(ctx, app.CompositeStatus{
			Status:                 app.StatusError,
			StatusHumanDescription: "group stopped",
			Metadata: map[string]any{
				"directive": DirectiveStop,
			},
		})
	} else {
		s.updateGroupStatus(ctx, app.CompositeStatus{
			Status:                 app.StatusSuccess,
			StatusHumanDescription: "group completed successfully",
		})
	}

	return execErr
}

// updateGroupStatus updates the group's composite status if a StepGroupID is set.
func (s *Signal) updateGroupStatus(ctx workflow.Context, status app.CompositeStatus) {
	if s.StepGroupID == "" {
		return
	}
	statusactivities.AwaitPkgStatusUpdateFlowStepGroupStatus(ctx, statusactivities.UpdateStatusRequest{
		ID:     s.StepGroupID,
		Status: status,
	})
}

// executeParallel dispatches all steps concurrently using executeSingleStep.
// Each step's full lifecycle (dispatch, await, auto-retry, user action) runs
// in its own goroutine. The group collects results and resolves directives.
func (s *Signal) executeParallel(ctx workflow.Context, l *zap.Logger) error {
	steps, err := s.getGroupSteps(ctx)
	if err != nil {
		return err
	}

	if len(steps) == 0 {
		return s.writeStepGroupDirective(ctx, DirectiveContinue)
	}

	l.Debug("dispatching steps in parallel",
		zap.Int("group_idx", s.GroupIdx),
		zap.Int("step_count", len(steps)))

	resultCh := workflow.NewChannel(ctx)

	for i := range steps {
		step := steps[i]
		workflow.Go(ctx, func(gCtx workflow.Context) {
			result := s.executeSingleStep(gCtx, l, &step)
			resultCh.Send(gCtx, result)
		})
	}

	var firstErr error
	hasStop := false
	hasRetryGroup := false

	for range steps {
		var result StepResult
		resultCh.Receive(ctx, &result)
		if result.Error != nil && firstErr == nil {
			firstErr = result.Error
		}
		if result.Directive == DirectiveStop {
			hasStop = true
		}
		if result.Directive == DirectiveRetryGroup {
			hasRetryGroup = true
		}
	}

	if firstErr != nil {
		if ctx.Err() != nil {
			return s.writeStepGroupDirective(ctx, DirectiveStop)
		}
		return firstErr
	}

	if hasStop {
		return s.writeStepGroupDirective(ctx, DirectiveStop)
	}

	if hasRetryGroup {
		return s.writeStepGroupDirective(ctx, DirectiveRetryGroup)
	}

	return s.writeWorkflowDirective(ctx, DirectiveContinue)
}

// dispatchStep enqueues an execute-workflow-step signal and returns the queue signal ID.
func (s *Signal) dispatchStep(ctx workflow.Context, step *app.WorkflowStep) (string, error) {
	sig := &executeworkflowstep.Signal{
		StepID:          step.ID,
		WorkflowID:      s.WorkflowID,
		WorkflowType:    s.WorkflowType,
		OwnerID:         s.OwnerID,
		OwnerType:       s.OwnerType,
		TargetQueueName: s.TargetQueueName,
		TargetQueueID:   step.TargetQueueID,
		// Forward stamped names so workflow_step lifecycle webhook events
		// carry human-readable identifiers without a per-event DB lookup.
		OrgID:     s.OrgID,
		OrgName:   s.OrgName,
		OwnerName: s.OwnerName,
	}

	// Mark step as queued
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status: app.StatusQueued,
		},
	}); err != nil {
		return "", errors.Wrapf(err, "unable to mark step %s as queued", step.Name)
	}

	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         s.OwnerID,
		OwnerType:       s.OwnerType,
		QueueName:       s.QueueName,
		QueueID:         step.StepQueueID,
		Signal:          sig,
		SignalOwnerID:   step.ID,
		SignalOwnerType: "install_workflow_steps",
	})
	if err != nil {
		return "", errors.Wrapf(err, "unable to enqueue execute-workflow-step signal for step %s", step.Name)
	}

	return enqueueResp.QueueSignalID, nil
}

// nextExecutableStep finds the next step in the group that is pending, queued, or not-attempted.
func (s *Signal) nextExecutableStep(steps []app.WorkflowStep) (*app.WorkflowStep, bool) {
	for i := range steps {
		step := &steps[i]
		switch step.Status.Status {
		case app.StatusPending, app.StatusNotAttempted, app.StatusQueued:
			return step, true
		}
	}
	return nil, false
}

// cancelRemainingSteps marks all non-terminal steps after the given step with
// the provided status. Use StatusDiscarded for both stop and skip-group
// directives.
func (s *Signal) cancelRemainingSteps(ctx workflow.Context, l *zap.Logger, steps []app.WorkflowStep, afterStepID string, status app.Status) {
	pastTrigger := false
	for _, step := range steps {
		if step.ID == afterStepID {
			pastTrigger = true
			continue
		}
		if !pastTrigger || isTerminalStatus(step.Status.Status) {
			continue
		}
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: status,
				Metadata: map[string]any{
					"reason": fmt.Sprintf("group step %s triggered stop", afterStepID),
				},
			},
		}); err != nil {
			l.Warn("failed to cancel remaining step", zap.String("step_id", step.ID), zap.Error(err))
		}
	}
}

// handleStepDispatchError handles errors from step dispatch or await.
func (s *Signal) handleStepDispatchError(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, err error) error {
	l.Error("step dispatch error",
		zap.String("step_id", step.ID),
		zap.String("step_name", step.Name),
		zap.Error(err))
	return errors.Wrapf(err, "step %s failed", step.Name)
}

// handleCancellation handles workflow cancellation during step execution.
func (s *Signal) handleCancellation(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()

	l.Debug("handling cancellation during step execution",
		zap.String("step_id", step.ID))

	// Cancel all tracked step signals
	for _, qsID := range s.stepSignalIDs {
		client.AwaitCancelSignal(cancelCtx, qsID)
	}

	s.writeStepGroupDirective(cancelCtx, DirectiveStop)
	return errors.New("group cancelled")
}
