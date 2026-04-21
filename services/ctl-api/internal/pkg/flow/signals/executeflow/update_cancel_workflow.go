package executeflow

import "go.temporal.io/sdk/workflow"

// CancelWorkflowResponse is the response from the "cancel-workflow" update handler.
type CancelWorkflowResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// cancelWorkflowHandler cancels the entire workflow. The main execute loop
// checks cancelRequested after each group and exits.
func (s *Signal) cancelWorkflowHandler(ctx workflow.Context) (*CancelWorkflowResponse, error) {
	s.cancelRequested = true
	return &CancelWorkflowResponse{WorkflowID: s.WorkflowID}, nil
}
