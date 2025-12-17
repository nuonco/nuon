package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetJobExecutionStatusRequest struct {
	JobExecutionID string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetJobExecutionStatus(ctx context.Context, req GetJobExecutionStatusRequest) (app.RunnerJobExecutionStatus, error) {
	jobExecution, err := a.getRunnerJobExecution(ctx, req.JobExecutionID)
	if err != nil {
		return app.RunnerJobExecutionStatusUnknown, fmt.Errorf("unable to get runner job execution: %w", err)
	}

	return jobExecution.Status, nil
}
