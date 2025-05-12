package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetHelmChartRequest struct {
	OwnerID string `validate:"required"`
}

// @temporal-gen activity
// @by-id OwnerID
func (a *Activities) GetHelmChart(ctx context.Context, req GetHelmChartRequest) (*app.HelmChart, error) {
	return a.getHelmChart(ctx, req.OwnerID)
}

func (a *Activities) getHelmChart(ctx context.Context, ownerID string) (*app.HelmChart, error) {
	helmChart := app.HelmChart{}
	res := a.db.WithContext(ctx).Model(&app.HelmChart{}).
		First(&helmChart, "owner_id = ?", ownerID)
	if res.Error != nil {
		return nil, res.Error
	}

	return &helmChart, nil
}
