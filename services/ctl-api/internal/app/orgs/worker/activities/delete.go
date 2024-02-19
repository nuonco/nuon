package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type DeleteRequest struct {
	OrgID string `validate:"required"`
}

func (a *Activities) Delete(ctx context.Context, req DeleteRequest) error {
	// delete apps
	res := a.db.WithContext(ctx).Unscoped().
		Where("org_id = ?", req.OrgID).
		Delete(&app.App{})
	if res.Error != nil {
		return fmt.Errorf("unable to delete org apps: %w", res.Error)
	}

	// delete org
	res = a.db.WithContext(ctx).Unscoped().Delete(&app.Org{
		ID: req.OrgID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete org: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("org not found %w", gorm.ErrRecordNotFound)
	}

	return nil
}
