package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// Hint sends a hint to the state-manager workflow for an install, triggering
// regeneration of the affected partials. If the workflow isn't running, it's auto-started.
func (c *Client) Hint(ctx context.Context, installID string, hint statemanager.HintType) (*statemanager.HintResponse, error) {
	wfID := workflowID(installID)
	handle, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, workflowNamespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   wfID,
			UpdateName:   statemanager.HintUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				statemanager.HintRequest{HintType: hint},
			},
		},
		StartWorkflowOperation: c.stateManagerStartOperation(installID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to send hint to state-manager")
	}

	var resp statemanager.HintResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable to get hint response")
	}
	return &resp, nil
}
