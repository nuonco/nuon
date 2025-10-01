package activities

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type UpdateFlowStepStartedAtRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) PkgWorkflowsFlowUpdateFlowStepStartedAt(ctx context.Context, req UpdateFlowStepStartedAtRequest) error {
	runner := app.WorkflowStep{
		ID: req.ID,
	}
	res := a.db.WithContext(ctx).Model(&runner).Updates(app.WorkflowStep{
		StartedAt: time.Now(),
	})
	if res.Error != nil {
		return generics.TemporalGormError(gorm.ErrRecordNotFound)
	}
	if res.RowsAffected < 1 {
		return generics.TemporalGormError(gorm.ErrRecordNotFound)
	}

	return nil
}
