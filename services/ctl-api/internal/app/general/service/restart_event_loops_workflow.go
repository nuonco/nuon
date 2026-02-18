package service

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type RestartGeneralEventLoopRequest struct{}

// @ID						RestartGeneralEventLoop
// @Summary				restart event loop reconciliation cron job
// @Description.markdown	restart_general_event_loop.md
// @Param					req	body	RestartGeneralEventLoopRequest	true	"Input"
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/general/restart-event-loop [post]
func (s *service) RestartGeneralEventLoop(ctx *gin.Context) {
	var req RestartGeneralEventLoopRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
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
