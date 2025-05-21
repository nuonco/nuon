package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type GetHelmChartRequest struct {
	OwnerID string `validate:"required"`
}

// @temporal-gen activity
// @by-id OwnerID
func (a *Activities) GetHelmChart(ctx context.Context, req GetHelmChartRequest) (*app.HelmChart, error) {
	return a.getOrCreateHelmChart(ctx, req.OwnerID)
}

func (a *Activities) getOrCreateHelmChart(ctx context.Context, ownerID string) (*app.HelmChart, error) {
	helmChart := app.HelmChart{}
	res := a.db.WithContext(ctx).Model(&app.HelmChart{}).
		First(&helmChart, "owner_id = ?", ownerID)

	if res.Error == nil {
		return &helmChart, nil
	}

	if res.Error != gorm.ErrRecordNotFound {
		return nil, res.Error
	}

	helmChart = app.HelmChart{
		OwnerID:   ownerID,
		OwnerType: "install_components",
	}
	if err := a.db.WithContext(ctx).Create(&helmChart).Error; err != nil {
		return nil, err
	}

	if err := a.db.WithContext(ctx).First(&helmChart, "owner_id = ?", ownerID).Error; err != nil {
		return nil, err
	}

	return &helmChart, nil
}
