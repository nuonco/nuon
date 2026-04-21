package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// RetryStepRequest is the input for the "retry-step" update handler.
type RetryStepRequest struct {
	StepID string `json:"step_id"`
}

// RetryStepResponse is the response from the "retry-step" update handler.
type RetryStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	Retryable  bool   `json:"retryable"`
}

// retryStepHandler owns the full retry lifecycle. It:
//  1. Tells the step signal it was retried (marks discarded, gets directive)
//  2. Clones the step
//  3. Re-dispatches the group signal so the clone gets executed
//
// This works regardless of whether the group signal is alive or dead.
func (s *Signal) retryStepHandler(ctx workflow.Context, req RetryStepRequest) (*RetryStepResponse, error) {
	l, _ := log.WorkflowLogger(ctx)

	// 1. Tell the step it was retried.
	retryResp, err := workflowactivities.AwaitForwardCreateStepRetry(ctx, workflowactivities.ForwardCreateStepRetryRequest{
		StepID: req.StepID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to forward retry to step %s: %w", req.StepID, err)
	}

	// 2. If retry-group, set the resume state so the flow re-dispatches the group.
	if retryResp.Directive == "retry-group" {
		s.resumeRequested = true
		s.resumeRunType = app.WorkflowRunTypeRetry
		s.resumeStepID = req.StepID
		s.resumeStartIdx = s.findGroupPositionForStep(ctx, req.StepID)
		return &RetryStepResponse{WorkflowID: s.WorkflowID, Retryable: true}, nil
	}

	// 3. Clone the step.
	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	flw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflow: %w", err)
	}

	_, err = workflowactivities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, workflowactivities.CreateFlowStepsRequest{
		Steps: []workflowactivities.CreateFlowStep{
			{
				FlowID:      flw.ID,
				OwnerID:     flw.OwnerID,
				OwnerType:   flw.OwnerType,
				Name:        step.Name,
				Signal:      step.Signal,
				QueueSignal: step.QueueSignal,
				Status: app.NewCompositeTemporalStatus(ctx, app.StatusPending, map[string]any{
					"is_retry":  true,
					"retry_idx": step.RetryIndex + 1,
				}),
				Idx:                 step.Idx + 1,
				ExecutionType:       step.ExecutionType,
				Metadata:            step.Metadata,
				Retryable:           step.Retryable,
				Skippable:           step.Skippable,
				GroupIdx:            step.GroupIdx,
				GroupRetryIdx:       step.GroupRetryIdx,
				WorkflowStepGroupID: step.WorkflowStepGroupID,
				StepTargetType:      step.StepTargetType,
				RetryIndex:          step.RetryIndex + 1,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to clone step: %w", err)
	}

	// 4. Re-dispatch the group signal so the clone gets executed.
	group, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepGroupByID(ctx, step.WorkflowStepGroupID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step group: %w", err)
	}

	l.Debug("re-dispatching group signal for retry",
		zap.String("step_id", req.StepID),
		zap.String("step_group_id", group.ID))

	if _, err := s.executeGroup(ctx, group, flw); err != nil {
		return nil, fmt.Errorf("unable to re-dispatch group: %w", err)
	}

	return &RetryStepResponse{
		WorkflowID: s.WorkflowID,
		Retryable:  true,
	}, nil
}
