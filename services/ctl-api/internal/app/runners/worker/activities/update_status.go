package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateStatusRequest struct {
	RunnerID          string           `validate:"required"`
	Status            app.RunnerStatus `validate:"required"`
	StatusDescription string           `validate:"required"`
}

// @temporal-gen-v2 activity
// @max-retries 2
// @local
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
		return generics.TemporalGormError(
			gorm.ErrRecordNotFound,
			fmt.Sprintf("no runner found: %s", req.RunnerID),
		)
	}

	return nil
}
