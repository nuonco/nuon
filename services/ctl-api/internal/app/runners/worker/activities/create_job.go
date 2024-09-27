package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateJobRequest struct {
	RunnerID  string
	OwnerType string
	OwnerID   string
	Op        app.RunnerJobOperationType
	Type      app.RunnerJobType
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
func (a *Activities) CreateJob(ctx context.Context, req *CreateJobRequest) (*app.RunnerJob, error) {
	job, err := a.helpers.CreateRunnerJob(ctx, req.RunnerID, req.OwnerType, req.OwnerID, req.Type, req.Op)
	if err != nil {
		return nil, fmt.Errorf("unable to create runner job: %w", err)
	}

	return job, nil
}
