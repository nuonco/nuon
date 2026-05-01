package service

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminMngVMShutDownRequest struct{}

// @ID						AdminMngVMShutDownRunner
// @Summary				shut down an install runner VM (admin)
// @Param					runner_id	path	string						true	"runner ID"
// @Param					req			body	AdminMngVMShutDownRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/runners/{runner_id}/mng/shutdown-vm [POST]
func (s *service) AdminMngVMShutDown(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminMngVMShutDownRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	s.evClient.Send(ctx, runnerID, &signals.Signal{
		Type: signals.OperationMngVMShutDown,
	})

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
