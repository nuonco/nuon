package workflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type GenerateWorkflowStepsRequest struct {
	WorkflowID string              `json:"workflow_id" validate:"required"`
	Steps      []*app.WorkflowStep `json:"steps" validate:"required"`
}

// @temporal-gen workflow
// @execution-timeout 1h
// @task-timeout 1m
// @id-callback WorkflowIDCallback
func (w *Workflows) GenerateWorkflowSteps(ctx workflow.Context, req *GenerateWorkflowStepsRequest) ([]*app.WorkflowStep, error) {
	fid := req.WorkflowID

	wflw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, fid)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflow by ID %s: %w", fid, err)
	}

	steps := req.Steps

	for idx, step := range steps {
		step.Idx = idx
		s, err := activities.AwaitPkgWorkflowsFlowCreateFlowStep(ctx, activities.CreateFlowStepRequest{
			FlowID:        fid,
			OwnerID:       wflw.OwnerID,
			OwnerType:     wflw.OwnerType,
			Status:        step.Status,
			Name:          step.Name,
			Signal:        step.Signal,
			Idx:           step.Idx,
			ExecutionType: step.ExecutionType,
			Metadata:      step.Metadata,
			Retryable:     step.Retryable,
			Skippable:     step.Skippable,
			GroupIdx:      step.GroupIdx,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create steps: %w", err)
		}
		step.ID = s.ID
	}

	return steps, nil
}
