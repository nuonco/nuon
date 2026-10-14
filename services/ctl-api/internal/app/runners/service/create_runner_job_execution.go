package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/demand"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/joberrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processjob"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type CreateRunnerJobExecutionRequest struct{}

// @ID						CreateRunnerJobExecution
// @Summary				create runner job execution
// @Description.markdown	create_runner_job_execution.md
// @Param					req				body	CreateRunnerJobExecutionRequest	true	"Input"
// @Param					runner_job_id	path	string							true	"runner job ID"
// @Param					X-Nuon-Client-Version	header	string		false	"Nuon Client Version"
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.RunnerJobExecution
// @Router					/v1/runner-jobs/{runner_job_id}/executions [POST]
func (s *service) CreateRunnerJobExecution(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")
	clientVersion := ctx.GetHeader("X-Nuon-Client-Version")

	runnerJob, err := s.getRunnerJob(ctx, runnerJobID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get runner job"))
		return
	}
	cctx.SetOrgIDGinContext(ctx, runnerJob.OrgID)

	if demand.SignalID(runnerJob) == "" && runnerJob.ExecutionCount >= runnerJob.MaxExecutions {
		if _, err := s.cancelRunnerJob(ctx, runnerJobID, joberrors.CancellationReasonAttemptsExhausted); err != nil {
			ctx.Error(errors.Wrap(err, "unable to cancel runner job"))
			return
		}

		ctx.Error(fmt.Errorf("runner job has exceeded max executions"))
		return
	}

	execution, err := s.claimRunnerJob(ctx, runnerJobID, clientVersion)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Wake the process_job workflow so its pickup poll detects the new
	// execution immediately instead of on its next tick.
	s.wakeProcessJobWorkflow(ctx, runnerJobID, processjob.PickupSignalName(runnerJobID))

	ctx.JSON(http.StatusCreated, execution)
}

// claimRunnerJob moves the job from available to in-progress and creates its execution in one transaction, so two
// runner processes polling the same job can't both execute it.
func (s *service) claimRunnerJob(ctx context.Context, runnerJobID, clientVersion string) (*app.RunnerJobExecution, error) {
	runnerJobExecution := app.RunnerJobExecution{
		RunnerJobID: runnerJobID,
		Status:      app.RunnerJobExecutionStatusPending,
	}
	if clientVersion != "" {
		runnerJobExecution.Metadata = pgtype.Hstore(map[string]*string{
			"client.version": &clientVersion,
		})
	}

	var processSignalID string
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var job app.RunnerJob
		if err := tx.Scopes(scopes.WithDisableViews).Where(app.RunnerJob{ID: runnerJobID}).First(&job).Error; err != nil {
			return err
		}
		isDemand := demand.SignalID(&job) != ""
		if isDemand {
			if _, err := demand.LockWorkflow(ctx, tx, job.FlowWorkflowID()); err != nil {
				return stderr.ErrConflict{Err: err, Description: "workflow is cancelled"}
			}
			if err := tx.Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.RunnerJob{ID: runnerJobID}).First(&job).Error; err != nil {
				return err
			}
			var count int64
			if err := tx.Model(&app.RunnerJobExecution{}).Where(app.RunnerJobExecution{RunnerJobID: runnerJobID}).Count(&count).Error; err != nil {
				return err
			}
			if count >= int64(job.MaxExecutions) {
				return stderr.ErrConflict{Err: fmt.Errorf("runner job has exceeded max executions"), Description: "runner job has exceeded max executions"}
			}
			var qs app.QueueSignal
			if err := tx.Where(app.QueueSignal{ID: demand.SignalID(&job)}).First(&qs).Error; err != nil {
				return err
			}
			if qs.Status.Status == app.StatusCancelled {
				return stderr.ErrConflict{Err: fmt.Errorf("signal is cancelled"), Description: "signal is cancelled"}
			}
		}
		res := tx.Model(&app.RunnerJob{}).
			Where(app.RunnerJob{ID: runnerJobID, Status: app.RunnerJobStatusAvailable}).
			Updates(app.RunnerJob{
				Status:            app.RunnerJobStatusInProgress,
				StatusDescription: "in-progress",
			})
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to claim runner job")
		}
		if res.RowsAffected == 0 {
			return stderr.ErrConflict{
				Err:         fmt.Errorf("runner job %s is not available", runnerJobID),
				Description: "runner job was already claimed",
			}
		}

		if err := tx.Create(&runnerJobExecution).Error; err != nil {
			return errors.Wrap(err, "unable to create runner job execution")
		}
		if isDemand {
			now := runnerJobExecution.CreatedAt
			if now.IsZero() {
				now = time.Now()
			}
			if err := tx.Scopes(scopes.WithDisableViews).Model(&job).Updates(map[string]any{"started_at": now, "status_v2": app.NewCompositeStatus(ctx, app.StatusInProgress)}).Error; err != nil {
				return err
			}
			if err := demand.Pickup(ctx, tx, &job, now); err != nil {
				return stderr.ErrConflict{Err: err, Description: "step is no longer executable"}
			}
			var runner app.Runner
			if err := tx.Preload("Queues").Where(app.Runner{ID: job.RunnerID}).First(&runner).Error; err != nil {
				return err
			}
			q := runner.GetQueueForGroup(job.Group)
			if q == nil {
				return fmt.Errorf("runner job queue is missing")
			}
			resp, err := s.queueClient.EnqueueSignalInTransaction(ctx, tx, &queueclient.EnqueueSignalRequest{
				QueueID: q.ID, OwnerID: job.ID, OwnerType: "runner_jobs", IdempotencyKey: "runner-demand-claimed:" + runnerJobExecution.ID,
				Signal: &processjob.Signal{RunnerID: job.RunnerID, JobID: job.ID, ExecutionID: runnerJobExecution.ID, Demand: true},
			})
			if err != nil {
				return err
			}
			processSignalID = resp.ID
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.queueClient.DispatchSignal(processSignalID)
	return &runnerJobExecution, nil
}
