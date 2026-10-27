package activities

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/driftresolution"
)

type AdmitResolveDriftApplyRequest struct {
	StepID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field StepID
func (a *Activities) AdmitResolveDriftApply(ctx context.Context, req AdmitResolveDriftApplyRequest) error {
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var step app.WorkflowStep
		if err := tx.Where(app.WorkflowStep{ID: req.StepID}).First(&step).Error; err != nil {
			return generics.TemporalGormError(err)
		}
		var wf app.Workflow
		if err := tx.Where(app.Workflow{ID: step.InstallWorkflowID}).First(&wf).Error; err != nil {
			return err
		}
		if driftresolution.ApplyTarget(&wf, &step) == nil {
			return nil
		}
		install, err := driftresolution.LockInstall(ctx, tx, wf.OwnerID)
		if err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Workflow{ID: wf.ID}).First(&wf).Error; err != nil {
			return err
		}
		target := driftresolution.ApplyTarget(&wf, &step)
		if target.ApplyStepID == step.ID {
			return nil
		}
		if target.ApplyStepID != "" {
			var previous app.WorkflowStep
			if err := tx.Where(app.WorkflowStep{ID: target.ApplyStepID}).First(&previous).Error; err != nil {
				return err
			}
			var activeSignals int64
			if err := tx.Model(&app.QueueSignal{}).
				Where(app.QueueSignal{OwnerID: previous.ID, OwnerType: "install_workflow_steps"}).
				Where("COALESCE(status->>'status', '') NOT IN ?", []app.Status{app.StatusSuccess, app.StatusCancelled, app.StatusError, app.StatusDiscarded}).
				Count(&activeSignals).Error; err != nil {
				return err
			}
			if previous.FinishedAt.IsZero() || previous.Status.Status == app.StatusSuccess || activeSignals > 0 {
				return driftresolution.TargetChanged(fmt.Sprintf("resolution apply is already admitted for step %s", target.ApplyStepID))
			}
			if previous.StepTargetID != "" {
				if target.Kind == "stack" {
					var version app.InstallStackVersion
					if err := tx.Where(app.InstallStackVersion{ID: previous.StepTargetID}).First(&version).Error; err != nil {
						return err
					}
					if version.Checksum != "" || version.TerraformChecksum != "" {
						return driftresolution.TargetChanged("the previous stack apply may still be running; resolve drift again after it settles")
					}
				}
				var jobs []app.RunnerJob
				if err := tx.Preload("Executions").Where(app.RunnerJob{OwnerID: previous.StepTargetID}).Find(&jobs).Error; err != nil {
					return err
				}
				for _, job := range jobs {
					if !job.Status.IsTerminal() {
						return driftresolution.TargetChanged("the previous runner job is still active")
					}
					for _, execution := range job.Executions {
						if execution.Status.IsRunning() {
							return driftresolution.TargetChanged("the previous runner execution is still active")
						}
					}
				}
			}
		}
		if err := driftresolution.CheckTarget(ctx, tx, install, target); err != nil {
			return err
		}
		switch target.Kind {
		case "component":
			var deploy app.InstallDeploy
			if err := tx.Where(app.InstallDeploy{InstallComponentID: target.ResourceID, InstallWorkflowID: &wf.ID}).
				Order("created_at DESC, id DESC").First(&deploy).Error; err != nil {
				return err
			}
			if deploy.Type == app.InstallDeployTypeApply && (deploy.ComponentBuildID != target.BuildID || deploy.PlannedInputsDigest != target.InputsDigest) {
				return driftresolution.TargetChanged("the fresh plan does not match this resolution's build and inputs; resolve drift again")
			}
		case "sandbox":
			var run app.InstallSandboxRun
			if err := tx.Where(app.InstallSandboxRun{InstallSandboxID: &target.ResourceID, InstallWorkflowID: &wf.ID}).
				Order("created_at DESC, id DESC").First(&run).Error; err != nil {
				return err
			}
			if run.PlannedInputsDigest != target.InputsDigest {
				return driftresolution.TargetChanged("the fresh plan does not match this resolution's inputs; resolve drift again")
			}
		}
		target.ApplyStepID = step.ID
		return tx.Model(&wf).Select("request").Updates(&wf).Error
	})
	var changed driftresolution.TargetChanged
	if errors.As(err, &changed) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "resolution-target-changed", err)
	}
	return err
}
