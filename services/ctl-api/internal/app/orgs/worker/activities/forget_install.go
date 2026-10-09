package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
)

type ForgetInstallRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) ForgetInstall(ctx context.Context, req ForgetInstallRequest) error {
	installOwnerType := plugins.TableName(a.db, app.Install{})

	// Cancel open install workflows before soft-deleting queue signals so
	// flowsClient can still find the execute-flow signal. Otherwise forgotten
	// installs leave hanging Temporal WFs that keep emitting lifecycle noise.
	if err := a.cancelOpenInstallWorkflows(ctx, req.InstallID, installOwnerType); err != nil {
		return err
	}

	// must run before the cascade delete below soft-deletes the queues
	var queueIDs []string
	if res := a.db.WithContext(ctx).
		Model(&app.Queue{}).
		Where(app.Queue{
			OwnerID:   req.InstallID,
			OwnerType: installOwnerType,
		}).
		Pluck("id", &queueIDs); res.Error != nil {
		return dbgenerics.TemporalGormError(res.Error, "unable to list install queues: %w")
	}

	var runnerGroupIDs []string
	if res := a.db.WithContext(ctx).
		Model(&app.RunnerGroup{}).
		Where(app.RunnerGroup{
			OwnerID:   req.InstallID,
			OwnerType: installOwnerType,
		}).
		Pluck("id", &runnerGroupIDs); res.Error != nil {
		return dbgenerics.TemporalGormError(res.Error, "unable to list install runner groups: %w")
	}

	var runnerIDs []string
	if len(runnerGroupIDs) > 0 {
		if res := a.db.WithContext(ctx).
			Model(&app.Runner{}).
			Where("runner_group_id IN ?", runnerGroupIDs).
			Pluck("id", &runnerIDs); res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to list install runners: %w")
		}
	}

	var runnerQueueIDs []string
	if len(runnerIDs) > 0 {
		if res := a.db.WithContext(ctx).
			Model(&app.Queue{}).
			Where(app.Queue{OwnerType: plugins.TableName(a.db, app.Runner{})}).
			Where("owner_id IN ?", runnerIDs).
			Pluck("id", &runnerQueueIDs); res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to list runner queues: %w")
		}
	}

	allQueueIDs := append(queueIDs, runnerQueueIDs...)
	if len(allQueueIDs) > 0 {
		if res := a.db.WithContext(ctx).
			Where("queue_id IN ?", allQueueIDs).
			Delete(&app.QueueEmitter{}); res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to delete queue emitters: %w")
		}

		if res := a.db.WithContext(ctx).
			Where("queue_id IN ?", allQueueIDs).
			Delete(&app.QueueSignal{}); res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to delete queue signals: %w")
		}
	}

	if len(runnerQueueIDs) > 0 {
		if res := a.db.WithContext(ctx).
			Where("id IN ?", runnerQueueIDs).
			Delete(&app.Queue{}); res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to delete runner queues: %w")
		}
	}

	if len(runnerIDs) > 0 {
		if res := a.db.WithContext(ctx).
			Where("id IN ?", runnerIDs).
			Delete(&app.Runner{}); res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to delete runners: %w")
		}
	}

	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := account.New(account.Params{DB: tx}).DeleteInstallServiceAccounts(ctx, req.InstallID); err != nil {
			return err
		}

		res := tx.Select(clause.Associations).Delete(&app.Install{ID: req.InstallID})
		if res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to delete install: %w")
		}
		return nil
	})
}

func (a *Activities) cancelOpenInstallWorkflows(ctx context.Context, installID, installOwnerType string) error {
	var workflows []app.Workflow
	if err := a.db.WithContext(ctx).
		Where(app.Workflow{
			OwnerID:   installID,
			OwnerType: installOwnerType,
		}).
		Find(&workflows).Error; err != nil {
		return dbgenerics.TemporalGormError(err, "unable to list install workflows: %w")
	}

	for i := range workflows {
		wf := &workflows[i]
		if !isCancelableWorkflowStatus(wf.Status.Status) {
			continue
		}
		if err := a.cancelInstallWorkflow(ctx, wf); err != nil {
			return err
		}
	}
	return nil
}

func isCancelableWorkflowStatus(status app.Status) bool {
	switch status {
	case app.StatusInProgress,
		app.StatusPending,
		app.AwaitingApproval,
		app.StatusFailedPendingRetry:
		return true
	default:
		return false
	}
}

func (a *Activities) cancelInstallWorkflow(ctx context.Context, wf *app.Workflow) error {
	if wf.Status.Status == app.StatusPending {
		return a.cancelWorkflowInDB(ctx, wf)
	}

	if _, err := a.flowsClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{
		InstallWorkflowID: wf.ID,
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return a.cancelWorkflowInDB(ctx, wf)
		}
		return fmt.Errorf("unable to cancel workflow %s: %w", wf.ID, err)
	}
	return a.markForgottenCancelReason(ctx, wf.ID)
}

const forgottenInstallCancelReason = "cancelled because install was forgotten"

func forgottenInstallCancelStatus(ctx context.Context) app.CompositeStatus {
	status := app.NewCompositeStatus(ctx, app.StatusCancelled)
	status.StatusHumanDescription = forgottenInstallCancelReason
	return status
}

func (a *Activities) markForgottenCancelReason(ctx context.Context, workflowID string) error {
	var wf app.Workflow
	if err := a.db.WithContext(ctx).First(&wf, "id = ?", workflowID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to load workflow %s after cancel: %w", workflowID, err)
	}
	wf.Status.StatusHumanDescription = forgottenInstallCancelReason
	if err := a.db.WithContext(ctx).Save(&wf).Error; err != nil {
		return fmt.Errorf("unable to set forget cancel reason on workflow %s: %w", workflowID, err)
	}
	return nil
}

func (a *Activities) cancelWorkflowInDB(ctx context.Context, wf *app.Workflow) error {
	wf.Status = forgottenInstallCancelStatus(ctx)
	wf.FinishedAt = time.Now()
	if err := a.db.WithContext(ctx).Save(wf).Error; err != nil {
		return fmt.Errorf("unable to cancel workflow %s in db: %w", wf.ID, err)
	}
	return nil
}
