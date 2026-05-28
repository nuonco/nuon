package activities

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateFlowFinishedAtRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ID
func (a *Activities) PkgWorkflowsFlowUpdateFlowFinishedAt(ctx context.Context, req UpdateFlowFinishedAtRequest) error {
	// Load-then-Save so Workflow.BeforeSave fires and recomputes name to
	// the past-tense title once finished_at is set.
	var runner app.Workflow
	if err := a.db.WithContext(ctx).Where("id = ?", req.ID).Take(&runner).Error; err != nil {
		return generics.TemporalGormError(gorm.ErrRecordNotFound)
	}
	runner.FinishedAt = time.Now()
	if err := a.db.WithContext(ctx).Save(&runner).Error; err != nil {
		return generics.TemporalGormError(gorm.ErrRecordNotFound)
	}
	return nil
}
