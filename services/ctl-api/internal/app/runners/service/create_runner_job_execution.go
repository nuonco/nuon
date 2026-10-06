package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/joberrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processjob"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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

	if runnerJob.ExecutionCount >= runnerJob.MaxExecutions {
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

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &runnerJobExecution, nil
}
