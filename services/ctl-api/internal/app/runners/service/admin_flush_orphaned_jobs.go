package service

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminFlushOrphanedJobsRequest struct{}

// @ID						AdminFlushOrphanedJobRequest
// @Summary				flush orphaned jobs on a runner
// @Description.markdown	flush_orphaned_jobs.md
// @Param					runner_id	path	string						true	"runner ID"
// @Param					req			body	AdminForceShutdownRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{boolean}	true
// @Router					/v1/runners/{runner_id}/flush-orphaned-jobs [POST]
func (s *service) AdminFlushOrphanedJobs(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminFlushOrphanedJobsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	s.evClient.Send(ctx, runnerID, &signals.Signal{
		Type: signals.OperationFlushOrphanedJobs,
	})

	ctx.JSON(http.StatusCreated, true)
}
