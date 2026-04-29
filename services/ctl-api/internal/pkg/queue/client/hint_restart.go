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
	if len(queueIDs) == 0 {
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if res := c.db.WithContext(ctx).
		Model(&app.Queue{}).
		Where("id IN ?", queueIDs).
		Update("metadata", c.db.Raw("COALESCE(metadata, ''::hstore) || hstore('restart_hint', ?)", now)); res.Error != nil {
		return errors.Wrap(res.Error, "unable to set restart hints")
	}

	return nil
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) HintRestartByOrg(ctx context.Context, orgID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if res := c.db.WithContext(ctx).
		Model(&app.Queue{}).
		Where("org_id = ?", orgID).
		Update("metadata", c.db.Raw("COALESCE(metadata, ''::hstore) || hstore('restart_hint', ?)", now)); res.Error != nil {
		return errors.Wrap(res.Error, "unable to set restart hints for org")
	}

	return nil
}
