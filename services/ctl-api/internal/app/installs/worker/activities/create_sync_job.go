package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateSyncJobRequest struct {
	RunnerID string
	Op       app.RunnerJobOperationType
	Type     app.RunnerJobType
}

// @temporal-gen activity
func (a *Activities) CreateSyncJob(ctx context.Context, req *CreateSyncJobRequest) (*app.RunnerJob, error) {
	job, err := a.runnersHelpers.CreateSyncJob(ctx, req.RunnerID, req.Type, req.Op)
	if err != nil {
		return nil, fmt.Errorf("unable to create install sandbox job: %w", err)
	}

	return job, nil
}
