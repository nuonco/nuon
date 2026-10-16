package activities

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/demand"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (a *Activities) PauseForRunner(ctx context.Context, workflowID, groupID string) (err error) {
	l := temporalzap.GetActivityLogger(ctx).With(zap.String("workflow_id", workflowID), zap.String("group_id", groupID))
	l.Info("pausing workflow for runner")
	defer func() {
		if err != nil {
			l.Error("unable to pause workflow", zap.Error(err))
		} else {
			l.Info("workflow paused for runner")
		}
	}()
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		flw, err := demand.LockWorkflow(ctx, tx, workflowID)
		if err != nil {
			return err
		}
		if err := demand.RequireActiveWorkflow(flw); err != nil {
			return err
		}
		var steps []app.WorkflowStep
		query := tx.Where(app.WorkflowStep{InstallWorkflowID: workflowID})
		if groupID != "" {
			query = query.Where(app.WorkflowStep{WorkflowStepGroupID: groupID})
		}
		if err := query.Find(&steps).Error; err != nil {
			return err
		}
		status := app.StatusWaitingForRunner
		for _, st := range steps {
			if st.ResultDirective == demand.AwaitRunner && st.Status.Status == app.StatusInProgress {
				status = app.StatusInProgress
				break
			}
		}
		if groupID != "" {
			return tx.Model(&app.WorkflowStepGroup{}).Where(app.WorkflowStepGroup{ID: groupID}).Update("status", app.NewCompositeStatus(ctx, status)).Error
		}
		flw.Status.Status = status
		flw.Status.StatusHumanDescription = "waiting for runner jobs"
		return tx.Scopes(scopes.WithDisableViews).Model(flw).Update("status", flw.Status).Error
	})
}

func (a *Activities) PrepareRunnerJobResume(ctx context.Context, workflowID, jobID string) (string, error) {
	l := temporalzap.GetActivityLogger(ctx).With(zap.String("workflow_id", workflowID), zap.String("job_id", jobID))
	l.Info("preparing workflow resume")
	var stepID string
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		flw, err := demand.LockWorkflow(ctx, tx, workflowID)
		if err != nil {
			return err
		}
		if err := demand.RequireActiveWorkflow(flw); err != nil {
			return err
		}
		var job app.RunnerJob
		if err := tx.Scopes(scopes.WithDisableViews).Where(app.RunnerJob{ID: jobID}).First(&job).Error; err != nil {
			return err
		}
		if job.FlowWorkflowID() != workflowID {
			return fmt.Errorf("job does not belong to workflow")
		}
		if !job.Status.IsTerminal() {
			return fmt.Errorf("job is not terminal")
		}
		var step app.WorkflowStep
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.WorkflowStep{ID: demand.StepID(&job)}).First(&step).Error; err != nil {
			return err
		}
		if step.RunnerJobID == nil || *step.RunnerJobID != job.ID || step.Status.Status == app.StatusCancelled {
			return nil
		}
		if step.Status.Status == app.StatusPending && step.ResultDirective == "" {
			stepID = step.ID
			return nil
		}
		if step.ResultDirective != demand.AwaitRunner {
			return nil
		}
		if err := tx.Model(&step).Updates(map[string]any{"status": app.NewCompositeStatus(ctx, app.StatusPending), "result_directive": ""}).Error; err != nil {
			return err
		}
		stepID = step.ID
		return nil
	})
	if err != nil {
		l.Error("unable to prepare workflow resume", zap.Error(err))
	} else {
		l.Info("workflow resume prepared", zap.String("step_id", stepID))
	}
	return stepID, err
}

func (a *Activities) CancelRunnerDemandWorkflow(ctx context.Context, workflowID string) (err error) {
	l := temporalzap.GetActivityLogger(ctx).With(zap.String("workflow_id", workflowID))
	l.Info("cancelling workflow runner jobs")
	defer func() {
		if err != nil {
			l.Error("unable to cancel workflow runner jobs", zap.Error(err))
		} else {
			l.Info("workflow runner jobs cancelled")
		}
	}()
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		flw, err := demand.LockWorkflow(ctx, tx, workflowID)
		if err != nil {
			return err
		}
		flw.Status.Status = app.StatusCancelled
		if flw.Status.Metadata == nil {
			flw.Status.Metadata = map[string]any{}
		}
		flw.Status.Metadata["cancel_requested_at"] = time.Now().Unix()
		if err := tx.Scopes(scopes.WithDisableViews).Model(flw).Update("status", flw.Status).Error; err != nil {
			return err
		}
		var jobs []app.RunnerJob
		if err := tx.Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "UPDATE"}).Where("metadata -> 'flow_workflow_id' = ? AND defined(metadata, ?)", workflowID, demand.SignalKey).Find(&jobs).Error; err != nil {
			return err
		}
		for _, job := range jobs {
			if _, err := demand.CancelJob(ctx, tx, &job, job.CompositeError); err != nil {
				return err
			}
			var qs app.QueueSignal
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.QueueSignal{ID: demand.SignalID(&job)}).First(&qs).Error; err != nil {
				return err
			}
			if qs.Status.Status == app.StatusSuccess || qs.Status.Status == app.StatusError {
				continue
			}
			if err := tx.Model(&app.QueueSignal{}).Where(app.QueueSignal{ID: demand.SignalID(&job)}).Update("status", app.NewCompositeStatus(ctx, app.StatusCancelled)).Error; err != nil {
				return err
			}
			if err := demand.UpdateOwner(ctx, tx, &job, app.StatusCancelled); err != nil {
				return err
			}
		}
		return nil
	})
}
