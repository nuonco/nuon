package activities

import (
	"context"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type UpdateInstallWorkflowStepTargetRequest struct {
	StepID         string `validate:"required"`
	StepTargetID   string `validate:"required"`
	StepTargetType string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) UpdateInstallWorkflowStepTarget(ctx context.Context, req UpdateInstallWorkflowStepTargetRequest) error {
	step := app.InstallWorkflowStep{
		ID: req.StepID,
	}

	res := a.db.WithContext(ctx).
		Model(&step).
		Updates(app.InstallWorkflowStep{
			StepTargetID:   req.StepTargetID,
			StepTargetType: req.StepTargetType,
		})
	if res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to update install action workflow run")
	}
	if res.RowsAffected < 1 {
		return generics.TemporalGormError(gorm.ErrRecordNotFound, "no update found")
	}
	return nil
}
