package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetActiveInstallStackVersionRequest struct {
	InstallID string `json:"install_id" validate:"required"`
}

// GetActiveInstallStackVersion returns the install's newest active stack version, or nil when none is active.
// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) GetActiveInstallStackVersion(ctx context.Context, req GetActiveInstallStackVersionRequest) (*app.InstallStackVersion, error) {
	var versions []app.InstallStackVersion
	if res := a.db.WithContext(ctx).
		Where(app.InstallStackVersion{InstallID: req.InstallID}).
		Where("status->>'status' = ?", app.InstallStackVersionStatusActive).
		Order("created_at DESC").
		Limit(1).
		Find(&versions); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get active install stack version")
	}
	if len(versions) == 0 {
		return nil, nil
	}
	return &versions[0], nil
}
