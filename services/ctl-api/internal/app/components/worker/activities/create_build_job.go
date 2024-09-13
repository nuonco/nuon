package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateBuildJobRequest struct {
	RunnerID string
	Op       app.RunnerJobOperationType
	Type     app.RunnerJobType
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
func (a *Activities) CreateBuildJob(ctx context.Context, req *CreateBuildJobRequest) (*app.RunnerJob, error) {
	job, err := a.runnersHelpers.CreateBuildJob(ctx, req.RunnerID, req.Type, req.Op)
	if err != nil {
		return nil, fmt.Errorf("unable to create build job: %w", err)
	}

	return job, nil
}
