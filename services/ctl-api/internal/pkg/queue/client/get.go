package client

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (c *Client) GetQueue(ctx context.Context, id string) (*app.Queue, error) {
	queue, err := c.getQueue(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	return queue, nil
}

func (c *Client) getQueue(ctx context.Context, id string) (*app.Queue, error) {
	var q app.Queue
	if res := c.db.WithContext(ctx).First(&q, "id = ?", id); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get queue")
	}

	return &q, nil
}

func (c *Client) getQueueStatus(ctx context.Context, id string) (*app.Queue, error) {
	return nil, nil
}
