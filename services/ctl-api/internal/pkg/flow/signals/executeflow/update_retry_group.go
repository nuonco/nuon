package executeflow

import (
	"go.temporal.io/sdk/workflow"
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
	panic("not yet implemented")
	return nil, nil
}
