package helpers

import (
	"context"

	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// getInstallActionWorkflows reads action workflows DB.
func (h *Helpers) getInstallActionWorkflows(ctx context.Context, installID string, withRunnerJobOutputs bool) ([]app.InstallActionWorkflow, error) {
	var acts []app.InstallActionWorkflow
	query := h.db.WithContext(ctx).
		Preload("Runs", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(
				scopes.WithOverrideTable(views.CustomViewName(db, &app.InstallActionWorkflowRun{}, "state_view_v1")),
			)
		}).
		Preload("ActionWorkflow")
	if withRunnerJobOutputs {
		query = query.Preload("Runs.RunnerJob", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "owner_id", "owner_type", "type", "outputs")
		})
	}
	res := query.Find(&acts, "install_id = ?", installID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install components")
	}

	return acts, nil
}
