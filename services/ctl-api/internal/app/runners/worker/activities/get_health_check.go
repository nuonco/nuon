package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetHealthCheckRequest struct {
	ID     string           `validate:"required"`
	Status app.RunnerStatus `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetHealthCheck(ctx context.Context, req *GetHealthCheckRequest) (*app.RunnerHealthCheck, error) {
	var runnerHC app.RunnerHealthCheck

	res := a.chDB.WithContext(ctx).
		Where(app.RunnerHealthCheck{
			RunnerStatus: req.Status,
		}).
		Order("created_at DESC").
		First(&runnerHC, "runner_id = ?", req.ID)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, temporal.NewNonRetryableApplicationError(
				"not found",
				"not found",
				res.Error)
		}

		return nil, errors.Wrap(res.Error, "unable to get runner health check")
	}

	return &runnerHC, nil
}
