package activities

import (
	"context"
	"encoding/json"
	stderrors "errors"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/appliedstate"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type SyncNoopDeployOutputsRequest struct {
	StepID string `validate:"required"`
}

// SyncNoopDeployOutputs copies outputs from the latest plan job to a new
// finished apply job on the deploy. This is called when a deploy plan is noop
// (infrastructure matches desired state) so that the deploy has outputs just
// like a normal apply would produce. The plan job already collected outputs
// via terraform output / helm get values.
//
// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) SyncNoopDeployOutputs(ctx context.Context, req *SyncNoopDeployOutputsRequest) error {
	step, err := a.PkgWorkflowsFlowGetFlowsStep(ctx, GetFlowStepRequest{
		FlowStepID: req.StepID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get step")
	}

	return stderrors.Join(a.syncNoopDeployOutputs(ctx, step), a.recordNoopApplied(ctx, step))
}

func (a *Activities) recordNoopApplied(ctx context.Context, step *app.WorkflowStep) error {
	var flw app.Workflow
	if err := a.db.WithContext(ctx).Select("id", "plan_only").Where(app.Workflow{ID: step.InstallWorkflowID}).First(&flw).Error; err != nil {
		return errors.Wrap(err, "unable to get workflow")
	}
	if flw.PlanOnly {
		return nil
	}
	var applySteps int64
	if err := a.db.WithContext(ctx).Model(&app.WorkflowStep{}).
		Where(app.WorkflowStep{InstallWorkflowID: step.InstallWorkflowID}).
		Where("group_idx = ? AND idx > ? AND execution_type <> ?", step.GroupIdx, step.Idx, app.WorkflowStepExecutionTypeSkipped).
		Count(&applySteps).Error; err != nil {
		return errors.Wrap(err, "unable to get apply step")
	}
	if applySteps == 0 {
		return nil
	}

	switch step.StepTargetType {
	case string(app.WorkflowStepTargetTypeInstallDeploys):
		var deploy app.InstallDeploy
		if err := a.db.WithContext(ctx).Select("id", "type").Where(app.InstallDeploy{ID: step.StepTargetID}).First(&deploy).Error; err != nil {
			return errors.Wrap(err, "unable to get install deploy")
		}
		if deploy.Type == app.InstallDeployTypeTeardown || deploy.Type == app.InstallDeployTypeRecover {
			return nil
		}
		if err := a.db.WithContext(ctx).Model(&app.InstallDeploy{}).Where(app.InstallDeploy{ID: deploy.ID}).
			Update("applied_inputs_digest", gorm.Expr("planned_inputs_digest")).Error; err != nil {
			return errors.Wrap(err, "unable to record noop deploy inputs")
		}
		return appliedstate.RecordDeploy(ctx, a.db, deploy.ID)
	case string(app.WorkflowStepTargetTypeInstallSandboxRuns):
		var run app.InstallSandboxRun
		if err := a.db.WithContext(ctx).Select("id", "run_type").Where(app.InstallSandboxRun{ID: step.StepTargetID}).First(&run).Error; err != nil {
			return errors.Wrap(err, "unable to get install sandbox run")
		}
		if run.RunType != app.SandboxRunTypeProvision && run.RunType != app.SandboxRunTypeReprovision {
			return nil
		}
		if err := a.db.WithContext(ctx).Model(&app.InstallSandboxRun{}).Where(app.InstallSandboxRun{ID: run.ID}).
			Update("applied_inputs_digest", gorm.Expr("planned_inputs_digest")).Error; err != nil {
			return errors.Wrap(err, "unable to record noop sandbox run inputs")
		}
		return appliedstate.RecordSandboxRun(ctx, a.db, run.ID)
	}
	return nil
}

func (a *Activities) syncNoopDeployOutputs(ctx context.Context, step *app.WorkflowStep) error {
	if step.StepTargetType != string(app.WorkflowStepTargetTypeInstallDeploys) {
		return nil
	}

	// Get the latest plan job (uses the view, so outputs are available).
	var planJob app.RunnerJob
	err := a.db.WithContext(ctx).
		Where("owner_id = ? AND operation = ?", step.StepTargetID, app.RunnerJobOperationTypeCreateApplyPlan).
		Order("created_at DESC").
		First(&planJob).Error
	if err != nil {
		return nil
	}

	if len(planJob.ParsedOutputs) == 0 {
		return nil
	}

	outputsJSON, err := json.Marshal(planJob.ParsedOutputs)
	if err != nil {
		return nil
	}

	// the blob upload in RunnerJobExecutionOutputs' BeforeCreate hook requires org_id on the context
	ctx = cctx.SetOrgIDContext(ctx, planJob.OrgID)

	// Create a finished apply job with the plan's outputs so the deploy
	// has outputs in the same structure the rest of the system expects.
	applyJob := &app.RunnerJob{
		CreatedByID:       planJob.CreatedByID,
		RunnerID:          planJob.RunnerID,
		OrgID:             planJob.OrgID,
		OwnerID:           planJob.OwnerID,
		OwnerType:         planJob.OwnerType,
		LogStreamID:       planJob.LogStreamID,
		Status:            app.RunnerJobStatusFinished,
		StatusDescription: "noop apply - outputs synced from plan",
		Type:              planJob.Type,
		Group:             planJob.Group,
		Operation:         app.RunnerJobOperationTypeApplyPlan,
		QueueTimeout:      planJob.QueueTimeout,
		AvailableTimeout:  planJob.AvailableTimeout,
		ExecutionTimeout:  planJob.ExecutionTimeout,
		MaxExecutions:     1,
	}

	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(applyJob).Error; err != nil {
			return errors.Wrap(err, "unable to create noop apply job")
		}

		execution := &app.RunnerJobExecution{
			RunnerJobID: applyJob.ID,
			Status:      app.RunnerJobExecutionStatusFinished,
		}
		if err := tx.Create(execution).Error; err != nil {
			return errors.Wrap(err, "unable to create noop apply execution")
		}

		executionOutputs := &app.RunnerJobExecutionOutputs{
			RunnerJobExecutionID: execution.ID,
			OrgID:                planJob.OrgID,
			Outputs:              outputsJSON,
			OutputsBlob:          &blobstore.Blob{},
		}
		executionOutputs.OutputsBlob.Set(string(outputsJSON))
		if err := tx.Create(executionOutputs).Error; err != nil {
			return errors.Wrap(err, "unable to create noop apply outputs")
		}

		return nil
	})
}
