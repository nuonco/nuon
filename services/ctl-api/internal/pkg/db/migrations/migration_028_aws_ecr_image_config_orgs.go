package migrations

import (
	"context"
	"errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

func (a *Migrations) migration028AWSECRConfigs(ctx context.Context) error {
	var objs []app.AWSECRImageConfig
	res := a.db.Unscoped().WithContext(ctx).
		Find(&objs)
	if res.Error != nil {
		return res.Error
	}

	for _, obj := range objs {
		var org app.Org
		res := a.db.Unscoped().WithContext(ctx).
			First(&org, "id = ?", obj.OrgID)
		if res.Error != nil {
			if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return res.Error
			}

			res := a.db.Unscoped().WithContext(ctx).
				Delete(&obj, "id = ?", obj.ID)
			if res.Error != nil {
				return res.Error
			}
		}
	}

	return nil
}
