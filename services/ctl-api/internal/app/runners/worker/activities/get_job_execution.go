package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetJobExecutionRequest struct {
	JobExecutionID string `validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
func (a *Activities) GetJobExecution(ctx context.Context, req GetJobExecutionRequest) (*app.RunnerJobExecution, error) {
	job, err := a.getRunnerJobExecution(ctx, req.JobExecutionID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner job execution: %w", err)
	}

	return job, nil
}

func (a *Activities) getRunnerJobExecution(ctx context.Context, jobExecutionID string) (*app.RunnerJobExecution, error) {
	jobExecution := app.RunnerJobExecution{}
	res := a.db.WithContext(ctx).
		First(&jobExecution, "id = ?", jobExecutionID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner job execution: %w", res.Error)
	}

	return &jobExecution, nil
}
