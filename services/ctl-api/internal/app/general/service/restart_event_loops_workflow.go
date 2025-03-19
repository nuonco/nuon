package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
)

type RestartGeneralEventLoopRequest struct{}

// @ID RestartGeneralEventLoop
// @Summary	restart event loop reconciliation cron job
// @Description.markdown	restart_general_event_loop.md
// @Param			req	body	RestartGeneralEventLoopRequest	true	"Input"
// @Tags			general/admin
// @Accept			json
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/general/restart-event-loop [post]
func (s *service) RestartGeneralEventLoop(ctx *gin.Context) {
	var req RestartGeneralEventLoopRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}

	s.evClient.Send(ctx, "general", &signals.Signal{
		Type: signals.OperationRestart,
	})

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
		"type":   string(signals.OperationRestart),
	})
}
