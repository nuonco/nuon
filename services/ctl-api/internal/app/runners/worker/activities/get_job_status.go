package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetJobStatusRequest struct {
	JobID string `validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
func (a *Activities) GetJobStatus(ctx context.Context, req GetJobRequest) (app.RunnerJobStatus, error) {
	job, err := a.getRunnerJob(ctx, req.JobID)
	if err != nil {
		return app.RunnerJobStatusUnknown, fmt.Errorf("unable to get runner job: %w", err)
	}

	return job.Status, nil
}
