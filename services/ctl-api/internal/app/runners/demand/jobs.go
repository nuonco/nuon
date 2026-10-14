package demand

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalKey = "runner_demand_signal_id"
const StepKey = "runner_demand_step_id"
const AwaitRunner = "await-runner"

func SignalID(job *app.RunnerJob) string { return generics.HstoreValue(job.Metadata, SignalKey) }
func StepID(job *app.RunnerJob) string   { return generics.HstoreValue(job.Metadata, StepKey) }

func LockWorkflow(ctx context.Context, tx *gorm.DB, workflowID string) (*app.Workflow, error) {
	var flw app.Workflow
	if err := tx.WithContext(ctx).Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Workflow{ID: workflowID}).First(&flw).Error; err != nil {
		return nil, err
	}
	if flw.Status.Status == app.StatusCancelled || flw.Status.Metadata["cancel_requested_at"] != nil {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("workflow %s is cancelled", workflowID), "CANCELLED", nil)
	}
	return &flw, nil
}

func UpdateOwner(ctx context.Context, tx *gorm.DB, job *app.RunnerJob, status app.Status) error {
	var owner any
	ownerStatus := string(status)
	switch job.OwnerType {
	case "install_deploys":
		owner = &app.InstallDeploy{ID: job.OwnerID}
		if status == app.StatusInProgress {
			ownerStatus = string(app.InstallDeployStatusExecuting)
		}
	case "install_sandbox_runs":
		owner = &app.InstallSandboxRun{ID: job.OwnerID}
		if status == app.StatusInProgress {
			ownerStatus = string(app.SandboxRunStatusApplying)
		}
	case "install_action_workflow_runs":
		owner = &app.InstallActionWorkflowRun{ID: job.OwnerID}
		if status == app.StatusInProgress {
			ownerStatus = string(app.InstallActionRunStatusInProgress)
		}
	default:
		return nil
	}
	if status == app.StatusCancelled {
		_, err := statusactivities.CancelIfUnfinished(ctx, tx, owner, string(status), true)
		return err
	}
	return tx.WithContext(ctx).Scopes(scopes.WithDisableViews).Model(owner).
		Where("status <> ?", "cancelled").Updates(map[string]any{
		"status": ownerStatus, "status_description": string(status), "status_v2": app.NewCompositeStatus(ctx, status),
	}).Error
}

func Pickup(ctx context.Context, tx *gorm.DB, job *app.RunnerJob, now time.Time) error {
	if StepID(job) == "" {
		return nil
	}
	status := app.NewCompositeStatus(ctx, app.StatusInProgress)
	res := tx.WithContext(ctx).Model(&app.WorkflowStep{}).
		Where(app.WorkflowStep{ID: StepID(job), RunnerJobID: &job.ID}).
		Where("status->>'status' <> ?", app.StatusCancelled).
		Updates(map[string]any{"status": status, "started_at": now, "finished_at": nil})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("runner job %s no longer belongs to an executable step", job.ID)
	}
	var step app.WorkflowStep
	if err := tx.Where(app.WorkflowStep{ID: StepID(job)}).First(&step).Error; err != nil {
		return err
	}
	if step.WorkflowStepGroupID != "" {
		if err := tx.Model(&app.WorkflowStepGroup{}).Where(app.WorkflowStepGroup{ID: step.WorkflowStepGroupID}).
			Where("status->>'status' <> ?", app.StatusCancelled).Update("status", status).Error; err != nil {
			return err
		}
	}
	var flw app.Workflow
	if err := tx.Scopes(scopes.WithDisableViews).Where(app.Workflow{ID: step.InstallWorkflowID}).First(&flw).Error; err != nil {
		return err
	}
	flw.Status.Status = app.StatusInProgress
	flw.Status.StatusHumanDescription = "runner picked up job"
	if err := tx.Scopes(scopes.WithDisableViews).Model(&flw).Update("status", flw.Status).Error; err != nil {
		return err
	}
	return UpdateOwner(ctx, tx, job, app.StatusInProgress)
}
