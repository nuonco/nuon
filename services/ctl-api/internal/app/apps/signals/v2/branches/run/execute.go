package run

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appsignals "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/checkchanges"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	// FETCH EVERYTHING FROM DB using run_id
	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	// Get related entities
	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, run.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	logger.Info("starting app branch run",
		"run_id", run.ID,
		"app_branch_id", branch.ID,
		"app_branch_name", branch.Name,
		"workflow_id", *run.WorkflowID,
		"force", run.Force,
	)

	// Update status to running
	_, err = activities.AwaitUpdateAppBranchRunStatus(ctx, &activities.UpdateAppBranchRunStatusRequest{
		RunID:  run.ID,
		Status: "running",
	})
	if err != nil {
		logger.Error("unable to update run status to running", "error", err)
		return err
		// Continue execution even if status update fails
	}

	// NOTE(jm): this is all mostly compatibility stuff from previous iterations of this tooling. Will remove once
	// app branches is landed.
	eventLoopReq := eventloop.EventLoopRequest{
		ID: branch.ID, // App branch ID is the event loop ID
	}

	// Create the WorkflowConductor with queue-based signal execution
	fc := &flow.WorkflowConductor[*appsignals.Signal]{
		Cfg: nil, // Not needed for app signals
		V:   nil, // Not needed for app signals
		MW:  nil, // Not needed for app signals

		Generators: getWorkflowStepGenerators(),
		ExecFn:     getExecuteFlowExecFn(eventLoopReq),
	}

	// Execute the flow directly - this will generate and execute workflow steps
	err = fc.Handle(ctx, eventLoopReq, *run.WorkflowID, 0)
	if err != nil {
		logger.Error("workflow execution failed", "error", err)

		// Mark run as failed
		_, updateErr := activities.AwaitUpdateAppBranchRunStatus(ctx, &activities.UpdateAppBranchRunStatusRequest{
			RunID:  run.ID,
			Status: "failed",
		})
		if updateErr != nil {
			logger.Error("unable to update run status to failed", "error", updateErr)
		}

		return fmt.Errorf("workflow execution failed: %w", err)
	}

	// Mark as success
	_, err = activities.AwaitUpdateAppBranchRunStatus(ctx, &activities.UpdateAppBranchRunStatusRequest{
		RunID:  run.ID,
		Status: "success",
	})
	if err != nil {
		logger.Error("unable to update run status to success", "error", err)
	}

	logger.Info("app branch run completed successfully",
		"run_id", run.ID,
		"app_branch_id", branch.ID,
		"workflow_id", *run.WorkflowID,
	)

	return nil
}

// getWorkflowStepGenerators returns the workflow step generator map
func getWorkflowStepGenerators() map[app.WorkflowType]flow.WorkflowStepGenerator {
	return map[app.WorkflowType]flow.WorkflowStepGenerator{
		app.WorkflowTypeAppBranchesRun: workflows.AppBranchRun,
	}
}

// getExecuteFlowExecFn returns the execution function for workflow steps
// This routes each step's signal through the queue system
func getExecuteFlowExecFn(eventLoopReq eventloop.EventLoopRequest) func(workflow.Context, eventloop.EventLoopRequest, *appsignals.Signal, app.WorkflowStep) error {
	return func(ctx workflow.Context, ereq eventloop.EventLoopRequest, sig *appsignals.Signal, step app.WorkflowStep) error {
		logger := workflow.GetLogger(ctx)

		logger.Info("enqueuing signal to queue",
			"step_name", step.Name,
			"owner_id", eventLoopReq.ID,
			"owner_type", "app_branches",
		)

		// TODO: fetch the queue ID for this app branch owner
		queueID := "TODO"

		// Enqueue the signal via the queue client
		enqueueResp, err := client.AwaitEnqueueSignal(ctx, &client.EnqueueSignalRequest{
			QueueID: queueID,
			Signal: &checkchanges.Signal{
				AppBranchID: "DNE",
			},
		})
		if err != nil {
			return errors.Wrapf(err, "unable to enqueue signal for step %s", step.Name)
		}

		logger.Info("waiting for queue signal to complete",
			"step_name", step.Name,
			"queue_signal_id", enqueueResp.ID,
			"workflow_id", enqueueResp.WorkflowID,
		)

		// Wait for the queue signal to complete execution
		_, err = client.AwaitAwaitSignal(ctx, enqueueResp.ID)
		if err != nil {
			return errors.Wrapf(err, "queue signal execution failed for step %s", step.Name)
		}

		logger.Info("queue signal completed successfully", "step_name", step.Name)

		return nil
	}
}
