package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateStatusRequest struct {
	RunnerID          string           `validate:"required"`
	Status            app.RunnerStatus `validate:"required"`
	StatusDescription string           `validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
func (a *Activities) UpdateStatus(ctx context.Context, req UpdateStatusRequest) error {
	runner := app.Runner{
		ID: req.RunnerID,
	}
	res := a.db.WithContext(ctx).Model(&runner).Updates(app.Runner{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update runner status: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no runner found: %s %w", req.RunnerID, gorm.ErrRecordNotFound)
	}

	return nil
}
