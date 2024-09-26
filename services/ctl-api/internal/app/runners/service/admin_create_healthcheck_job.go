package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

type AdminCreateHealthCheckJobRequest struct{}

// @ID AdminCreateRunnerHealthCheck
// @Summary	create a health check
// @Description.markdown runner_health_check.md
// @Param			runner_id	path	string						true	"runner ID"
// @Param			req	body	AdminCreateHealthCheckJobRequest	true "Input"
// @Tags runners/admin
// @Accept			json
// @Produce		json
// @Success		201	{boolean}	true
// @Router			/v1/runners/{runner_id}/health-check-job [POST]
func (s *service) AdminCreateHealthCheck(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminCreateHealthCheckJobRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	job, err := s.adminCreateJob(ctx, runnerID, app.RunnerJobTypeHealthCheck)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create health check job: %w", err))
		return
	}

	s.evClient.Send(ctx, runnerID, &signals.Signal{
		Type:  signals.OperationJobQueued,
		JobID: job.ID,
	})

	ctx.JSON(http.StatusCreated, true)
}
