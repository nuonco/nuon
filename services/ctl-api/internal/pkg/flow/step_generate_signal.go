package flow

// step_generate_signal.go — signal-based step generation.
//
// When a Workflow has GenerateStepsSignal set, this path is used instead of the
// legacy generator-map + child-workflow path. The flow is:
//
//  1. Idempotency check — if steps already exist for this workflow, return them.
//  2. Enqueue the generate-steps signal to the target queue (e.g. install-signals).
//  3. Send the "FetchSteps" update to the signal's handler workflow.
//  4. Receive the generated GenerateStepsResult (steps + groups).
//  5. Assign IDs, inject step context (stepID + flowID) into each signal.
//  6. Persist groups and steps to DB.

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
// generated steps and groups, then persists them directly — no child workflow needed.
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

	// 4. Send "FetchSteps" update and receive the generated result.
	result, err := queueclient.AwaitFetchSteps(ctx, queueclient.FetchStepsRequest{
		QueueSignalID: enqueueResp.QueueSignalID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch steps from generate-steps signal")
	}

	steps := result.Steps
	groups := result.Groups

	// 5. Pre-generate group IDs and build GroupIdx→GroupID map.
	groupIDByIdx := make(map[int]string)
	if len(groups) > 0 {
		for _, g := range groups {
			g.ID = domains.NewWorkflowStepGroupID()
			g.WorkflowID = flw.ID
			groupIDByIdx[g.GroupIdx] = g.ID
		}

		// Persist groups to DB.
		groupsReq := activities.CreateFlowStepGroupsRequest{
			Groups: make([]activities.CreateFlowStepGroup, 0, len(groups)),
		}
		for _, g := range groups {
			groupsReq.Groups = append(groupsReq.Groups, activities.CreateFlowStepGroup{
				ID:         g.ID,
				WorkflowID: flw.ID,
				GroupIdx:   g.GroupIdx,
				Parallel:   g.Parallel,
				Name:       g.Name,
				Status:     g.Status,
				Labels:     g.Labels,
			})
		}
		if _, err := activities.AwaitPkgWorkflowsFlowCreateFlowStepGroups(ctx, groupsReq); err != nil {
			return nil, errors.Wrap(err, "unable to persist workflow step groups")
		}
	}

	// 6. Pre-generate step IDs and inject step context into signals.
	for _, step := range steps {
		step.ID = domains.NewWorkflowStepID()
		if groupID, ok := groupIDByIdx[step.GroupIdx]; ok {
			step.WorkflowStepGroupID = groupID
		}
		if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
			signal.ApplyStepContext(step.QueueSignal.Signal, step.ID, flw.ID)
		}
	}

	// 7. Persist steps to DB.
	stepsReq := activities.CreateFlowStepsRequest{
		Steps: make([]activities.CreateFlowStep, 0, len(steps)),
	}
	for idx, step := range steps {
		step.Idx = idx * 100
		stepsReq.Steps = append(stepsReq.Steps, activities.CreateFlowStep{
			ID:                  step.ID,
			FlowID:              flw.ID,
			OwnerID:             flw.OwnerID,
			OwnerType:           flw.OwnerType,
			Status:              step.Status,
			Name:                step.Name,
			Signal:              step.Signal,
			QueueSignal:         step.QueueSignal,
			Idx:                 step.Idx,
			ExecutionType:       step.ExecutionType,
			Metadata:            step.Metadata,
			Retryable:           step.Retryable,
			Skippable:           step.Skippable,
			GroupIdx:            step.GroupIdx,
			WorkflowStepGroupID: step.WorkflowStepGroupID,
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

	flw.StepGroups = make([]app.WorkflowStepGroup, len(groups))
	for i, g := range groups {
		flw.StepGroups[i] = *g
	}

	return flw, nil
}
