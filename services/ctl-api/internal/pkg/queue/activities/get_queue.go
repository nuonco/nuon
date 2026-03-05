package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @as-wrapper
// @wrapper-prefix QueueInternal
// @by-field queueID
// @start-to-close-timeout 5m
func (a *Activities) getQueue(ctx context.Context, queueID string) (*app.Queue, error) {
	var queue app.Queue

	if res := a.db.WithContext(ctx).
		Where(app.Queue{
			ID: queueID,
		}).
		First(&queue); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get queue")
	}

	return &queue, nil
}
