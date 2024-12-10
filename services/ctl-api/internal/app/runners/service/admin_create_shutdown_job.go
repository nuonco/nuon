package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

type AdminCreateShutDownJobRequest struct{}

// @ID AdminShutDownRunner
// @Summary	shut down a runner
// @Description.markdown shut_down_runner.md
// @Param			runner_id	path	string						true	"runner ID"
// @Param			req				body	AdminCreateShutDownJobRequest	true	"Input"
// @Tags runners/admin
// @Security AdminEmail
// @Accept			json
// @Produce		json
// @Success		201	{boolean}	true
// @Router			/v1/runners/{runner_id}/shutdown-job [POST]
func (s *service) AdminCreateShutDown(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminCreateShutDownJobRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	job, err := s.adminCreateJob(ctx, runnerID, app.RunnerJobTypeShutDown)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create shutdown job: %w", err))
		return
	}

	s.evClient.Send(ctx, runnerID, &signals.Signal{
		Type:  signals.OperationProcessJob,
		JobID: job.ID,
	})

	ctx.JSON(http.StatusCreated, true)
}
