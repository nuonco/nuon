package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
)

type ReconcileEventLoopsRequest struct {
}

// @ID ReconcileEventLoops
// @Summary	restart event loop reconciliation cron job
// @Description.markdown	reconcile_event_loops.md
// @Param			req	body	ReconcileEventLoopsRequest	true	"Input"
// @Tags			general/admin
// @Accept			json
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/general/reconcile-event-loops [post]
func (s *service) ReconcileEventLoops(ctx *gin.Context) {
	var req ReconcileEventLoopsRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}
	// use the event loop name as a constant
	s.evClient.Send(ctx, "general", &signals.Signal{
		Type: signals.OperationReconcile,
	})

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
		"type":   string(signals.OperationReconcile),
	})
}
