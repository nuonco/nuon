package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateBuildStatusV2 struct {
	BuildID           string                   `validate:"required"`
	Status            app.ComponentBuildStatus `validate:"required"`
	StatusDescription string                   `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) UpdateBuildStatusV2(ctx context.Context, req UpdateBuildStatusV2) error {
	currentApp := app.ComponentBuild{
		ID: req.BuildID,
	}

	compStatus := app.NewCompositeStatus(ctx, app.Status(req.Status))
	compStatus.StatusHumanDescription = req.StatusDescription
	res := a.db.WithContext(ctx).Model(&currentApp).Updates(app.ComponentBuild{
		StatusV2: compStatus,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update build: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no build found: %s %w", req.BuildID, gorm.ErrRecordNotFound)
	}

	return nil
}
