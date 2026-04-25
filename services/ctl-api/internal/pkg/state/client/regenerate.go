package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// Regenerate checks all partials for changes and updates stale ones.
func (c *Client) Regenerate(ctx context.Context, installID string) (*statemanager.RegenerateResponse, error) {
	wfID := workflowID(installID)
	handle, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, workflowNamespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   wfID,
			UpdateName:   statemanager.RegenerateUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				statemanager.RegenerateRequest{},
			},
		},
		StartWorkflowOperation: c.stateManagerStartOperation(installID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to send regenerate to state-manager")
	}

	var resp statemanager.RegenerateResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable to get regenerate response")
	}
	return &resp, nil
}
