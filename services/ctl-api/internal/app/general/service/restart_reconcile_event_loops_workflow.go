package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
)

type RestartEventLoopReconcileCronRequest struct {
}

// @ID RestartEventLoopReconcileCron
// @Summary	restart event loop reconciliation cron job
// @Description.markdown	restart_event_loop_reconcile_cron.md
// @Param			req	body	RestartEventLoopReconcileCronRequest	true	"Input"
// @Tags			general/admin
// @Accept			json
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/general/restart-event-loop-reconcile-cron [post]
func (s *service) RestartEventLoopReconcileCron(ctx *gin.Context) {
	var req RestartEventLoopReconcileCronRequest
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
