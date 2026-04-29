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

// EagerStepGroupsResult holds the result of an eager step generation — the eager
// groups are persisted and ready for execution, while remaining groups can be
// fetched later via CompleteStepGeneration.
type EagerStepGroupsResult struct {
	Workflow      *app.Workflow
	QueueSignalID string
}

// generateStepsViaSignal enqueues the workflow's GenerateStepsSignal, fetches the
// generated steps and groups, then persists them directly — no child workflow needed.
func generateStepsViaSignal(ctx workflow.Context, cfg StepConfig, flw *app.Workflow) (*app.Workflow, error) {
	// 1. Idempotency: if steps already exist (e.g. after continue-as-new), return them.
	existingSteps, err := activities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, flw.ID)
	if err == nil && len(existingSteps) > 0 {
		flw.Steps = existingSteps
		return flw, nil
	}

	queueSignalID, err := enqueueGenerateStepsSignal(ctx, cfg, flw)
	if err != nil {
		return nil, err
	}

	// Fetch ALL steps at once.
	result, err := queueclient.AwaitFetchSteps(ctx, queueclient.FetchStepsRequest{
		QueueSignalID: queueSignalID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch steps from generate-steps signal")
	}

	return persistGenerateResult(ctx, flw, result)
}

// generateEagerStepGroups enqueues the generate-steps signal, fetches the eager
// step groups via the "eager-step-groups" update, and persists them.
// The caller can start executing these groups immediately, then call
// completeStepGeneration() to fetch and persist the remaining groups.
func generateEagerStepGroups(ctx workflow.Context, cfg StepConfig, flw *app.Workflow) (*EagerStepGroupsResult, error) {
	// 1. Idempotency: if steps already exist (e.g. after continue-as-new), return them.
	existingSteps, err := activities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, flw.ID)
	if err == nil && len(existingSteps) > 0 {
		flw.Steps = existingSteps
		return &EagerStepGroupsResult{Workflow: flw}, nil
	}

	queueSignalID, err := enqueueGenerateStepsSignal(ctx, cfg, flw)
	if err != nil {
		return nil, err
	}

	// Fetch the eager step groups.
	result, err := queueclient.AwaitFetchEagerStepGroups(ctx, queueclient.FetchEagerStepGroupsRequest{
		QueueSignalID: queueSignalID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch eager step groups from generate-steps signal")
	}

	flw, err = persistGenerateResult(ctx, flw, result)
	if err != nil {
		return nil, err
	}

	return &EagerStepGroupsResult{
		Workflow:      flw,
		QueueSignalID: queueSignalID,
	}, nil
}

// completeStepGeneration fetches ALL steps via "FetchSteps" and persists any
// groups/steps not already in the DB. This is called after group 0 finishes
// (or concurrently) to ensure all remaining groups are available.
func completeStepGeneration(ctx workflow.Context, cfg StepConfig, flw *app.Workflow, queueSignalID string) (*app.Workflow, error) {
	if queueSignalID == "" {
		// No early start was used — steps are already complete.
		return flw, nil
	}

	result, err := queueclient.AwaitFetchSteps(ctx, queueclient.FetchStepsRequest{
		QueueSignalID: queueSignalID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch remaining steps from generate-steps signal")
	}

	// Filter out groups/steps already persisted (group 0).
	existingSteps, _ := activities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, flw.ID)
	existingGroupIDs := make(map[int]bool)
	for _, s := range existingSteps {
		existingGroupIDs[s.GroupIdx] = true
	}

	var remainingGroups []*app.WorkflowStepGroup
	for _, g := range result.Groups {
		if !existingGroupIDs[g.GroupIdx] {
			remainingGroups = append(remainingGroups, g)
		}
	}

	var remainingSteps []*app.WorkflowStep
	for _, s := range result.Steps {
		if !existingGroupIDs[s.GroupIdx] {
			remainingSteps = append(remainingSteps, s)
		}
	}

	if len(remainingGroups) == 0 && len(remainingSteps) == 0 {
		return flw, nil
	}

	remaining := &app.GenerateStepsResult{
		Steps:  remainingSteps,
		Groups: remainingGroups,
	}

	return persistGenerateResult(ctx, flw, remaining)
}

// enqueueGenerateStepsSignal sets up and enqueues the generate-steps signal,
// returning the queue signal ID for subsequent update calls.
func enqueueGenerateStepsSignal(ctx workflow.Context, cfg StepConfig, flw *app.Workflow) (string, error) {
	type workflowIDSetter interface {
		SetWorkflowID(id string)
	}
	if setter, ok := flw.GenerateStepsSignal.Signal.(workflowIDSetter); ok {
		setter.SetWorkflowID(flw.ID)
	}

	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		QueueName:       cfg.TargetQueueName,
		Signal:          flw.GenerateStepsSignal.Signal,
		SignalOwnerID:   flw.ID,
		SignalOwnerType: "install_workflows",
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to enqueue generate-steps signal")
	}

	return enqueueResp.QueueSignalID, nil
}

// persistGenerateResult assigns IDs to groups/steps, injects step context, and
// persists them to the DB. It appends to the workflow's existing Steps/StepGroups.
func persistGenerateResult(ctx workflow.Context, flw *app.Workflow, result *app.GenerateStepsResult) (*app.Workflow, error) {
	steps := result.Steps
	groups := result.Groups

	// Pre-generate group IDs and build GroupIdx→GroupID map.
	groupIDByIdx := make(map[int]string)
	if len(groups) > 0 {
		for _, g := range groups {
			g.ID = domains.NewWorkflowStepGroupID()
			g.WorkflowID = flw.ID
			groupIDByIdx[g.GroupIdx] = g.ID
		}

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
			})
		}
		if _, err := activities.AwaitPkgWorkflowsFlowCreateFlowStepGroups(ctx, groupsReq); err != nil {
			return nil, errors.Wrap(err, "unable to persist workflow step groups")
		}
	}

	// Pre-generate step IDs and inject step context into signals.
	for _, step := range steps {
		step.ID = domains.NewWorkflowStepID()
		if groupID, ok := groupIDByIdx[step.GroupIdx]; ok {
			step.WorkflowStepGroupID = groupID
		}
		if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
			signal.ApplyStepContext(step.QueueSignal.Signal, step.ID, flw.ID)
		}
	}

	// Determine starting Idx offset based on existing steps.
	startIdx := len(flw.Steps)

	stepsReq := activities.CreateFlowStepsRequest{
		Steps: make([]activities.CreateFlowStep, 0, len(steps)),
	}
	for idx, step := range steps {
		step.Idx = (startIdx + idx) * 100
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
			StepQueueID:         step.StepQueueID,
			TargetQueueID:       step.TargetQueueID,
		})
	}

	resp, err := activities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, stepsReq)
	if err != nil {
		return nil, errors.Wrap(err, "unable to persist workflow steps")
	}

	for _, step := range resp {
		flw.Steps = append(flw.Steps, *step)
	}

	for _, g := range groups {
		flw.StepGroups = append(flw.StepGroups, *g)
	}

	return flw, nil
}
