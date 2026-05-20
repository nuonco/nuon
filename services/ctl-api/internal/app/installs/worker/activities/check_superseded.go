package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type CheckSandboxRunSupersededRequest struct {
	SandboxRunID string `validate:"required"`
}

// CheckSandboxRunSuperseded checks whether a newer sandbox run has been applied
// for the same sandbox since this run's plan was created.
//
// Returns false (not superseded) if the run has no planned_at set (backward
// compatible with records created before this field existed).
//
// @temporal-gen-v2 activity
func (a *Activities) CheckSandboxRunSuperseded(ctx context.Context, req CheckSandboxRunSupersededRequest) (bool, error) {
	var run app.InstallSandboxRun
	if err := a.db.WithContext(ctx).First(&run, "id = ?", req.SandboxRunID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("unable to find sandbox run: %w", err)
	}

	if run.PlannedAt == nil || run.InstallSandboxID == nil {
		return false, nil
	}

	var newerRun app.InstallSandboxRun
	res := a.db.WithContext(ctx).
		Where(app.InstallSandboxRun{InstallSandboxID: run.InstallSandboxID}).
		Where("applied_at > ? AND id != ?", *run.PlannedAt, run.ID).
		First(&newerRun)
	if res.Error == gorm.ErrRecordNotFound {
		return false, nil
	}
	if res.Error != nil {
		return false, fmt.Errorf("unable to check for newer sandbox runs: %w", res.Error)
	}
	return true, nil
}

type CheckDeploySupersededRequest struct {
	DeployID string `validate:"required"`
}

// CheckDeploySuperseded checks whether a newer deploy has been applied for the
// same component since this deploy's plan was created.
//
// Returns false (not superseded) if the deploy has no planned_at set (backward
// compatible with records created before this field existed).
//
// @temporal-gen-v2 activity
func (a *Activities) CheckDeploySuperseded(ctx context.Context, req CheckDeploySupersededRequest) (bool, error) {
	var deploy app.InstallDeploy
	if err := a.db.WithContext(ctx).First(&deploy, "id = ?", req.DeployID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("unable to find deploy: %w", err)
	}

	if deploy.PlannedAt == nil {
		return false, nil
	}

	var newerDeploy app.InstallDeploy
	res := a.db.WithContext(ctx).
		Where(app.InstallDeploy{ComponentID: deploy.ComponentID, InstallComponentID: deploy.InstallComponentID}).
		Where("applied_at > ? AND id != ?", *deploy.PlannedAt, deploy.ID).
		First(&newerDeploy)
	if res.Error == gorm.ErrRecordNotFound {
		return false, nil
	}
	if res.Error != nil {
		return false, fmt.Errorf("unable to check for newer deploys: %w", res.Error)
	}
	return true, nil
}
