package client

import (
	"context"

	"github.com/pkg/errors"
)

// Terminate hard-terminates the queue workflow via the Temporal API. This is distinct
// from Stop, which sends a graceful stop update. After termination, child handler
// workflows are also terminated; use Restart to recover a terminated queue.
func (c *Client) Terminate(ctx context.Context, queueID string, reason string) error {
	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue")
	}

	nsClient, err := c.tClient.GetNamespaceClient(q.Workflow.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to get namespace client")
	}

	if err := nsClient.TerminateWorkflow(ctx, q.Workflow.ID, "", reason); err != nil {
		return errors.Wrap(err, "unable to terminate queue workflow")
	}

	return nil
}
