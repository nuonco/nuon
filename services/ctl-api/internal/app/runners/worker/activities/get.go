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
// @by-id RunnerID
func (a *Activities) Get(ctx context.Context, req GetRequest) (*app.Runner, error) {
	runner, err := a.getRunner(ctx, req.RunnerID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner: %w", err)
	}

	return runner, nil
}

func (a *Activities) getRunner(ctx context.Context, runnerID string) (*app.Runner, error) {
	runner := app.Runner{}
	res := a.db.WithContext(ctx).
		Preload("Org").
		Preload("Org.CreatedBy").
		Preload("Org.RunnerGroup").
		Preload("Org.RunnerGroup.Runners").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Settings").
		First(&runner, "id = ?", runnerID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner: %w", res.Error)
	}

	return &runner, nil
}
