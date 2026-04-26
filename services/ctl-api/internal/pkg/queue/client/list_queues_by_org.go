package client

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) ListQueuesByOrgID(ctx context.Context, orgID string) ([]app.Queue, error) {
	var queues []app.Queue
	if res := c.db.WithContext(ctx).Where(&app.Queue{OrgID: &orgID}).Find(&queues); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to list queues by org")
	}
	return queues, nil
}
