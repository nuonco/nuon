package client

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Terminate performs full teardown of a queue:
//  1. Cancels any non-terminal queue signals (sends cancel update to their handler
//     workflows, which exits them and marks DB rows cancelled).
//  2. Stops and deletes the emitters (cancels their Temporal workflows).
//  3. Stops the queue workflow.
//  4. Soft-deletes the queue record.
//
// Order is important: cancelling signals first ensures handler workflows can exit
// cleanly via the cancel update before the queue workflow is stopped. Without
// this step, handler workflows would be left orphaned in Temporal waiting for
// `execute` updates that never come.
func (c *Client) Terminate(ctx context.Context, queueID string) error {
	// Cancel non-terminal signals first so their handler workflows exit cleanly.
	var pendingSignals []app.QueueSignal
	if res := c.db.WithContext(ctx).
		Where("queue_id = ?", queueID).
		Where("status->>'status' NOT IN (?, ?, ?)", app.StatusSuccess, app.StatusError, app.StatusCancelled).
		Find(&pendingSignals); res.Error != nil {
		c.l.Warn("unable to load pending queue signals for cancellation during terminate",
			zap.String("queue-id", queueID), zap.Error(res.Error))
	}

	for _, qs := range pendingSignals {
		if _, err := c.CancelSignal(ctx, qs.ID); err != nil {
			c.l.Warn("unable to cancel queue signal during terminate",
				zap.String("queue-id", queueID),
				zap.String("queue-signal-id", qs.ID),
				zap.Error(err))
			// Best effort: ensure DB status is cancelled even if the update failed,
			// so AwaitSignal callers don't block forever.
			c.updateQueueSignalStatus(ctx, qs.ID, app.StatusCancelled)
		}
	}

	var emitters []app.QueueEmitter
	if res := c.db.WithContext(ctx).Where("queue_id = ?", queueID).Find(&emitters); res.Error != nil {
		return errors.Wrap(res.Error, "unable to get emitters for queue")
	}

	for _, em := range emitters {
		// Cancel the emitter's Temporal workflow
		if err := c.tClient.CancelWorkflowInNamespace(ctx, em.Workflow.Namespace, em.Workflow.ID, ""); err != nil {
			c.l.Warn("unable to cancel emitter workflow during terminate", zap.String("emitter-id", em.ID), zap.Error(err))
		}

		// Delete the emitter record
		if res := c.db.WithContext(ctx).Delete(&em); res.Error != nil {
			c.l.Warn("unable to delete emitter during terminate", zap.String("emitter-id", em.ID), zap.Error(res.Error))
		}
	}

	// Stop the queue workflow
	if err := c.Stop(ctx, queueID); err != nil {
		c.l.Warn("unable to stop queue workflow during terminate", zap.String("queue-id", queueID), zap.Error(err))
	}

	// Soft-delete the queue record
	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue for soft-delete")
	}

	if res := c.db.WithContext(ctx).Delete(q); res.Error != nil {
		return errors.Wrap(res.Error, "unable to soft-delete queue")
	}

	c.l.Debug("queue terminated",
		zap.String("queue-id", queueID),
		zap.Int("cancelled-signals", len(pendingSignals)),
		zap.Int("stopped-emitters", len(emitters)))
	return nil
}
