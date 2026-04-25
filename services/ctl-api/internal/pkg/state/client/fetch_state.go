package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	pkgstate "github.com/nuonco/nuon/pkg/types/state"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// FetchState returns the current cached state from the state-manager workflow.
// If the workflow isn't running, it's auto-started. If no cached state exists, the
// workflow performs a full regeneration before returning.
// This is the primary way to get state from within other Temporal workflows.
func (c *Client) FetchState(ctx context.Context, installID string) (*pkgstate.State, error) {
	wfID := workflowID(installID)
	handle, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, workflowNamespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   wfID,
			UpdateName:   statemanager.FetchStateUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				statemanager.FetchStateRequest{},
			},
		},
		StartWorkflowOperation: c.stateManagerStartOperation(installID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to send fetch-state to state-manager")
	}

	var resp statemanager.FetchStateResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable to get fetch-state response")
	}
	return resp.State, nil
}
