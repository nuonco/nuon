package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateSandboxJobRequest struct {
	InstallID string
	RunnerID  string
	OwnerType string
	OwnerID   string
	Op        app.RunnerJobOperationType
	Metadata  map[string]string
}

// @temporal-gen activity
func (a *Activities) CreateSandboxJob(ctx context.Context, req *CreateSandboxJobRequest) (*app.RunnerJob, error) {
	job, err := a.runnersHelpers.CreateInstallSandboxJob(ctx,
		req.RunnerID,
		req.OwnerType,
		req.OwnerID,
		app.RunnerJobTypeSandboxTerraform,
		req.Op,
		req.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create install sandbox job: %w", err)
	}

	return job, nil
}
