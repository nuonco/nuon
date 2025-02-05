package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateHealthCheckRequest struct {
	RunnerID string           `validate:"required"`
	Status   app.RunnerStatus `validate:"required"`
}

// @temporal-gen activity
// @by-id RunnerID
func (a *Activities) CreateHealthCheck(ctx context.Context, req CreateHealthCheckRequest) (*app.RunnerHealthCheck, error) {
	hc := app.RunnerHealthCheck{
		RunnerID:     req.RunnerID,
		RunnerStatus: req.Status,
	}

	if res := a.chDB.WithContext(ctx).Create(&hc); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create health check")
	}

	return &hc, nil
}
