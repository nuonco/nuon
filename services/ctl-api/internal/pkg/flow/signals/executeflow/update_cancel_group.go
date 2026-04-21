package executeflow

import "go.temporal.io/sdk/workflow"

// CancelGroupRequest is the input for the "cancel-group" update handler.
type CancelGroupRequest struct {
	// StepID is any step in the group to cancel. The handler resolves the
	// group from this step's GroupIdx and cancels all in-flight steps.
	StepID string `json:"step_id"`
}

// CancelGroupResponse is the response from the "cancel-group" update handler.
type CancelGroupResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// cancelGroupHandler cancels the workflow. The main execute loop checks
// cancelRequested after each group and exits. The group signal's Cancel()
// method handles propagating cancellation to in-flight steps.
func (s *Signal) cancelGroupHandler(ctx workflow.Context, req CancelGroupRequest) (*CancelGroupResponse, error) {
	s.cancelRequested = true
	return &CancelGroupResponse{WorkflowID: s.WorkflowID}, nil
}
