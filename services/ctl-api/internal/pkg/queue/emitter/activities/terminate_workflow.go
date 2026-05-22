package activities

import (
	"context"
)

type TerminateWorkflowRequest struct {
	WorkflowID string `validate:"required"`
	Reason     string
}

// @temporal-gen-v2 activity
func (a *Activities) TerminateWorkflow(ctx context.Context, req *TerminateWorkflowRequest) error {
	return a.tClient.TerminateWorkflow(ctx, req.WorkflowID, "", req.Reason)
}
