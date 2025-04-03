package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// getInstallComponents reads components deployed to an install from the DB.
func (h *Helpers) getInstallComponents(ctx context.Context, installID string) ([]app.InstallComponent, error) {
	install := &app.Install{}
	res := h.db.WithContext(ctx).
		Preload("InstallComponents", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_components.created_at DESC")
		}).
		Preload("InstallComponents.InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(scopes.WithOverrideTable("install_deploys_latest_view_v1"))
		}).
		Preload("InstallComponents.InstallDeploys.RunnerJobs", func(db *gorm.DB) *gorm.DB {
			return db.Order("runner_jobs_view_v1.created_at DESC")
		}).
		Preload("InstallComponents.Component").
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install components: %w", res.Error)
	}

	return install.InstallComponents, nil
}
