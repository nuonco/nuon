package activities

import (
	"context"
	"fmt"
)

type TerminateWorkflowRequest struct {
	WorkflowID string `validate:"required"`
	Namespace  string `validate:"required"`
	Reason     string
}

// @temporal-gen-v2 activity
func (a *Activities) TerminateWorkflow(ctx context.Context, req *TerminateWorkflowRequest) error {
	client, err := a.tClient.GetNamespaceClient(req.Namespace)
	if err != nil {
		return fmt.Errorf("unable to get namespace client for %s: %w", req.Namespace, err)
	}
	return client.TerminateWorkflow(ctx, req.WorkflowID, "", req.Reason)
}
