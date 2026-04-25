package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"

	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// EnsureStateManager starts the state-manager workflow for an install, or restarts it if already running.
func (c *Client) EnsureStateManager(ctx context.Context, installID string) error {
	wfID := workflowID(installID)
	c.l.Debug("ensuring state-manager workflow", zap.String("install_id", installID), zap.String("workflow_id", wfID))

	_, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, workflowNamespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   wfID,
			UpdateName:   statemanager.RestartUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageAccepted,
			Args: []any{
				statemanager.RestartRequest{},
			},
		},
		StartWorkflowOperation: c.stateManagerStartOperation(installID),
	})
	if err != nil {
		return errors.Wrap(err, "unable to ensure state-manager workflow")
	}

	return nil
}
