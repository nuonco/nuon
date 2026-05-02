package client

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) HintRestart(ctx context.Context, queueIDs []string) error {
	for _, queueID := range queueIDs {
		var queue app.Queue
		if res := c.db.WithContext(ctx).
			Where(app.Queue{ID: queueID}).
			First(&queue); res.Error != nil {
			return errors.Wrap(res.Error, "unable to find queue for restart hint")
		}

		if queue.OwnerType != "runners" || queue.OwnerID == "" {
			continue
		}

		if res := c.db.WithContext(ctx).
			Model(&app.Runner{ID: queue.OwnerID}).
			Update("restart_requested", true); res.Error != nil {
			return errors.Wrap(res.Error, "unable to set restart_requested on runner")
		}
	}

	return nil
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) HintRestartByOrg(ctx context.Context, orgID string) error {
	// Set restart_requested on all runners in the org.
	if res := c.db.WithContext(ctx).
		Model(&app.Runner{}).
		Where(app.Runner{OrgID: orgID}).
		Update("restart_requested", true); res.Error != nil {
		return errors.Wrap(res.Error, "unable to set restart_requested for org runners")
	}

	return nil
}

type RequestCANAllRequest struct{}

type RequestCANAllResponse struct {
	RowsAffected int64 `json:"rows_affected"`
}

// RequestCANAll sets restart_hint on all queues via status_v2 metadata.
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) RequestCANAll(ctx context.Context, _ *RequestCANAllRequest) (*RequestCANAllResponse, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res := c.db.WithContext(ctx).Exec(`
		UPDATE queues
		SET status_v2 = jsonb_set(
			jsonb_set(
				COALESCE(status_v2::jsonb, '{}'::jsonb),
				'{metadata}',
				COALESCE(status_v2::jsonb -> 'metadata', '{}'::jsonb)
			),
			'{metadata,restart_hint}',
			to_jsonb(?::text)
		)
		WHERE deleted_at = 0
	`, now)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to set restart hint on all queues")
	}

	return &RequestCANAllResponse{RowsAffected: res.RowsAffected}, nil
}
