package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm/clause"
)

type DeleteRequest struct {
	InstallID string `validate:"required"`
}

func (a *Activities) Delete(ctx context.Context, req DeleteRequest) error {
	res := a.db.WithContext(ctx).
		Select(clause.Associations).
		Delete(&app.Install{
			ID: req.InstallID,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to delete install: %w", res.Error)
	}

	return nil
}
