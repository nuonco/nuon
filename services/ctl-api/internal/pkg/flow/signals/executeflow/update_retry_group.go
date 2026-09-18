package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// RetryGroupRequest is the input for the "retry-group" update handler.
type RetryGroupRequest struct {
	// StepID is any step in the group to retry. The handler resolves the
	// group from this step's GroupIdx.
	StepID string `json:"step_id"`
}

// RetryGroupResponse is the response from the "retry-group" update handler.
type RetryGroupResponse struct {
	WorkflowID string `json:"workflow_id"`
	Retryable  bool   `json:"retryable"`
}

func (s *Signal) retryGroupHandler(ctx workflow.Context, req RetryGroupRequest) (*RetryGroupResponse, error) {
	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	if err := s.cloneGroupForRetry(ctx, step.GroupIdx); err != nil {
		return nil, fmt.Errorf("unable to clone group for retry: %w", err)
	}

	// resumeRequested must be set last. The paused Execute loop acts on this
	// flag the instant it flips, and reads the fields written just above it.
	// The DB lookup above pauses this handler long enough for Execute to run,
	// so setting the flag first means starting the retry from a stale
	// resumeStartIdx.
	s.resumeRunType = app.WorkflowRunTypeRetry
	s.resumeStepID = req.StepID
	s.resumeStartIdx = s.findGroupPositionForStep(ctx, req.StepID)

	s.resumeRequested = true

	return &RetryGroupResponse{WorkflowID: s.WorkflowID, Retryable: true}, nil
}
