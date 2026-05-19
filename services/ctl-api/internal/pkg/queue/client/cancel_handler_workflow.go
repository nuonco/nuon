package client

import (
	"context"

	"github.com/pkg/errors"
	"go.temporal.io/api/serviceerror"
)

// CancelHandlerWorkflow cancels a queue-signal handler workflow. Used by the
// emitter when stale-dropping signals so the handler exits cleanly instead of
// looping on a now-soft-deleted row. NotFound is treated as success.
func (c *Client) CancelHandlerWorkflow(ctx context.Context, namespace, workflowID string) error {
	if err := c.tClient.CancelWorkflowInNamespace(ctx, namespace, workflowID, ""); err != nil {
		var notFoundErr *serviceerror.NotFound
		if errors.As(err, &notFoundErr) {
			return nil
		}
		return errors.Wrap(err, "unable to cancel handler workflow")
	}
	return nil
}
