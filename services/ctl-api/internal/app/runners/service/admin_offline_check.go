package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
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
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	s.evClient.Send(ctx, runnerID, &signals.Signal{
		Type: signals.OperationOfflineCheck,
	})

	ctx.JSON(http.StatusCreated, true)
}
