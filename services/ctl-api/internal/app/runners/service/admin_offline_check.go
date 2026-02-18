package service

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminOfflineCheckRequest struct{}

// @ID						AdminOfflineCheckRunner
// @Summary				check a runner for being offline
// @Description.markdown	offline_check_runner.md
// @Param					runner_id	path	string							true	"runner ID"
// @Param					req			body	AdminOfflineCheckRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{boolean}	true
// @Router					/v1/runners/{runner_id}/offline-check [POST]
func (s *service) AdminOfflineCheck(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminOfflineCheckRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	s.evClient.Send(ctx, runnerID, &signals.Signal{
		Type: signals.OperationOfflineCheck,
	})

	ctx.JSON(http.StatusCreated, true)
}
