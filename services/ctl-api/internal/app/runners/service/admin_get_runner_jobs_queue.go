package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@ID						AdminGetRunnerJobsQueue
//	@Summary				get runner jobs queue
//	@Description.markdown	admin_get_runner_jobs_queue.md
//	@Tags					runners/admin
//	@Param					runner_id	path	string	true	"runner ID jobs to fetch"
//	@Accept					json
//	@Produce				json
//	@Failure				400	{object}	stderr.ErrResponse
//	@Failure				401	{object}	stderr.ErrResponse
//	@Failure				403	{object}	stderr.ErrResponse
//	@Failure				404	{object}	stderr.ErrResponse
//	@Failure				500	{object}	stderr.ErrResponse
//	@Success				200	{array}		app.RunnerJob
//	@Router					/v1/runners/{runner_id}/jobs/queue [get]
func (s *service) AdminGetRunnerJobsQueue(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	runnerJobs, err := s.getRunnerJobsQueue(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, runnerJobs)
}

func (s *service) getRunnerJobsQueue(ctx context.Context, runnerID string) ([]*app.RunnerJob, error) {
	runnerJobs := []*app.RunnerJob{}

	var jobs []*app.RunnerJob
	res := s.db.WithContext(ctx).
		Where("runner_id = ? AND status IN ?", runnerID, []app.RunnerJobStatus{
			app.RunnerJobStatusQueued,
			app.RunnerJobStatusAvailable,
			app.RunnerJobStatusInProgress,
		}).
		Order("created_at desc").Find(&jobs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner job queue: %w", res.Error)
	}

	return runnerJobs, nil
}
