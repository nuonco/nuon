package executeflow

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// executeFlow runs the workflow conductor with run-based execution.
// Each execution segment (initial, retry, skip, resume) is tracked as a WorkflowRun.
// The flow pauses at approval points and errors, waiting for update handlers to resume.
func (s *Signal) executeFlow(ctx workflow.Context) error {
	// Create and execute the initial run
	run, err := s.createRun(ctx, app.WorkflowRunTypeInitial, "", 0)
	if err != nil {
		return err
	}

	for {
		runErr := s.executeRun(ctx, run)

		if runErr == nil {
			// Run completed without error. Check if workflow is fully done
			// or if we paused at an approval/directive point.
			if s.isWorkflowComplete(ctx) {
				s.updateRunStatus(ctx, run.ID, app.StatusSuccess)
				return nil
			}
			// Paused at approval - update run status and wait for resume
			s.updateRunStatus(ctx, run.ID, app.AwaitingApproval)
		} else {
			// Actual execution error
			s.updateRunStatus(ctx, run.ID, app.StatusError)

			if !s.checkRetryable(ctx) {
				return runErr
			}

			// Mark workflow as failed, awaiting retry
			_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: s.WorkflowID,
				Status: app.CompositeStatus{
					Status:                 app.StatusError,
					StatusHumanDescription: "workflow failed, awaiting retry",
					Metadata: map[string]any{
						"error_message":  runErr.Error(),
						"awaiting_retry": true,
					},
				},
			})
		}

		// Wait for a resume or cancel signal from an update handler
		if err := workflow.Await(ctx, func() bool {
			return s.resumeRequested || s.cancelRequested
		}); err != nil {
			return err
		}

		if s.cancelRequested {
			return runErr
		}

		// Create a new run for the resume
		s.resumeRequested = false
		run, err = s.createRun(ctx, s.resumeRunType, s.resumeStepID, s.resumeStartIdx)
		if err != nil {
			return err
		}
	}
}

// executeRun executes a single workflow run, directly managing step generation
// and execution without going through the WorkflowConductor.
func (s *Signal) executeRun(ctx workflow.Context, run *app.WorkflowRun) error {
	cfg := s.stepConfig()
	startIdx := run.StartFromIdx

	for {
		err := s.handle(ctx, startIdx)
		if err == nil {
			return nil
		}

		// Handle ContinueAsNew (batch size limit)
		if cerr, ok := err.(*flow.ContinueAsNewErr); ok && cerr != nil {
			startIdx = cerr.StartFromStepIdx
			continue
		}

		// ApprovalPauseErr means we stopped at an approval - return nil to enter wait loop
		if _, ok := err.(*flow.ApprovalPauseErr); ok {
			return nil
		}

		// Actual failure
		_ = cfg // suppress unused warning in case of early return refactors
		return err
	}
}

// handle is the inlined equivalent of WorkflowConductor.Handle(). It manages
// the full lifecycle of a flow execution: generate steps, execute steps, update status.
func (s *Signal) handle(ctx workflow.Context, startFromStepIdx int) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil
	}

	flw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow object")
	}
	if flw.Status.Status == app.StatusCancelled {
		return errors.New("workflow already cancelled")
	}

	defer func() {
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer cancelCtxCancel()

			if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(cancelCtx, statusactivities.UpdateStatusRequest{
				ID: s.WorkflowID,
				Status: app.CompositeStatus{
					Status: app.StatusCancelled,
				},
			}); err != nil {
				l.Error("unable to update status on cancellation", zap.Error(err))
			}
		}
	}()

	if startFromStepIdx == 0 {
		if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowStartedAtByID(ctx, s.WorkflowID); err != nil {
			return err
		}
	}

	// Generate steps
	l.Debug("generating steps for workflow")
	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.WorkflowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "generating steps for workflow",
		},
	}); err != nil {
		return err
	}

	cfg := s.stepConfig()
	flw, err = flow.GenerateSteps(ctx, cfg, flw, nil)
	if err != nil {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: s.WorkflowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: "error while generating steps",
				Metadata: map[string]any{
					"error_message": err.Error(),
				},
			},
		}); err != nil {
			return err
		}

		return errors.Wrap(err, "unable to generate workflow steps")
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.WorkflowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "successfully generated all steps",
		},
	}); err != nil {
		return err
	}

	// Execute steps
	l.Debug("executing steps for workflow")
	err = flow.ExecuteStepsViaChildWorkflow(ctx, cfg, flw, startFromStepIdx)
	if err != nil {
		if _, ok := err.(*flow.ContinueAsNewErr); ok {
			return err
		}
	}

	if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, s.WorkflowID); err != nil {
		l.Error("unable to update finished at", zap.Error(err))
	}

	if err != nil {
		status := app.CompositeStatus{
			Status:                 app.StatusError,
			StatusHumanDescription: "error while executing steps",
			Metadata: map[string]any{
				"error_message": err.Error(),
			},
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID:     s.WorkflowID,
			Status: status,
		}); err != nil {
			return err
		}

		return errors.Wrap(err, "unable to execute workflow steps")
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.WorkflowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusSuccess,
			StatusHumanDescription: "successfully executed workflow",
		},
	}); err != nil {
		return err
	}

	return nil
}

// stepConfig returns the StepConfig for this signal.
func (s *Signal) stepConfig() flow.StepConfig {
	return flow.StepConfig{
		QueueName:       s.StepQueueName,
		TargetQueueName: s.StepTargetQueueName,
		OwnerID:         s.OwnerID,
		OwnerType:       s.OwnerType,
	}
}

// createRun creates a WorkflowRun record to track this execution segment.
func (s *Signal) createRun(ctx workflow.Context, runType app.WorkflowRunType, triggerStepID string, startFromIdx int) (*app.WorkflowRun, error) {
	return workflowactivities.AwaitPkgWorkflowsFlowCreateWorkflowRun(ctx, workflowactivities.CreateWorkflowRunRequest{
		WorkflowID:    s.WorkflowID,
		Type:          runType,
		TriggerStepID: triggerStepID,
		StartFromIdx:  startFromIdx,
	})
}

// updateRunStatus updates the status of a workflow run.
func (s *Signal) updateRunStatus(ctx workflow.Context, runID string, status app.Status) {
	workflowactivities.AwaitPkgWorkflowsFlowUpdateWorkflowRunStatus(ctx, workflowactivities.UpdateWorkflowRunStatusRequest{
		RunID: runID,
		Status: app.CompositeStatus{
			Status: status,
		},
	})
}

// isWorkflowComplete checks if all steps in the workflow have terminal statuses.
func (s *Signal) isWorkflowComplete(ctx workflow.Context) bool {
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, s.WorkflowID)
	if err != nil {
		return false
	}

	for _, step := range steps {
		switch step.Status.Status {
		case app.StatusSuccess, app.StatusAutoSkipped, app.StatusUserSkipped,
			app.StatusDiscarded, app.StatusCancelled,
			app.WorkflowStepApprovalStatusApproved,
			app.WorkflowStepNoDrift, app.WorkflowStepDrifted:
			continue
		default:
			return false
		}
	}

	return true
}

// checkRetryable checks if the workflow is still eligible for retry.
func (s *Signal) checkRetryable(ctx workflow.Context) bool {
	resp, err := workflowactivities.AwaitCheckWorkflowRetryable(ctx, workflowactivities.CheckWorkflowRetryableRequest{
		WorkflowID: s.WorkflowID,
	})
	if err != nil {
		return false
	}
	return resp.Retryable
}
