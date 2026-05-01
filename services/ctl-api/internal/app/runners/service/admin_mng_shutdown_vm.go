package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
)

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

	s.evClient.Send(ctx, runner.ID, &signals.Signal{
		Type: signals.OperationMngVMShutDown,
	})

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
