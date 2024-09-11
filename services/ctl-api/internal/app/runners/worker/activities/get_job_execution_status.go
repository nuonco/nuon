package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetJobExecutionStatusRequest struct {
	JobExecutionID string `validate:"required"`
}

// @await-gen
// @schedule-to-close-timeout 5s
func (a *Activities) GetJobExecutionStatus(ctx context.Context, req GetJobExecutionStatusRequest) (app.RunnerJobExecutionStatus, error) {
	jobExecution, err := a.getRunnerJobExecution(ctx, req.JobExecutionID)
	if err != nil {
		return app.RunnerJobExecutionStatusUnknown, fmt.Errorf("unable to get runner job execution: %w", err)
	}

	return jobExecution.Status, nil
}
