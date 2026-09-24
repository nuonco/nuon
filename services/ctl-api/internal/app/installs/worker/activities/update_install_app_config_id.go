package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateInstallAppConfigIDInput struct {
	InstallID      string `json:"install_id" validate:"required"`
	NewAppConfigID string `json:"new_app_config_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) UpdateInstallAppConfigID(ctx context.Context, input *UpdateInstallAppConfigIDInput) error {
	res := a.db.WithContext(ctx).
		Model(&app.Install{}).
		Where(app.Install{ID: input.InstallID}).
		Updates(map[string]interface{}{
			"app_config_id": input.NewAppConfigID,
			// jsonb_build_object is variadic "any", so the parameter needs an
			// explicit cast for postgres to infer a type at parse time.
			"app_config_ref": gorm.Expr(
				"COALESCE(NULLIF(app_config_ref, 'null'::jsonb), '{}'::jsonb) || jsonb_build_object('expected_config_id', ?::text)",
				input.NewAppConfigID,
			),
		})
	if res.Error != nil {
		return fmt.Errorf("unable to update install app_config_id: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("install not found: %s %w", input.InstallID, gorm.ErrRecordNotFound)
	}

	return nil
}
