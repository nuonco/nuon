package helpers

import (
	"context"

	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// getInstallActionWorkflows reads action workflows DB.
func (h *Helpers) getInstallActionWorkflows(ctx context.Context, installID string) ([]app.InstallActionWorkflow, error) {
	var acts []app.InstallActionWorkflow
	res := h.db.WithContext(ctx).
		Preload("Runs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_action_workflow_runs.created_at DESC").
				Limit(5)
		}).
		Preload("Runs.RunnerJob").
		Preload("ActionWorkflow").
		Find(&acts, "install_id = ?", installID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install components")
	}

	return acts, nil
}
