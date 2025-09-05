package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallForInstallComponentRequest struct {
	InstallComponentID string `json:"component_id" validate:"required"`
}

// @temporal-gen activity
// @by-id InstallComponentID
func (a *Activities) GetInstallForInstallComponent(ctx context.Context, req GetInstallForInstallComponentRequest) (*app.Install, error) {
	var component app.InstallComponent

	res := a.db.WithContext(ctx).
		Where(app.InstallComponent{
			ID: req.InstallComponentID,
		}).
		First(&component)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get install component")
	}

	return a.getInstall(ctx, component.InstallID)
}
