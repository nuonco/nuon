package client

import (
	"context"

	"github.com/pkg/errors"

	pkgstate "github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GetState reads the latest cached state from the install_states table.
// This is a pure DB read, not a Temporal operation.
// Use this for HTTP handlers and non-workflow contexts where last-persisted state is fine.
func (c *Client) GetState(ctx context.Context, installID string) (*pkgstate.State, error) {
	var is app.InstallState
	res := c.db.WithContext(ctx).
		Where(app.InstallState{InstallID: installID}).
		Order("created_at DESC").
		First(&is)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to find install state")
	}

	if !is.StaleAt.Empty() {
		is.State.StaleAt = &is.StaleAt.Time
	}

	// Hydrate labels fresh — they are mutable and not persisted in the state snapshot.
	if err := c.hydrateLabels(ctx, installID, is.State); err != nil {
		return nil, errors.Wrap(err, "unable to hydrate labels")
	}

	return is.State, nil
}

func (c *Client) hydrateLabels(ctx context.Context, installID string, is *pkgstate.State) error {
	var install app.Install
	if err := c.db.WithContext(ctx).Select("id", "labels").First(&install, "id = ?", installID).Error; err != nil {
		return errors.Wrap(err, "unable to get install labels")
	}
	is.Labels = map[string]string(install.Labels)
	return nil
}
