package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"gorm.io/gorm/clause"
)

type DeleteInstallComponentRequest struct {
	InstallComponentID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallComponentID
func (a *Activities) DeleteInstallComponent(ctx context.Context, req DeleteInstallComponentRequest) error {
	res := a.db.WithContext(ctx).
		Select(clause.Associations).
		Delete(&app.InstallComponent{
			ID: req.InstallComponentID,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to delete install: %w", res.Error)
	}

	return nil
}
