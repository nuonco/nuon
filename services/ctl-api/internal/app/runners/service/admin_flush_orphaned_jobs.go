package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/flushorphanedjobs"
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

	if err := s.helpers.EnqueueRunnerSignal(ctx, runnerID, &flushorphanedjobs.Signal{RunnerID: runnerID}); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue flush-orphaned-jobs signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, true)
}
