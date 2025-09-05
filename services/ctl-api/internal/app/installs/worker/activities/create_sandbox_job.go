package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateSandboxJobRequest struct {
	InstallID   string                     `validate:"required"`
	RunnerID    string                     `validate:"required"`
	OwnerType   string                     `validate:"required"`
	OwnerID     string                     `validate:"required"`
	Op          app.RunnerJobOperationType `validate:"required"`
	Metadata    map[string]string          `validate:"required"`
	LogStreamID string                     `validate:"required"`
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
		req.LogStreamID,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create install sandbox job: %w", err)
	}

	return job, nil
}
