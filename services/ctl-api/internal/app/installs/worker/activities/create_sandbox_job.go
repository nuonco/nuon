package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

type CreateSandboxJobRequest struct {
	InstallID string
	RunnerID  string
	OwnerType string
	OwnerID   string
	Op        app.RunnerJobOperationType
}

// @temporal-gen activity
func (a *Activities) CreateSandboxJob(ctx context.Context, req *CreateSandboxJobRequest) (*app.RunnerJob, error) {
	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	ctx = middlewares.SetAccountIDContext(ctx, install.CreatedByID)
	ctx = middlewares.SetOrgIDContext(ctx, install.OrgID)

	job, err := a.runnersHelpers.CreateInstallSandboxJob(ctx, req.RunnerID, req.OwnerType, req.OwnerID, app.RunnerJobTypeSandboxTerraform, req.Op)
	if err != nil {
		return nil, fmt.Errorf("unable to create install sandbox job: %w", err)
	}

	return job, nil
}
