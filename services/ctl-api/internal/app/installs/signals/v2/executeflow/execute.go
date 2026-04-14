package executeflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	v2workflows "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/workflows/v2"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
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

// executeFlow runs the workflow conductor with v2 generators and queue-based execution.
// It handles ContinueAsNewErr by looping internally since the signal Execute() runs
// inside a queue handler workflow and cannot use Temporal's ContinueAsNew directly.
//
// After a step failure, if the workflow is retryable, the signal stays alive and waits
// for a "retry-step" update handler to be called. This enables reactive retries without
// creating new rerun-flow signals.
func (s *Signal) executeFlow(ctx workflow.Context) error {
	eventLoopReq := eventloop.EventLoopRequest{
		ID: s.installID,
	}

	fc := s.newConductor()

	// Initial execution
	startFromStepIdx := 0
	for {
		err := fc.Handle(ctx, eventLoopReq, s.InstallWorkflowID, startFromStepIdx)
		if err == nil {
			return nil
		}

		cerr, ok := err.(*flow.ContinueAsNewErr)
		if ok && cerr != nil {
			startFromStepIdx = cerr.StartFromStepIdx
			continue
		}

		// Step failed - check if workflow is retryable
		if !s.checkRetryable(ctx) {
			return err
		}

		// Mark workflow as failed, awaiting retry
		_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: s.InstallWorkflowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: "workflow failed, awaiting retry",
				Metadata: map[string]any{
					"error_message":  err.Error(),
					"awaiting_retry": true,
				},
			},
		})

		// Wait for retry-step update
		if err := workflow.Await(ctx, func() bool {
			return s.retryRequested
		}); err != nil {
			return err
		}

		// Retry requested - execute rerun from the failed step
		s.retryRequested = false
		rerunErr := s.retry(ctx)
		if rerunErr == nil {
			return nil
		}

		// Rerun also failed - loop back to check retryable again
	}
}

// checkRetryable checks if the workflow is still eligible for retry.
func (s *Signal) checkRetryable(ctx workflow.Context) bool {
	resp, err := activities.AwaitCheckWorkflowRetryable(ctx, activities.CheckWorkflowRetryableRequest{
		WorkflowID: s.InstallWorkflowID,
	})
	if err != nil {
		return false
	}
	return resp.Retryable
}

// retry executes a rerun of the workflow from the failed step.
func (s *Signal) retry(ctx workflow.Context) error {
	fc := s.newConductor()
	continueFromIdx := 0
	for {
		err := fc.Rerun(ctx, eventloop.EventLoopRequest{ID: s.installID}, flow.RerunInput{
			ContinueFromIdx: continueFromIdx,
			FlowID:          s.InstallWorkflowID,
			StepID:          s.retryStepID,
			Operation:       flow.RerunOperation(s.retryOperation),
		})
		if err == nil {
			return nil
		}
		cerr, ok := err.(*flow.ContinueAsNewErr)
		if ok && cerr != nil {
			continueFromIdx = cerr.StartFromStepIdx
			continue
		}
		return err
	}
}
