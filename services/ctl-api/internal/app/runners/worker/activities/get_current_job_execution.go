package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetCurrentJobExecutionRequest struct {
	JobID string `validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
// @by-id JobID
func (a *Activities) GetCurrentJobExecution(ctx context.Context, req GetCurrentJobExecutionRequest) (*app.RunnerJobExecution, error) {
	jobExecution, err := a.getCurrentJobExecution(ctx, req.JobID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner job execution: %w", err)
	}

	return jobExecution, nil
}

func (a *Activities) getCurrentJobExecution(ctx context.Context, jobID string) (*app.RunnerJobExecution, error) {
	jobExecution := app.RunnerJobExecution{}
	res := a.db.WithContext(ctx).
		Where(app.RunnerJobExecution{
			RunnerJobID: jobID,
		}).
		Order("created_at desc").
		Limit(1).
		First(&jobExecution)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get current job execution: %w", res.Error)
	}

	return &jobExecution, nil
}
