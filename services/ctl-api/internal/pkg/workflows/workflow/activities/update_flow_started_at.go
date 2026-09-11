package activities

import (
	"context"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"gorm.io/gorm/clause"
)

type UpdateFlowStartedAtRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ID
func (a *Activities) PkgWorkflowsFlowUpdateFlowStartedAt(ctx context.Context, req UpdateFlowStartedAtRequest) error {
	runner := app.Workflow{
		ID: req.ID,
	}
	// A retried start activity must not replace the first execution timestamp.
	res := a.db.WithContext(ctx).Model(&runner).
		Where(app.Workflow{ID: req.ID}, "id").
		Where(clause.Eq{Column: "started_at", Value: nil}).
		Updates(app.Workflow{StartedAt: time.Now()})
	if res.Error != nil {
		return generics.TemporalGormError(res.Error)
	}
	if res.RowsAffected < 1 {
		return generics.TemporalGormError(a.db.WithContext(ctx).Where(app.Workflow{ID: req.ID}, "id").Take(&runner).Error)
	}

	return nil
}
