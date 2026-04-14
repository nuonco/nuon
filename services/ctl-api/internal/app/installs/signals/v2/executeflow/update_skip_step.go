package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// SkipStepRequest is the input for the "skip-step" update handler.
type SkipStepRequest struct {
	StepID string `json:"step_id"`
}

// SkipStepResponse is the response from the "skip-step" update handler.
type SkipStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	Skippable  bool   `json:"skippable"`
}

func (s *Signal) skipStepHandler(ctx workflow.Context, req SkipStepRequest) (*SkipStepResponse, error) {
	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	// Only error state steps can be skipped
	if step.Status.Status != app.StatusError {
		return &SkipStepResponse{
			WorkflowID: s.InstallWorkflowID,
			Skippable:  false,
		}, nil
	}

	if !step.Skippable {
		return &SkipStepResponse{
			WorkflowID: s.InstallWorkflowID,
			Skippable:  false,
		}, nil
	}

	s.skipRequested = true
	s.skipStepID = req.StepID

	// If this is a plan step (approval execution type), find the apply step in the same group
	// so we can skip both plan and apply together.
	if step.ExecutionType == app.WorkflowStepExecutionTypeApproval {
		allSteps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, s.InstallWorkflowID)
		if err == nil {
			for _, s2 := range allSteps {
				if s2.GroupIdx == step.GroupIdx &&
					s2.GroupRetryIdx == step.GroupRetryIdx &&
					s2.ID != step.ID &&
					s2.ExecutionType == app.WorkflowStepExecutionTypeSystem {
					s.skipAdditionalStepIDs = append(s.skipAdditionalStepIDs, s2.ID)
					break
				}
			}
		}
	}

	return &SkipStepResponse{
		WorkflowID: s.InstallWorkflowID,
		Skippable:  true,
	}, nil
}
