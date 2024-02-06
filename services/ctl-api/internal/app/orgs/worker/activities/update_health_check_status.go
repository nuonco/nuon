package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type UpdateHealthCheckStatusRequest struct {
	OrgHealthCheckID  string                   `validate:"required"`
	Status            app.OrgHealthCheckStatus `validate:"required"`
	StatusDescription string                   `validate:"required"`
}

func (a *Activities) UpdateHealthCheckStatus(ctx context.Context, req UpdateHealthCheckStatusRequest) error {
	org := app.OrgHealthCheck{
		ID: req.OrgHealthCheckID,
	}
	res := a.db.WithContext(ctx).Model(&org).Updates(app.OrgHealthCheck{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update org health check: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no org health check found: %s %w", req.OrgHealthCheckID, gorm.ErrRecordNotFound)
	}

	return nil
}
