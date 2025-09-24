package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetQueueRequest struct {
	QueueID string `validate:"required"`
}

// @temporal-gen activity
// @by-id QueueID
func (a *Activities) GetQueue(ctx context.Context, req *GetQueueRequest) (*app.Queue, error) {
	return a.getQueue(ctx, req.QueueID)
}

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
