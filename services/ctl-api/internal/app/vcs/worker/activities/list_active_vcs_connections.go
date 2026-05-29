package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ListActiveVCSConnectionIDsRequest struct{}

// ListActiveVCSConnectionIDs returns the IDs of all VCS connections. Each
// connection had a per-connection health-check cron emitter; the sweep enqueues
// the existing healthcheck signal for each instead.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ListActiveVCSConnectionIDs(ctx context.Context, _ *ListActiveVCSConnectionIDsRequest) ([]string, error) {
	var ids []string
	if res := a.db.WithContext(ctx).
		Model(&app.VCSConnection{}).
		Pluck("id", &ids); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to list vcs connections")
	}

	return ids, nil
}
