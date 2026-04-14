package executeflow

import "go.temporal.io/sdk/workflow"

// RetryStepRequest is the input for the "retry-step" update handler.
type RetryStepRequest struct {
	StepID    string `json:"step_id"`
	Operation string `json:"operation"` // "retry-step" or "skip-step"
}

// RetryStepResponse is the response from the "retry-step" update handler.
type RetryStepResponse struct {
	WorkflowID string `json:"workflow_id"`
}

func (s *Signal) retryStepHandler(ctx workflow.Context, req RetryStepRequest) (*RetryStepResponse, error) {
	s.retryRequested = true
	s.retryStepID = req.StepID
	s.retryOperation = req.Operation
	return &RetryStepResponse{WorkflowID: s.InstallWorkflowID}, nil
}
