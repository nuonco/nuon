package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type MngVMShutDownRequest struct{}

// @ID						MngVMShutDown
// @Summary				shut down an install runner VM
// @Param					runner_id	path	string						true	"runner ID"
// @Param					req			body	MngVMShutDownRequest	true	"Input"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/runners/{runner_id}/mng/shutdown-vm [POST]
// @Deprecated
func (s *service) MngVMShutDown(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	runnerID := ctx.Param("runner_id")
	runner, err := s.getOrgRunner(ctx, runnerID, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner %s: %w", runnerID, err))
		return
	}

	var req MngVMShutDownRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	// Reuse the log stream from the most recent mng process if one exists
	var logStreamID string
	var mngProcess app.RunnerProcess
	if res := s.db.WithContext(ctx).
		Where(app.RunnerProcess{RunnerID: runner.ID, Type: app.RunnerProcessTypeMng}).
		Order("created_at DESC").
		First(&mngProcess); res.Error == nil && mngProcess.LogStreamID != nil {
		logStreamID = *mngProcess.LogStreamID
	}

	job, err := s.helpers.CreateMngJob(ctx, runner.ID, logStreamID, app.RunnerJobTypeMngVMShutDown, map[string]string{
		"shutdown_type": "vm",
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create mng vm shutdown job: %w", err))
		return
	}

	job.Status = app.RunnerJobStatusAvailable
	job.StatusDescription = string(app.RunnerJobStatusAvailable)
	if res := s.db.WithContext(ctx).Save(job); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to update job status: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
