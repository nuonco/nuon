package activities

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateGCPAccountRegion struct {
	InstallID string `validate:"required"`
	Region    string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateGCPAccountRegion(ctx context.Context, req *UpdateGCPAccountRegion) error {
	res := a.db.WithContext(ctx).
		Model(&app.GCPAccount{}).
		Where("install_id = ?", req.InstallID).
		Update("region", req.Region)
	if res.Error != nil {
		return generics.TemporalGormError(res.Error)
	}
	if res.RowsAffected < 1 {
		return generics.TemporalGormError(gorm.ErrRecordNotFound)
	}

	return nil
}
