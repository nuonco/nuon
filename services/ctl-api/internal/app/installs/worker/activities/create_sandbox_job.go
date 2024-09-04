package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateSandboxJobRequest struct {
	RunnerID string
	Op       app.RunnerJobOperationType
}

// @await-gen
func (a *Activities) CreateSandboxJob(ctx context.Context, req *CreateSandboxJobRequest) (*app.RunnerJob, error) {
	job, err := a.runnersHelpers.CreateInstallSandboxJob(ctx, req.RunnerID, app.RunnerJobTypeSandboxTerraform, req.Op)
	if err != nil {
		return nil, fmt.Errorf("unable to create install sandbox job: %w", err)
	}

	return job, nil
}
