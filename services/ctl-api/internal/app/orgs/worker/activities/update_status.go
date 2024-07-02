package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateStatusRequest struct {
	OrgID             string        `validate:"required"`
	Status            app.OrgStatus `validate:"required"`
	StatusDescription string        `validate:"required"`
}

func (a *Activities) UpdateStatus(ctx context.Context, req UpdateStatusRequest) error {
	org := app.Org{
		ID: req.OrgID,
	}
	res := a.db.WithContext(ctx).Model(&org).Updates(app.Org{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update org: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no org found: %s %w", req.OrgID, gorm.ErrRecordNotFound)
	}

	return nil
}
