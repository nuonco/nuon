package demand

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalKey = "runner_demand_signal_id"
const StepKey = "runner_demand_step_id"
const AwaitRunner = "await-runner"

var ErrJobChanged = errors.New("runner job workflow identity changed")

func SignalID(job *app.RunnerJob) string { return generics.HstoreValue(job.Metadata, SignalKey) }
func StepID(job *app.RunnerJob) string   { return generics.HstoreValue(job.Metadata, StepKey) }

func LockWorkflow(ctx context.Context, tx *gorm.DB, workflowID string) (*app.Workflow, error) {
	var flw app.Workflow
	if err := tx.WithContext(ctx).Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Workflow{ID: workflowID}).First(&flw).Error; err != nil {
		return nil, err
	}
	return &flw, nil
}

func RequireActiveWorkflow(flw *app.Workflow) error {
	if flw == nil {
		return nil
	}
	if flw.Status.Status == app.StatusCancelled || flw.Status.Metadata["cancel_requested_at"] != nil {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("workflow %s is cancelled", flw.ID), "CANCELLED", nil)
	}
	return nil
}

func LockJob(ctx context.Context, tx *gorm.DB, job *app.RunnerJob) (*app.Workflow, error) {
	var flw *app.Workflow
	if SignalID(job) != "" {
		var err error
		flw, err = LockWorkflow(ctx, tx, job.FlowWorkflowID())
		if err != nil {
			return nil, err
		}
	}
	var locked app.RunnerJob
	if err := tx.WithContext(ctx).Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.RunnerJob{ID: job.ID}).First(&locked).Error; err != nil {
		return nil, err
	}
	if SignalID(&locked) != SignalID(job) || locked.FlowWorkflowID() != job.FlowWorkflowID() || StepID(&locked) != StepID(job) {
		return nil, ErrJobChanged
	}
	*job = locked
	return flw, nil
}

func CancelJob(ctx context.Context, tx *gorm.DB, job *app.RunnerJob, compositeError *compositeerrors.CompositeErrorData) (bool, error) {
	if job.Status.IsTerminal() {
		return false, nil
	}
	job.Status, job.FinishedAt, job.CompositeError = app.RunnerJobStatusCancelled, time.Now(), compositeError
	job.StatusV2 = app.NewCompositeStatus(ctx, app.StatusCancelled)
	if err := tx.WithContext(ctx).Scopes(scopes.WithDisableViews).Model(job).Updates(map[string]any{
		"status": job.Status, "status_v2": job.StatusV2, "finished_at": job.FinishedAt, "composite_error": compositeError,
	}).Error; err != nil {
		return false, err
	}
	err := tx.WithContext(ctx).Model(&app.RunnerJobExecution{}).Where(app.RunnerJobExecution{RunnerJobID: job.ID}).
		Where("status IN ?", []app.RunnerJobExecutionStatus{app.RunnerJobExecutionStatusPending, app.RunnerJobExecutionStatusInitializing, app.RunnerJobExecutionStatusInProgress, app.RunnerJobExecutionStatusCleaningUp}).
		Updates(map[string]any{"status": app.RunnerJobExecutionStatusCancelled, "status_v2": job.StatusV2}).Error
	return err == nil, err
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
