package activities

import (
	"context"

	"gorm.io/gorm"

	pkgstate "github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetLatestInstallStateRequest struct {
	InstallID string `validate:"required"`
}

// GetLatestInstallState fetches the most recent install state from DB with no side effects.
//
// @temporal-gen-v2 activity
// @by-field InstallID
// @start-to-close-timeout 10s
func (a *Activities) GetLatestInstallState(ctx context.Context, req *GetLatestInstallStateRequest) (*pkgstate.State, error) {
	var is app.InstallState
	res := a.db.WithContext(ctx).
		Where("install_id = ?", req.InstallID).
		Order("created_at DESC").
		First(&is)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}

	// Must go through the getter: the state lives in the blob, and callers use this
	// as the base they merge regenerated partials into, so an empty return here
	// silently drops every partial they did not regenerate.
	return is.GetState(ctx)
}
