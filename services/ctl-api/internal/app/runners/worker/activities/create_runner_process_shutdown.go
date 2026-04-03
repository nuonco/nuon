package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateRunnerProcessShutdownRequest struct {
	RunnerProcessID string                        `validate:"required"`
	Type            app.RunnerProcessShutdownType `validate:"required"`
	CompositeStatus app.CompositeStatus
}

// @temporal-gen-v2 activity
// @by-field RunnerProcessID
func (a *Activities) CreateRunnerProcessShutdown(ctx context.Context, req CreateRunnerProcessShutdownRequest) (*app.RunnerProcessShutdown, error) {
	shutdown := app.RunnerProcessShutdown{
		RunnerProcessID: req.RunnerProcessID,
		Type:            req.Type,
		CompositeStatus: req.CompositeStatus,
	}

	if res := a.db.WithContext(ctx).Create(&shutdown); res.Error != nil {
		return nil, fmt.Errorf("unable to create runner process shutdown: %w", res.Error)
	}

	return &shutdown, nil
}
