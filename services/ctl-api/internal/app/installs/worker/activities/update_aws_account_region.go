package activities

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateAWSAccountRegion struct {
	InstallID string `validate:"required"`
	Region    string `validate:"required"`
}

// UpdateAWSAccountRegion writes the customer-chosen AWS region back to the
// install's aws_account row, so reads of `install.AWSAccount.Region` see the
// real value once phone-home has settled. Region is now optional at install
// creation, so the field starts empty and must be backfilled from stack
// outputs. Mirrors UpdateGCPAccountRegion.
//
// @temporal-gen-v2 activity
func (a *Activities) UpdateAWSAccountRegion(ctx context.Context, req *UpdateAWSAccountRegion) error {
	res := a.db.WithContext(ctx).
		Model(&app.AWSAccount{}).
		Where("install_id = ?", req.InstallID).
		Updates(map[string]interface{}{
			"region": req.Region,
		})
	if res.Error != nil {
		return generics.TemporalGormError(res.Error)
	}
	if res.RowsAffected < 1 {
		return generics.TemporalGormError(gorm.ErrRecordNotFound)
	}

	return nil
}
