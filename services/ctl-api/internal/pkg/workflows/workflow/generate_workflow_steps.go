package workflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type GenerateWorkflowStepsRequest struct {
	SignalID   string              `json:"signal_id" validate:"required"`
	WorkflowID string              `json:"workflow_id" validate:"required"`
	Steps      []*app.WorkflowStep `json:"steps" validate:"required"`
}

// @temporal-gen workflow
// @execution-timeout 1h
// @task-timeout 1m
// @id-template {{.Req.SignalID}}-execute-workflow-{{.Req.WorkflowID}}-generate-steps
func (w *Workflows) GenerateWorkflowSteps(ctx workflow.Context, req *GenerateWorkflowStepsRequest) ([]*app.WorkflowStep, error) {
	fid := req.WorkflowID

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
