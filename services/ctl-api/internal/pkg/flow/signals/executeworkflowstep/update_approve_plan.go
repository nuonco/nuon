package executeworkflowstep

import "go.temporal.io/sdk/workflow"

// ApprovePlanRequest is the input for the "approve-plan" update handler.
type ApprovePlanRequest struct {
	ApprovalResponseID string `json:"approval_response_id"`
}

func (s *Signal) approvePlanHandler(ctx workflow.Context, req ApprovePlanRequest) error {
	s.approved = true
	return nil
}
