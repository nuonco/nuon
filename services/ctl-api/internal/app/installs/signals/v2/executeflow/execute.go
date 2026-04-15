package executeflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	v2workflows "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/workflows/v2"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func getWorkflowStepGenerators() map[app.WorkflowType]flow.WorkflowStepGenerator {
	return map[app.WorkflowType]flow.WorkflowStepGenerator{
		app.WorkflowTypeManualDeploy:               v2workflows.ManualDeploySteps,
		app.WorkflowTypeDriftRun:                   v2workflows.ManualDeploySteps,
		app.WorkflowTypeDeployComponents:           v2workflows.DeployAllComponents,
		app.WorkflowTypeTeardownComponent:          v2workflows.TeardownComponent,
		app.WorkflowTypeTeardownComponents:         v2workflows.TeardownComponents,
		app.WorkflowTypeInputUpdate:                v2workflows.InputUpdate,
		app.WorkflowTypeActionWorkflowRun:          v2workflows.RunActionWorkflow,
		app.WorkflowTypeProvision:                  v2workflows.Provision,
		app.WorkflowTypeReprovision:                v2workflows.Reprovision,
		app.WorkflowTypeReprovisionSandbox:         v2workflows.ReprovisionSandbox,
		app.WorkflowTypeDriftRunReprovisionSandbox: v2workflows.ReprovisionSandbox,
		app.WorkflowTypeDeprovision:                v2workflows.Deprovision,
		app.WorkflowTypeDeprovisionSandbox:         v2workflows.DeprovisionSandbox,
		app.WorkflowTypeSyncSecrets:                v2workflows.SyncSecrets,
	}
}

func (s *Signal) newConductor() *flow.WorkflowConductor[*signals.Signal] {
	return &flow.WorkflowConductor[*signals.Signal]{
		Generators:          getWorkflowStepGenerators(),
		StepChildWorkflow:   true,
		StepQueueName:       "install-workflow-steps",
		StepTargetQueueName: "install-signals",
		StepOwnerID:         s.installID,
		StepOwnerType:       "installs",
	}
}

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
				ID: s.InstallWorkflowID,
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

// executeRun executes a single workflow run, handling ContinueAsNew internally.
// Returns nil when the run completes (either all steps done or paused at approval).
// Returns an error for actual execution failures.
func (s *Signal) executeRun(ctx workflow.Context, run *app.WorkflowRun) error {
	fc := s.newConductor()
	eventLoopReq := eventloop.EventLoopRequest{ID: s.installID}
	startIdx := run.StartFromIdx

	for {
		err := fc.Handle(ctx, eventLoopReq, s.InstallWorkflowID, startIdx)
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
		return err
	}
}

// createRun creates a WorkflowRun record to track this execution segment.
func (s *Signal) createRun(ctx workflow.Context, runType app.WorkflowRunType, triggerStepID string, startFromIdx int) (*app.WorkflowRun, error) {
	return workflowactivities.AwaitPkgWorkflowsFlowCreateWorkflowRun(ctx, workflowactivities.CreateWorkflowRunRequest{
		WorkflowID:    s.InstallWorkflowID,
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
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, s.InstallWorkflowID)
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
	resp, err := installactivities.AwaitCheckWorkflowRetryable(ctx, installactivities.CheckWorkflowRetryableRequest{
		WorkflowID: s.InstallWorkflowID,
	})
	if err != nil {
		return false
	}
	return resp.Retryable
}
