package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type SetHealthCheckRunnerJobRequest struct {
	HealthCheckID string `validate:"required"`
	RunnerJobID   string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) SetHealthCheckRunnerJob(ctx context.Context, req SetHealthCheckRunnerJobRequest) error {
	hc := app.RunnerHealthCheck{
		ID: req.HealthCheckID,
	}

	job, err := a.helpers.GetRunnerJob(ctx, req.RunnerJobID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner job")
	}

	res := a.chDB.WithContext(ctx).Model(&hc).Updates(app.RunnerHealthCheck{
		RunnerJob: *job,
	})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update healthcheck with runner job")
	}

	return nil
}
