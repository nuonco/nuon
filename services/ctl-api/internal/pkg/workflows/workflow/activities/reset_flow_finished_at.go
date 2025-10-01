package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type ResetFlowFinishedAtRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) PkgWorkflowsFlowResetFlowFinishedAt(ctx context.Context, req ResetFlowFinishedAtRequest) error {
	iwf := app.Workflow{
		ID: req.ID,
	}

	// temporary path to implement reset of finished_at for a specific workflow
	// ideally we'd want to do via gorm
	res := a.db.Raw("UPDATE install_workflows SET finished_at = NULL WHERE id = ? ", iwf.ID).Scan(&iwf)
	if res.Error != nil {
		return generics.TemporalGormError(res.Error)
	}

	return nil
}
