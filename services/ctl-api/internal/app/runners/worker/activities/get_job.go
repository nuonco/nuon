package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetJobRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
// @by-id ID
func (a *Activities) GetJob(ctx context.Context, req GetJobRequest) (*app.RunnerJob, error) {
	job, err := a.getRunnerJob(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner job: %w", err)
	}

	return job, nil
}

func (a *Activities) getRunnerJob(ctx context.Context, jobID string) (*app.RunnerJob, error) {
	runnerJob := app.RunnerJob{}
	res := a.db.WithContext(ctx).
		Preload("Org").
		First(&runnerJob, "id = ?", jobID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get runner job: %w")
	}

	return &runnerJob, nil
}
