package activities

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateWarningsRequest struct {
	RunnerID string         `validate:"required"`
	Warnings pq.StringArray `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateWarnings(ctx context.Context, req UpdateWarningsRequest) error {
	runner := app.Runner{
		ID: req.RunnerID,
	}
	res := a.db.WithContext(ctx).Model(&runner).Update("warnings", req.Warnings)
	if res.Error != nil {
		return fmt.Errorf("unable to update runner warnings: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no runner found: %s %w", req.RunnerID, gorm.ErrRecordNotFound)
	}

	return nil
}
