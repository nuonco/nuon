package activities

import (
	"context"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type ResetFlowFinishedAtRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ID
func (a *Activities) PkgWorkflowsFlowResetFlowFinishedAt(ctx context.Context, req ResetFlowFinishedAtRequest) error {
	// Load-then-Save so Workflow.BeforeSave fires and recomputes name back
	// to its in-progress form when finished_at goes to NULL.
	var iwf app.Workflow
	if err := a.db.WithContext(ctx).Where("id = ?", req.ID).Take(&iwf).Error; err != nil {
		return generics.TemporalGormError(err)
	}
	iwf.FinishedAt = time.Time{}
	if err := a.db.WithContext(ctx).Save(&iwf).Error; err != nil {
		return generics.TemporalGormError(err)
	}
	return nil
}
