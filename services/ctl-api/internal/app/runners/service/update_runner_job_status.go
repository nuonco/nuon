package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/credentialcleanup"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/jobcompleted"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

type UpdateRunnerJobRequest struct {
	Status app.RunnerJobStatus `json:"status"`
}

//		@ID						UpdateRunnerJob
//		@Summary				update a runner job
//		@Description.markdown	update_runner_job.md
//		@Param					req				body	UpdateRunnerJobRequest	true	"Input"
//		@Param					runner_job_id	path	string					true	"runner job ID"
//		@Tags					runners/runner
//		@Accept					json
//		@Produce				json
//		@Security				APIKey
//		@Security				OrgID
//	 @Deprecated     true
//		@Failure				400	{object}	stderr.ErrResponse
//		@Failure				401	{object}	stderr.ErrResponse
//		@Failure				403	{object}	stderr.ErrResponse
//		@Failure				404	{object}	stderr.ErrResponse
//		@Failure				500	{object}	stderr.ErrResponse
//		@Success				200	{object}	app.RunnerJob
//		@Router					/v1/runner-jobs/{runner_job_id} [PATCH]
func (s *service) UpdateRunnerJob(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")

	var req UpdateRunnerJobRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	job, err := s.updateRunnerJob(ctx, runnerJobID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update runner job: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, job)
}

func (s *service) updateRunnerJob(ctx context.Context, runnerJobID string, req UpdateRunnerJobRequest) (*app.RunnerJob, error) {
	job := app.RunnerJob{}
	res := s.db.WithContext(ctx).
		Model(app.RunnerJob{
			ID: runnerJobID,
		}).
		Updates(app.RunnerJob{
			Status: req.Status,
		})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update runner job status: %w", res.Error)
	}
	if req.Status.IsTerminal() {
		var existing app.RunnerJob
		if err := s.db.WithContext(ctx).Select("id", "runner_id").Where(app.RunnerJob{ID: runnerJobID}).Take(&existing).Error; err != nil {
			return nil, fmt.Errorf("load job for completion: %w", err)
		}
		// The status update above already succeeded; a failure to enqueue the
		// cleanup signal is logged, not returned, so it doesn't fail this
		// request for the runner that just reported completion.
		// CleanupInstallTokens sweeps up any job accounts this leaves behind.
		if err := s.enqueueJobCompleted(ctx, existing.RunnerID, existing.ID); err != nil {
			s.l.Error("unable to enqueue job-completed cleanup signal",
				zap.String("job_id", existing.ID), zap.String("runner_id", existing.RunnerID), zap.Error(err))
		}
	}

	return &job, nil
}

func (s *service) enqueueJobCompleted(ctx context.Context, runnerID, jobID string) error {
	if s.queueClient == nil {
		return fmt.Errorf("runner queue client is unavailable")
	}
	if runnerID == "" {
		return fmt.Errorf("runner id is required")
	}
	q, err := s.queueClient.GetQueueByOwnerAndName(ctx, runnerID, "runners", queuenames.RunnerSignalsQueueName)
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			return fmt.Errorf("runner-signals queue not found for runner %s", runnerID)
		}
		return fmt.Errorf("find runner-signals queue: %w", err)
	}
	if err := credentialcleanup.EnqueueOnQueue(ctx, s.queueClient, credentialcleanup.Queue{
		ID:        q.ID,
		OwnerID:   runnerID,
		OwnerType: plugins.TableName(s.db, app.Runner{}),
	}, credentialcleanup.Request{ServiceAccountSubject: jobID}); err != nil {
		return err
	}
	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:        q.ID,
		OwnerID:        runnerID,
		OwnerType:      plugins.TableName(s.db, app.Runner{}),
		Signal:         &jobcompleted.Signal{JobID: jobID, RunnerID: runnerID},
		IdempotencyKey: "job-completed:" + jobID,
	}); err != nil {
		return fmt.Errorf("enqueue job-completed: %w", err)
	}
	return nil
}

// @ID						UpdateRunnerJobV2
// @Summary				update a runner job
// @Description.markdown	update_runner_job.md
// @Param					req				body	UpdateRunnerJobRequest	true	"Input"
// @Param					runner_id	path	string				true	"runner ID"
// @Param					job_id	path	string					true	"job ID"
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.RunnerJob
// @Router					/v1/runners/{runner_id}/jobs/{job_id} [PATCH]
func (s *service) UpdateRunnerJobV2(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	_, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner: %w", err))
		return
	}

	runnerJobID := ctx.Param("job_id")

	var req UpdateRunnerJobRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	job, err := s.updateRunnerJob(ctx, runnerJobID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update runner job: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, job)
}
