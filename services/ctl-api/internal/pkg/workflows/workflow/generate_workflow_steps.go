package workflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type GenerateWorkflowStepsRequest struct {
	WorkflowID string              `json:"workflow_id" validate:"required"`
	Steps      []*app.WorkflowStep `json:"steps" validate:"required"`
}

// @temporal-gen workflow
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
			result[i] = &existingSteps[i]
		}
		return result, nil
	}

	wflw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, fid)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflow by ID %s: %w", fid, err)
	}

	steps := req.Steps

	stepsReq := activities.CreateFlowStepsRequest{
		Steps: make([]activities.CreateFlowStep, 0, len(steps)),
	}

	for idx, step := range steps {
		step.Idx = idx
		stepsReq.Steps = append(stepsReq.Steps, activities.CreateFlowStep{
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
