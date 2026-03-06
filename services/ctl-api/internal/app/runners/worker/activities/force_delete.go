package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ForceDeleteRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) ForceDelete(ctx context.Context, req ForceDeleteRequest) error {
	res := a.db.WithContext(ctx).
		Unscoped().
		Select(clause.Associations).
		Delete(&app.Runner{
			ID: req.RunnerID,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to force delete runner: %w", res.Error)
	}

	return nil
}
