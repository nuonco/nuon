package executeworkflowstep

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// Execute runs the full workflow step lifecycle. This replicates the logic from
// WorkflowConductor.executeFlowStep but as a self-contained signal that fetches
// its own state from the database.
func (s *Signal) Execute(ctx workflow.Context) error {
	defer func() { s.finished = true }()

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow logger")
	}

	// Fetch step and workflow from the database
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, s.StepID)
	if err != nil {
		return errors.Wrap(err, "unable to get step")
	}

	flw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow")
	}

	// Check if step is in executable state
	if step.Status.Status != app.StatusPending && step.Status.Status != app.StatusNotAttempted && step.Status.Status != app.StatusQueued {
		l.Debug("step not in executable state, exiting",
			zap.String("step_id", step.ID),
			zap.String("step_status", string(step.Status.Status)),
			zap.String("workflow_id", flw.ID))
		return nil
	}

	defer func() {
		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepFinishedAtByID(ctx, step.ID); err != nil {
			l.Error("unable to update finished at", zap.Error(err))
		}
	}()

	// Update flow status to in-progress
	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flw.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "executing step " + step.Name,
			Metadata:               map[string]any{},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step")
	}

	// Execute the inner signal
	stepErr := s.executeInnerSignal(ctx, step)
	if stepErr != nil {
		// If the context was cancelled, Cancel() already handled status updates.
		if ctx.Err() != nil {
			return nil
		}
		return s.handleStepError(ctx, l, step, flw, stepErr)
	}

	if s.canceled {
		return nil
	}

	// Refetch the step after signal execution to gather new state (e.g. step target ID)
	step, err = activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get step")
	}

	// Non-approval steps: mark success and return
	if step.ExecutionType != app.WorkflowStepExecutionTypeApproval {
		l.Debug("step type non approval, step successful",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusSuccess,
			},
		}); err != nil {
			return errors.Wrap(err, "unable to mark step as success")
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: flw.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusInProgress,
				StatusHumanDescription: "finished executing step " + step.Name,
				Metadata: map[string]any{
					"step_idx": step.Idx,
					"status":   "ok",
				},
			},
		}); err != nil {
			return errors.Wrap(err, "unable to update flow status after step")
		}

		return nil
	}

	if s.canceled {
		return nil
	}

	// Approval steps: delegate to plan processing
	return s.processPlan(ctx, step, flw)
}

// executeInnerSignal handles the actual signal dispatch for a step.
// It updates step status, then enqueues the inner signal to the install-signals
// queue and awaits completion.
func (s *Signal) executeInnerSignal(ctx workflow.Context, step *app.WorkflowStep) error {
	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepStartedAtByID(ctx, step.ID); err != nil {
		return err
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status: app.StatusInProgress,
		},
	}); err != nil {
		return err
	}

	if step.ExecutionType == app.WorkflowStepExecutionTypeSkipped {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusSuccess,
			},
		}); err != nil {
			return err
		}
		return nil
	}

	if step.QueueSignal == nil {
		return nil
	}

	sig := step.QueueSignal.Signal
	if sig == nil {
		return nil
	}

	// Ensure the inner signal references this step's ID, not the original step's.
	// Cloned (retry) steps copy QueueSignal from the original, so the embedded
	// step ID fields (WorkflowStepID, FlowStepID, etc.) may be stale.
	signal.ApplyStepContext(sig, step.ID, s.WorkflowID)

	// Inject retry count so signals can branch on retry index or group generation.
	signal.ApplyRetryCount(sig, step.RetryIndex, step.GroupRetryIdx)

	logger := workflow.GetLogger(ctx)
	logger.Info("enqueuing signal to target queue",
		"step_name", step.Name,
		"step_id", step.ID,
		"owner_id", s.OwnerID,
		"owner_type", s.OwnerType,
		"target_queue", s.TargetQueueName,
	)

	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         s.OwnerID,
		OwnerType:       s.OwnerType,
		QueueName:       s.TargetQueueName,
		Signal:          sig,
		SignalOwnerID:   step.ID,
		SignalOwnerType: "install_workflow_steps",
	})
	if err != nil {
		return errors.Wrapf(err, "unable to enqueue signal for step %s", step.Name)
	}

	// Track the inner signal ID so Cancel() can propagate cancellation
	s.innerQueueSignalID = enqueueResp.QueueSignalID

	logger.Info("waiting for queue signal to complete",
		"step_name", step.Name,
		"queue_signal_id", enqueueResp.QueueSignalID,
	)

	_, err = client.AwaitAwaitSignal(ctx, enqueueResp.QueueSignalID)
	if err != nil {
		return errors.Wrapf(err, "queue signal execution failed for step %s", step.Name)
	}

	logger.Info("queue signal completed successfully", "step_name", step.Name)
	return nil
}
