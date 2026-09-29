package flow

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type EagerStepGroupsResult struct {
	Workflow      *app.Workflow
	QueueSignalID string
}

func generateStepsViaSignal(ctx workflow.Context, cfg StepConfig, flw *app.Workflow) (*app.Workflow, error) {
	// why: 1. Idempotency: if steps already exist (e.g. after continue-as-new), return them.
	existingSteps, err := activities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, flw.ID)
	if err == nil && len(existingSteps) > 0 {
		flw.Steps = existingSteps
		return flw, nil
	}

	queueSignalID, err := enqueueGenerateStepsSignal(ctx, cfg, flw)
	if err != nil {
		return nil, err
	}

	result, err := queueclient.AwaitFetchSteps(ctx, queueclient.FetchStepsRequest{
		QueueSignalID: queueSignalID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch steps from generate-steps signal")
	}

	return persistGenerateResult(ctx, flw, result)
}

func generateEagerStepGroups(ctx workflow.Context, cfg StepConfig, flw *app.Workflow) (*EagerStepGroupsResult, error) {
	// why: 1. Idempotency: if steps already exist (e.g. after continue-as-new), return them.
	existingSteps, err := activities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, flw.ID)
	if err == nil && len(existingSteps) > 0 {
		flw.Steps = existingSteps
		return &EagerStepGroupsResult{Workflow: flw}, nil
	}

	queueSignalID, err := enqueueGenerateStepsSignal(ctx, cfg, flw)
	if err != nil {
		return nil, err
	}

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

func completeStepGeneration(ctx workflow.Context, cfg StepConfig, flw *app.Workflow, queueSignalID string) (*app.Workflow, error) {
	if queueSignalID == "" {
		return flw, nil
	}

	result, err := queueclient.AwaitFetchSteps(ctx, queueclient.FetchStepsRequest{
		QueueSignalID: queueSignalID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch remaining steps from generate-steps signal")
	}

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

func enqueueGenerateStepsSignal(ctx workflow.Context, cfg StepConfig, flw *app.Workflow) (string, error) {
	type workflowIDSetter interface {
		SetWorkflowID(id string)
	}
	if setter, ok := flw.GenerateStepsSignal.Signal.(workflowIDSetter); ok {
		setter.SetWorkflowID(flw.ID)
	}

	type lifecycleIdentitySetter interface {
		SetLifecycleIdentity(orgID, orgName, ownerID, ownerName string)
	}
	if setter, ok := flw.GenerateStepsSignal.Signal.(lifecycleIdentitySetter); ok {
		setter.SetLifecycleIdentity(flw.OrgID, flw.Org.Name, cfg.OwnerID, flw.OwnerName)
	}

	queueName := cfg.GenerateStepsQueueName
	if queueName == "" {
		queueName = cfg.TargetQueueName
	}

	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		QueueName:       queueName,
		Signal:          flw.GenerateStepsSignal.Signal,
		SignalOwnerID:   flw.ID,
		SignalOwnerType: "install_workflows",
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to enqueue generate-steps signal")
	}

	return enqueueResp.QueueSignalID, nil
}

func persistGenerateResult(ctx workflow.Context, flw *app.Workflow, result *app.GenerateStepsResult) (*app.Workflow, error) {
	steps := result.Steps
	groups := result.Groups

	groupIDByIdx := make(map[int]string)
	if len(groups) > 0 {
		for _, g := range groups {
			g.ID = domains.NewWorkflowStepGroupID()
			g.WorkflowID = flw.ID
			groupIDByIdx[g.GroupIdx] = g.ID
		}

		groupParallel := make(map[int]bool, len(groups))
		for _, g := range groups {
			groupParallel[g.GroupIdx] = g.Parallel
		}
		groupTimeouts := make(map[int]time.Duration, len(groups))
		for _, step := range steps {
			t := step.Timeout
			if t < 0 {
				groupTimeouts[step.GroupIdx] = signal.UnboundedTimeout
				continue
			}
			if t == 0 || groupTimeouts[step.GroupIdx] == signal.UnboundedTimeout {
				continue
			}
			if groupParallel[step.GroupIdx] {
				if t > groupTimeouts[step.GroupIdx] {
					groupTimeouts[step.GroupIdx] = t
				}
			} else {
				groupTimeouts[step.GroupIdx] += t
			}
		}

		groupsReq := activities.CreateFlowStepGroupsRequest{
			Groups: make([]activities.CreateFlowStepGroup, 0, len(groups)),
		}
		for _, g := range groups {
			g.Timeout = groupTimeouts[g.GroupIdx]
			groupsReq.Groups = append(groupsReq.Groups, activities.CreateFlowStepGroup{
				ID:         g.ID,
				WorkflowID: flw.ID,
				GroupIdx:   g.GroupIdx,
				Parallel:   g.Parallel,
				Name:       g.Name,
				Status:     g.Status,
				Timeout:    g.Timeout,
			})
		}
		if _, err := activities.AwaitPkgWorkflowsFlowCreateFlowStepGroups(ctx, groupsReq); err != nil {
			return nil, errors.Wrap(err, "unable to persist workflow step groups")
		}
	}

	for _, step := range steps {
		step.ID = domains.NewWorkflowStepID()
		if groupID, ok := groupIDByIdx[step.GroupIdx]; ok {
			step.WorkflowStepGroupID = groupID
		}
		if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
			signal.ApplyStepContext(step.QueueSignal.Signal, step.ID, flw.ID)
		}
	}

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
			SkipOnFailure:       step.SkipOnFailure,
			GroupIdx:            step.GroupIdx,
			WorkflowStepGroupID: step.WorkflowStepGroupID,
			StepQueueID:         step.StepQueueID,
			TargetQueueID:       step.TargetQueueID,
			Timeout:             step.Timeout,
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
