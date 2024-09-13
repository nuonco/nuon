package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
func (a *Activities) Get(ctx context.Context, req GetRequest) (*app.Runner, error) {
	runner, err := a.getRunner(ctx, req.RunnerID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner: %w", err)
	}

	// TODO(jm): remove this once we implement the health check
	runner.Status = app.RunnerStatusActive
	return runner, nil
}

func (a *Activities) getRunner(ctx context.Context, runnerID string) (*app.Runner, error) {
	runner := app.Runner{}
	res := a.db.WithContext(ctx).
		First(&runner, "id = ?", runnerID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner: %w", res.Error)
	}

	return &runner, nil
}
