package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type DeleteRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
func (a *Activities) Delete(ctx context.Context, req DeleteRequest) error {
	res := a.db.WithContext(ctx).
		Select(clause.Associations).
		Delete(&app.Runner{
			ID: req.RunnerID,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to delete runner: %w", res.Error)
	}

	return nil
}
