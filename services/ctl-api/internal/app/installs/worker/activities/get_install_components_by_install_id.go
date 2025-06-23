package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallComponentsRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetInstallComponents(ctx context.Context, req GetInstallComponentIDsRequest) ([]app.InstallComponent, error) {
	comps, err := a.helpers.GetInstallComponents(ctx, req.InstallID)
	if err != nil {
		return nil, generics.TemporalGormError(err, "get install components")
	}
	return comps, nil
}
