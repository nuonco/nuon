package client

import (
	"context"

	"github.com/pkg/errors"
)

func (c *Client) StaleDropInFlightSignals(ctx context.Context, reason string, staleIDs []string) error {
	if len(staleIDs) == 0 {
		return nil
	}

	if res := c.db.WithContext(ctx).Exec(`
		UPDATE queue_signals
		SET status = jsonb_set(status, '{status}', '"error"'::jsonb)
		           || jsonb_build_object('metadata', jsonb_build_object('stale_drop', ?::text)),
		    updated_at = now()
		WHERE id IN (?)`, reason, staleIDs); res.Error != nil {
		return errors.Wrap(res.Error, "unable to mark stale in-flight signals as failed")
	}

	return nil
}
