package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateDeployJobRequest struct {
	RunnerID  string
	OwnerType string
	OwnerID   string
	Op        app.RunnerJobOperationType
	Type      app.RunnerJobType
}

// @temporal-gen activity
func (a *Activities) CreateDeployJob(ctx context.Context, req *CreateDeployJobRequest) (*app.RunnerJob, error) {
	job, err := a.runnersHelpers.CreateDeployJob(ctx, req.RunnerID, req.Type, req.Op)
	if err != nil {
		return nil, fmt.Errorf("unable to create deploy job: %w", err)
	}

	return job, nil
}
