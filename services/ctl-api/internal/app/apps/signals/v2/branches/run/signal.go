package run

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appsignals "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "app-branch-run"

type Signal struct {
	RunID string `json:"run_id" validate:"required"` // The app branch run ID - everything else fetched from DB
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	// Use playground validator for struct tag validation
	v := validator.New()
	if err := v.Struct(s); err != nil {
		return errors.Wrap(err, "validation failed")
	}

	// Validate run exists and fetch all related data
	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return errors.Wrap(err, "app branch run not found")
	}

	// Validate run has required relationships
	if run.AppBranchID == "" {
		return errors.New("run missing app_branch_id")
	}
	if run.AppBranchConfigID == "" {
		return errors.New("run missing app_branch_config_id")
	}

	return nil
}

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

	config, err := activities.AwaitGetAppConfigByIDByAppConfigID(ctx, run.AppBranchConfigID)
	if err != nil {
		return fmt.Errorf("unable to get app branch config: %w", err)
	}

	logger.Info("starting app branch run",
		"run_id", run.ID,
		"app_branch_id", branch.ID,
		"app_branch_name", branch.Name,
		"config_id", config.ID,
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
		// Continue execution even if status update fails
	}

	// Embed WorkflowConductor directly to execute the flow
	// No need for indirection through a child workflow - just call it directly

	// Build the event loop request
	eventLoopReq := eventloop.EventLoopRequest{
		ID: branch.ID, // App branch ID is the event loop ID
	}

	// Create the WorkflowConductor with queue-based signal execution
	fc := &flow.WorkflowConductor[*appsignals.Signal]{
		Cfg:        nil, // Not needed for app signals
		V:          nil, // Not needed for app signals
		MW:         nil, // Not needed for app signals
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

		// Unmarshal the signal JSON from the workflow step to get the actual signal to execute
		// The step contains the serialized signal that needs to be executed
		var queueSig queuesignal.Signal
		if err := json.Unmarshal(step.Signal.SignalJSON, &queueSig); err != nil {
			return errors.Wrapf(err, "unable to unmarshal signal JSON for step %s", step.Name)
		}

		// Look up the signal constructor in the catalog to create a properly typed signal
		sigConstructor, ok := catalog.SignalCatalog[queuesignal.SignalType(step.Signal.Type)]
		if !ok {
			return errors.Errorf("signal type %s not found in catalog", step.Signal.Type)
		}

		// Create a new signal instance and unmarshal the JSON into it
		typedSignal := sigConstructor()
		if err := json.Unmarshal(step.Signal.SignalJSON, typedSignal); err != nil {
			return errors.Wrapf(err, "unable to unmarshal typed signal for step %s", step.Name)
		}

		logger.Info("enqueuing signal to queue",
			"step_name", step.Name,
			"signal_type", step.Signal.Type,
			"owner_id", eventLoopReq.ID,
			"owner_type", "app_branches",
		)

		// Enqueue the signal to the queue owned by the app branch
		// This routes the signal through the queue system for execution
		enqueueResp, err := workflowactivities.AwaitEnqueueSignalToOwner(ctx, &workflowactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   eventLoopReq.ID, // App branch ID
			OwnerType: "app_branches",
			Signal:    typedSignal,
		})
		if err != nil {
			return errors.Wrapf(err, "unable to enqueue signal for step %s", step.Name)
		}

		logger.Info("waiting for queue signal to complete",
			"step_name", step.Name,
			"queue_signal_id", enqueueResp.QueueSignalID,
			"workflow_id", enqueueResp.WorkflowID,
		)

		// Wait for the queue signal to complete execution
		_, err = client.AwaitAwaitSignal(ctx, enqueueResp.QueueSignalID)
		if err != nil {
			return errors.Wrapf(err, "queue signal execution failed for step %s", step.Name)
		}

		logger.Info("queue signal completed successfully", "step_name", step.Name)

		return nil
	}
}
