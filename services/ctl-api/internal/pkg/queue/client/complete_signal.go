package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) CompleteSignal(ctx context.Context, signalID, updateName string) error {
	var signal app.QueueSignal
	if err := c.db.WithContext(ctx).First(&signal, "id = ?", signalID).Error; err != nil {
		return errors.Wrap(err, "unable to get queue signal")
	}

	// The signal runs in a specific workflow.
	// We should use that workflow ID to send the update.
	_, err := c.tClient.UpdateWorkflowInNamespace(ctx, signal.Workflow.Namespace, tclient.UpdateWorkflowOptions{
		WorkflowID:   signal.Workflow.ID,
		UpdateName:   updateName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
	})
	if err != nil {
		return errors.Wrap(err, "unable to complete signal")
	}

	return nil
}
