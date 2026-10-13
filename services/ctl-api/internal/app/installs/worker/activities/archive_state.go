package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type ArchiveStateRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) ArchiveState(ctx context.Context, req *ArchiveStateRequest) error {
	retainCount := 50

	err := a.db.WithContext(ctx).
		Scopes(scopes.WithDisableViews).
		Where("install_id = ? AND archived = false AND id NOT IN (?) AND id NOT IN (?)",
			req.InstallID,
			a.db.Model(&app.InstallState{}).
				Select("id").
				Where("install_id = ?", req.InstallID).
				Order("created_at DESC").
				Limit(retainCount),
			a.db.Model(&app.InstallStateHead{}).
				Select("current_state_id").Where(app.InstallStateHead{InstallID: req.InstallID}),
		).
		Select("archived", "state").
		Updates(&app.InstallState{
			Archived: true,
			State:    nil,
		}).Error
	if err != nil {
		return err
	}

	return nil
}
