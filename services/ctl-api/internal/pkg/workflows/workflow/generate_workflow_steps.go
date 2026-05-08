package workflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type GenerateWorkflowStepsRequest struct {
	WorkflowID string              `json:"workflow_id" validate:"required"`
	Steps      []*app.WorkflowStep `json:"steps" validate:"required"`
}

// @temporal-gen-v2 workflow
// @execution-timeout 1h
// @task-timeout 1m
// @id-template {{.CallerID}}-generate-steps
func (w *Workflows) GenerateWorkflowSteps(ctx workflow.Context, req *GenerateWorkflowStepsRequest) ([]*app.WorkflowStep, error) {
	fid := req.WorkflowID

	// Check if steps already exist - return them for idempotency.
	// This is critical for continue-as-new semantics where this child workflow
	// may be called multiple times across workflow runs.
	existingSteps, err := activities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, fid)
	if err == nil && len(existingSteps) > 0 {
		result := make([]*app.WorkflowStep, len(existingSteps))
		for i := range existingSteps {
			step := existingSteps[i]
			result[i] = &step
		}
		return result, nil
	}

	wflw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, fid)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflow by ID %s: %w", fid, err)
	}

	steps := req.Steps

	// Pre-generate step IDs and inject step context (stepID, flowID) into signals
	// before persisting, so signals have access to their own step ID and the parent flow ID.
	for _, step := range steps {
		step.ID = domains.NewWorkflowStepID()
		if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
			signal.ApplyStepContext(step.QueueSignal.Signal, step.ID, fid)
		}
	}

	stepsReq := activities.CreateFlowStepsRequest{
		Steps: make([]activities.CreateFlowStep, 0, len(steps)),
	}

	for idx, step := range steps {
		step.Idx = idx * 100
		stepsReq.Steps = append(stepsReq.Steps, activities.CreateFlowStep{
			ID:            step.ID,
			FlowID:        fid,
			OwnerID:       wflw.OwnerID,
			OwnerType:     wflw.OwnerType,
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
			Timeout:       step.Timeout,
		})
	}

	resp, err := activities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, stepsReq)
	if err != nil {
		return nil, fmt.Errorf("unable to create steps: %w", err)
	}

	for i, wflwStep := range resp {
		steps[i].ID = wflwStep.ID
	}

	return steps, nil
}
