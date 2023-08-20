package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateBuildStatus struct {
	BuildID           string `validate:"required"`
	Status            string `validate:"required"`
	StatusDescription string `validate:"required"`
}

func (a *Activities) UpdateBuildStatus(ctx context.Context, req UpdateBuildStatus) error {
	currentApp := app.ComponentBuild{
		ID: req.BuildID,
	}
	res := a.db.WithContext(ctx).Model(&currentApp).Updates(app.ComponentBuild{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update build: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no build found: %s", req.BuildID)
	}

	return nil
}
