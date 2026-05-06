package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallComponentsRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) GetInstallComponents(ctx context.Context, req GetInstallComponentsRequest) ([]app.InstallComponent, error) {
	var components []app.InstallComponent
	res := a.db.WithContext(ctx).
		Where(app.InstallComponent{
			InstallID: req.InstallID,
		}).
		Preload("Component").
		Find(&components)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "get install components")
	}

	return components, nil
}
