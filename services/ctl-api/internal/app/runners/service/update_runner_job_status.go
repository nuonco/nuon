package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
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
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
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

	return &job, nil
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
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	job, err := s.updateRunnerJob(ctx, runnerJobID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update runner job: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, job)
}
