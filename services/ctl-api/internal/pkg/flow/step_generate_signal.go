package flow

// step_generate_signal.go — signal-based step generation.
//
// When a Workflow has GenerateStepsSignal set, this path is used instead of the
// legacy generator-map + child-workflow path. The flow is:
//
//  1. Idempotency check — if steps already exist for this workflow, return them.
//  2. Enqueue the generate-steps signal to the target queue (e.g. install-signals).
//  3. Send the "FetchSteps" update to the signal's handler workflow.
//  4. Receive the generated []*app.WorkflowStep.
//  5. Assign IDs, inject step context (stepID + flowID) into each signal.
//  6. Persist steps to DB via the CreateFlowSteps activity.

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// generateStepsViaSignal enqueues the workflow's GenerateStepsSignal, fetches the
// generated steps, then persists them directly — no child workflow needed.
func generateStepsViaSignal(ctx workflow.Context, cfg StepConfig, flw *app.Workflow) (*app.Workflow, error) {
	// 1. Idempotency: if steps already exist (e.g. after continue-as-new), return them.
	existingSteps, err := activities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, flw.ID)
	if err == nil && len(existingSteps) > 0 {
		flw.Steps = existingSteps
		return flw, nil
	}

	// 2. Set the workflow ID on the signal so its Execute() can look up the workflow.
	type workflowIDSetter interface {
		SetWorkflowID(id string)
	}
	if setter, ok := flw.GenerateStepsSignal.Signal.(workflowIDSetter); ok {
		setter.SetWorkflowID(flw.ID)
	}

	// 3. Enqueue the generate-steps signal to the target queue.
	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		QueueName:       cfg.TargetQueueName,
		Signal:          flw.GenerateStepsSignal.Signal,
		SignalOwnerID:   flw.ID,
		SignalOwnerType: "install_workflows",
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to enqueue generate-steps signal")
	}

	// 4. Send "FetchSteps" update and receive the generated steps.
	steps, err := queueclient.AwaitFetchSteps(ctx, queueclient.FetchStepsRequest{
		QueueSignalID: enqueueResp.QueueSignalID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch steps from generate-steps signal")
	}

	// 5. Pre-generate step IDs and inject step context into signals.
	for _, step := range steps {
		step.ID = domains.NewWorkflowStepID()
		if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
			signal.ApplyStepContext(step.QueueSignal.Signal, step.ID, flw.ID)
		}
	}

	// 6. Persist steps to DB.
	stepsReq := activities.CreateFlowStepsRequest{
		Steps: make([]activities.CreateFlowStep, 0, len(steps)),
	}
	for idx, step := range steps {
		step.Idx = idx * 10
		stepsReq.Steps = append(stepsReq.Steps, activities.CreateFlowStep{
			ID:            step.ID,
			FlowID:        flw.ID,
			OwnerID:       flw.OwnerID,
			OwnerType:     flw.OwnerType,
			Status:        step.Status,
			Name:          step.Name,
			Signal:        step.Signal,
			QueueSignal:   step.QueueSignal,
			Idx:           step.Idx,
			ExecutionType: step.ExecutionType,
			Metadata:      step.Metadata,
			Retryable:     step.Retryable,
			Skippable:     step.Skippable,
			GroupIdx:      step.GroupIdx,
		})
	}

	resp, err := activities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, stepsReq)
	if err != nil {
		return nil, errors.Wrap(err, "unable to persist workflow steps")
	}

	flw.Steps = make([]app.WorkflowStep, len(resp))
	for i, step := range resp {
		flw.Steps[i] = *step
	}

	return flw, nil
}
