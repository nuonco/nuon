package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRunnerStatusRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetRunnerStatus(ctx context.Context, req GetRunnerStatusRequest) (app.RunnerStatus, error) {
	runner, err := a.getRunner(ctx, req.ID)
	if err != nil {
		return app.RunnerStatusUnknown, fmt.Errorf("unable to get runner status: %w", err)
	}

	return runner.Status, nil
}
