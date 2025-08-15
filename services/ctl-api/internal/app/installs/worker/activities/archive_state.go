package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type ArchiveStateRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) ArchiveState(ctx context.Context, req *ArchiveStateRequest) error {
	retainCount := 50 // Number of states to retain

	err := a.db.WithContext(ctx).
		Where("install_id = ? AND id NOT IN (?)",
			req.InstallID,
			a.db.Model(&app.InstallState{}).
				Select("id").
				Where("install_id = ?", req.InstallID).
				Order("created_at DESC").
				Limit(retainCount),
		).
		Select("archived", "state").
		Updates(&app.InstallState{
			Archived: true,
			State:    nil, // Clear the state to avoid keeping large data in archived states
		}).Error
	if err != nil {
		return err
	}

	return nil
}
