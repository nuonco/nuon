package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

type GetInstallActionWorkflowStateRequest struct {
	InstallActionWorkflowID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallActionWorkflowID
func (a *Activities) GetInstallActionWorkflowState(ctx context.Context, req GetInstallActionWorkflowStateRequest) (*app.InstallActionWorkflow, error) {
	var act app.InstallActionWorkflow
	res := a.db.WithContext(ctx).
		Preload("Runs", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(
				scopes.WithOverrideTable(views.CustomViewName(db, &app.InstallActionWorkflowRun{}, "state_view_v1")),
			)
		}).
		Preload("Runs.RunnerJob").
		Preload("ActionWorkflow").
		Find(&act, "id= ?", req.InstallActionWorkflowID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install action workflow state")
	}

	return &act, nil
}
