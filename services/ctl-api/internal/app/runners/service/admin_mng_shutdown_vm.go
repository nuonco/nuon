package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// AdminMngVMShutDownRequest represents the request body for shutting down an install runner VM (admin).
type AdminMngVMShutDownRequest struct{}

// @ID						AdminMngVMShutDownRunner
// @Summary				shut down an install runner VM (admin)
// @Param					runner_id	path	string						true	"runner ID"
// @Param					req			body	AdminMngVMShutDownRequest	false	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/runners/{runner_id}/mng/shutdown-vm [POST]
func (s *service) AdminMngVMShutDown(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner %s: %w", runnerID, err))
		return
	}

	// Create a fresh log stream for this shutdown job
	ls := app.LogStream{
		OwnerType: "runner_operations",
		OwnerID:   runner.ID,
		Open:      true,
	}
	if res := s.db.WithContext(ctx).Create(&ls); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create log stream: %w", res.Error))
		return
	}

	job, err := s.helpers.CreateMngJob(ctx, runner.ID, ls.ID, app.RunnerJobTypeMngVMShutDown, map[string]string{
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
