package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminCancelRunnerJobRequest struct{}

// @ID						AdminCancelRunnerJob
// @Summary				cancel a runner job
// @Description.markdown	admin_cancel_runner_job.md
// @Param					runner_job_id	path	string					true	"runner ID"
// @Param					req				body	CancelRunnerJobRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{boolean}	true
// @Router					/v1/runner-jobs/{runner_job_id}/cancel [POST]
func (s *service) AdminCancelRunnerJob(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")

	var req AdminCancelRunnerJobRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if _, err := s.cancelRunnerJob(ctx, runnerJobID); err != nil {
		ctx.Error(fmt.Errorf("unable to cancel job: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}
