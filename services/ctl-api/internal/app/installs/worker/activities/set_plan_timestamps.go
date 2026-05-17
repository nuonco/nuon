package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SetDeployPlannedAtRequest struct {
	DeployID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) SetDeployPlannedAt(ctx context.Context, req SetDeployPlannedAtRequest) error {
	now := time.Now().UTC()
	res := a.db.WithContext(ctx).
		Model(&app.InstallDeploy{}).
		Where(app.InstallDeploy{ID: req.DeployID}).
		Update("planned_at", now)
	if res.Error != nil {
		return fmt.Errorf("unable to set planned_at on deploy: %w", res.Error)
	}
	return nil
}

type SetDeployAppliedAtRequest struct {
	DeployID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) SetDeployAppliedAt(ctx context.Context, req SetDeployAppliedAtRequest) error {
	now := time.Now().UTC()
	res := a.db.WithContext(ctx).
		Model(&app.InstallDeploy{}).
		Where(app.InstallDeploy{ID: req.DeployID}).
		Update("applied_at", now)
	if res.Error != nil {
		return fmt.Errorf("unable to set applied_at on deploy: %w", res.Error)
	}
	return nil
}

type SetSandboxRunPlannedAtRequest struct {
	SandboxRunID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) SetSandboxRunPlannedAt(ctx context.Context, req SetSandboxRunPlannedAtRequest) error {
	now := time.Now().UTC()
	res := a.db.WithContext(ctx).
		Model(&app.InstallSandboxRun{}).
		Where(app.InstallSandboxRun{ID: req.SandboxRunID}).
		Update("planned_at", now)
	if res.Error != nil {
		return fmt.Errorf("unable to set planned_at on sandbox run: %w", res.Error)
	}
	return nil
}

type SetSandboxRunAppliedAtRequest struct {
	SandboxRunID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) SetSandboxRunAppliedAt(ctx context.Context, req SetSandboxRunAppliedAtRequest) error {
	now := time.Now().UTC()
	res := a.db.WithContext(ctx).
		Model(&app.InstallSandboxRun{}).
		Where(app.InstallSandboxRun{ID: req.SandboxRunID}).
		Update("applied_at", now)
	if res.Error != nil {
		return fmt.Errorf("unable to set applied_at on sandbox run: %w", res.Error)
	}
	return nil
}
