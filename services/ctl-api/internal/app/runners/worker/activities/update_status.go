package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/runnerstatussignals"
)

type UpdateStatusRequest struct {
	RunnerID          string           `validate:"required"`
	Status            app.RunnerStatus `validate:"required"`
	StatusDescription string           `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateStatus(ctx context.Context, req UpdateStatusRequest) error {
	runner := app.Runner{
		ID: req.RunnerID,
	}
	res := a.db.WithContext(ctx).Model(&runner).Updates(app.Runner{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update runner status: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no runner found: %s %w", req.RunnerID, gorm.ErrRecordNotFound)
	}

	a.helpers.SignalOrgProvisionForRunner(ctx, zap.NewNop(), req.RunnerID, runnerstatussignals.ReasonStatusChanged)

	return nil
}
