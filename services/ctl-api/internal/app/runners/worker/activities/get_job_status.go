package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetJobStatusRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetJobStatus(ctx context.Context, req GetJobStatusRequest) (app.RunnerJobStatus, error) {
	job, err := a.getRunnerJob(ctx, req.ID)
	if err != nil {
		return app.RunnerJobStatusUnknown, fmt.Errorf("unable to get runner job: %w", err)
	}

	return job.Status, nil
}
