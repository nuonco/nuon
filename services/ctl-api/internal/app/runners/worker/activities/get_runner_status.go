package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRunnerStatusRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
func (a *Activities) GetRunnerStatus(ctx context.Context, req GetRunnerStatusRequest) (app.RunnerStatus, error) {
	// NOTE(jm): remove this once the runner health checks are added
	return app.RunnerStatusActive, nil

	runner, err := a.getRunner(ctx, req.RunnerID)
	if err != nil {
		return app.RunnerStatusUnknown, fmt.Errorf("unable to get runner status: %w", err)
	}

	return runner.Status, nil
}
