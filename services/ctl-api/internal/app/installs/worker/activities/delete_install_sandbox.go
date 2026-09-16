package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type DeleteInstallSandboxRequest struct {
	InstallSandboxID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallSandboxID
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
