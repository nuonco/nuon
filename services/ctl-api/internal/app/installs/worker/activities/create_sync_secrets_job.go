package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateSyncSecretsJobRequest struct {
	InstallID string                     `validate:"required"`
	RunnerID  string                     `validate:"required"`
	OwnerType string                     `validate:"required"`
	OwnerID   string                     `validate:"required"`
	Op        app.RunnerJobOperationType `validate:"required"`
	Metadata  map[string]string          `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateSyncSecretsJob(ctx context.Context, req *CreateSyncSecretsJobRequest) (*app.RunnerJob, error) {
	job, err := a.runnersHelpers.CreateSyncSecretsJob(ctx,
		req.RunnerID,
		req.OwnerType,
		req.OwnerID,
		app.RunnerJobTypeSandboxSyncSecrets,
		req.Op,
		req.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create install sandbox job: %w", err)
	}

	return job, nil
}
