package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SaveRunnerJobPlanRequest struct {
	JobID         string                  `validate:"required"`
	CompositePlan plantypes.CompositePlan `validate:"required"`
	// Deprecated: but kept for backward compatibility
	PlanJSON string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) SaveRunnerJobPlan(ctx context.Context, req *SaveRunnerJobPlanRequest) error {
	if err := a.runnersHelpers.WriteJobPlan(
		ctx,
		req.JobID,
		[]byte(req.PlanJSON),
		req.CompositePlan,
	); err != nil {
		return fmt.Errorf("unable to write job plan: %w", err)
	}

	digest := ""
	if req.CompositePlan.DeployPlan != nil {
		digest = req.CompositePlan.DeployPlan.InputsDigest
	} else if req.CompositePlan.SandboxRunPlan != nil {
		digest = req.CompositePlan.SandboxRunPlan.InputsDigest
	}
	if digest == "" {
		return nil
	}
	var job app.RunnerJob
	if err := a.db.WithContext(ctx).Select("id", "owner_id", "owner_type", "operation").
		Where(app.RunnerJob{ID: req.JobID}).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to get runner job for input evidence: %w", err)
	}
	var model any
	switch job.OwnerType {
	case "install_deploys":
		model = &app.InstallDeploy{}
	case "install_sandbox_runs":
		model = &app.InstallSandboxRun{}
	default:
		return nil
	}
	query := a.db.WithContext(ctx).Model(model).Where("id = ?", job.OwnerID)
	var err error
	switch job.Operation {
	case app.RunnerJobOperationTypeCreateApplyPlan, app.RunnerJobOperationTypeCreateTeardownPlan:
		updates := map[string]any{"planned_inputs_digest": digest}
		if job.OwnerType == "install_deploys" && job.Operation == app.RunnerJobOperationTypeCreateApplyPlan && req.CompositePlan.DeployPlan != nil {
			updates["dependency_build_ids"] = req.CompositePlan.DeployPlan.DependencyBuildIDs
		}
		err = query.Updates(updates).Error
	case app.RunnerJobOperationTypeApplyPlan:
		err = query.Update("applied_inputs_digest", gorm.Expr("CASE WHEN planned_inputs_digest = ? THEN ? ELSE '' END", digest, digest)).Error
	}
	if err != nil {
		return fmt.Errorf("unable to save runner job input evidence: %w", err)
	}
	return nil
}
