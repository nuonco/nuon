package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// ForceRegenerate triggers a full rebuild of all state partials.
func (c *Client) ForceRegenerate(ctx context.Context, installID string) (*statemanager.ForceRegenerateResponse, error) {
	wfID := workflowID(installID)
	handle, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, workflowNamespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   wfID,
			UpdateName:   statemanager.ForceRegenerateUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				statemanager.ForceRegenerateRequest{},
			},
		},
		StartWorkflowOperation: c.stateManagerStartOperation(installID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to send force-regenerate to state-manager")
	}

	var resp statemanager.ForceRegenerateResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable to get force-regenerate response")
	}
	return &resp, nil
}
