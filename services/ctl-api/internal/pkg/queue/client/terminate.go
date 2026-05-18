package client

import (
	"context"

	"github.com/pkg/errors"
	"go.temporal.io/api/serviceerror"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (c *Client) TerminateStrict(ctx context.Context, queueID string) error {
	var emitters []app.QueueEmitter
	if res := c.db.WithContext(ctx).Where("queue_id = ?", queueID).Find(&emitters); res.Error != nil {
		return errors.Wrap(res.Error, "unable to get emitters for queue")
	}

	for _, em := range emitters {
		if err := c.tClient.CancelWorkflowInNamespace(ctx, em.Workflow.Namespace, em.Workflow.ID, ""); err != nil {
			var notFoundErr *serviceerror.NotFound
			if !errors.As(err, &notFoundErr) {
				return errors.Wrap(err, "unable to cancel emitter workflow")
			}
		}
		if res := c.db.WithContext(ctx).Delete(&em); res.Error != nil {
			return errors.Wrap(res.Error, "unable to delete emitter")
		}
	}

	if err := c.Stop(ctx, queueID); err != nil {
		var notFoundErr *serviceerror.NotFound
		if !errors.Is(err, gorm.ErrRecordNotFound) && !errors.As(err, &notFoundErr) {
			return errors.Wrap(err, "unable to stop queue workflow")
		}
	}

	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return errors.Wrap(err, "unable to get queue for soft-delete")
	}
	if res := c.db.WithContext(ctx).Delete(q); res.Error != nil {
		return errors.Wrap(res.Error, "unable to soft-delete queue")
	}

	c.l.Debug("queue terminated (strict)", zap.String("queue-id", queueID))
	return nil
}

// Terminate stops all emitters (cancelling their Temporal workflows), deletes the emitter
// records, stops the queue workflow, and soft-deletes the queue record.
func (c *Client) Terminate(ctx context.Context, queueID string) error {
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

	c.l.Debug("queue terminated", zap.String("queue-id", queueID))
	return nil
}
