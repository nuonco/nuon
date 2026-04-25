package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// Stop terminates the state-manager workflow for an install.
func (c *Client) Stop(ctx context.Context, installID string) error {
	wfID := workflowID(installID)
	_, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, workflowNamespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   wfID,
			UpdateName:   statemanager.StopUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageAccepted,
			Args: []any{
				statemanager.StopRequest{},
			},
		},
		StartWorkflowOperation: c.stateManagerStartOperation(installID),
	})
	if err != nil {
		return errors.Wrap(err, "unable to stop state-manager")
	}
	return nil
}
