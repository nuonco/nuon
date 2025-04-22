package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm/clause"
)

type DeleteInstallSandboxRequest struct {
	InstallSandboxID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallSandboxID
func (a *Activities) DeleteInstallSanbox(ctx context.Context, req DeleteInstallSandboxRequest) error {
	res := a.db.WithContext(ctx).
		Select(clause.Associations).
		Delete(&app.InstallSandbox{
			ID: req.InstallSandboxID,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to delete install: %w", res.Error)
	}

	return nil
}
