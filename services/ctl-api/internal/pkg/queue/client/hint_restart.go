package client

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) HintRestart(ctx context.Context, queueIDs []string) error {
	if len(queueIDs) == 0 {
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if res := c.db.WithContext(ctx).Exec(`
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
		WHERE id IN ? AND deleted_at = 0
	`, now, queueIDs); res.Error != nil {
		return errors.Wrap(res.Error, "unable to set restart hints")
	}

	return nil
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) HintRestartByOrg(ctx context.Context, orgID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if res := c.db.WithContext(ctx).Exec(`
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
		WHERE org_id = ? AND deleted_at = 0
	`, now, orgID); res.Error != nil {
		return errors.Wrap(res.Error, "unable to set restart hints for org")
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
