package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
)

type GetLatestActiveSandboxBuildRequest struct {
	AppConfigID  string `json:"app_config_id" validate:"required"`
	SandboxRunID string `json:"sandbox_run_id,omitempty"`
}

// @temporal-gen-v2 activity
// @by-field AppConfigID
// @start-to-close-timeout 30s
func (a *Activities) GetLatestActiveSandboxBuild(ctx context.Context, req GetLatestActiveSandboxBuildRequest) (*app.AppSandboxBuild, error) {
	if req.SandboxRunID == "" {
		return configdiff.LatestActiveSandboxBuild(ctx, a.db, req.AppConfigID)
	}
	var selected *app.AppSandboxBuild
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var run app.InstallSandboxRun
		if err := tx.Preload("AppSandboxConfig").Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.InstallSandboxRun{ID: req.SandboxRunID}).First(&run).Error; err != nil {
			return err
		}
		if run.AppSandboxConfig.AppConfigID != req.AppConfigID {
			return fmt.Errorf("sandbox run does not use the requested app config")
		}
		if run.AppSandboxBuildID != "" {
			var build app.AppSandboxBuild
			if err := tx.Where(app.AppSandboxBuild{ID: run.AppSandboxBuildID, AppID: run.AppSandboxConfig.AppID}).First(&build).Error; err != nil {
				return err
			}
			selected = &build
			return nil
		}
		var err error
		selected, err = configdiff.LatestActiveSandboxBuild(ctx, tx, req.AppConfigID)
		if err != nil || selected == nil {
			return err
		}
		return tx.Model(&app.InstallSandboxRun{}).Where(app.InstallSandboxRun{ID: run.ID}).
			Update("app_sandbox_build_id", selected.ID).Error
	})
	return selected, err
}
