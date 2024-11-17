package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID AdminGetRunnerJob
// @Summary	get a runner job
// @Description.markdown admin_get_runner_job.md
// @Param			runner_job_id	path	string					true	"runner ID"
// @Tags runners/admin
// @Security AdminEmail
// @Accept			json
// @Produce		json
// @Success		200				{object}	app.RunnerJob
// @Router			/v1/runner-jobs/{runner_job_id} [GET]
func (s *service) AdminGetRunnerJob(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")

	runnerJob, err := s.getRunnerJob(ctx, runnerJobID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runnerJob)
}
