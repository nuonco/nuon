package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals/promotion"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type RestartGeneralEventLoopRequest struct{}

// @ID						RestartGeneralEventLoop
// @Summary				restart event loop reconciliation cron job
// @Description.markdown	restart_general_event_loop.md
// @Param					req	body	RestartGeneralEventLoopRequest	true	"Input"
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/general/restart-event-loop [post]
func (s *service) RestartGeneralEventLoop(ctx *gin.Context) {
	q, err := s.generalHelpers.EnsureGeneralQueue(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to ensure general queue: %w", err))
		return
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal:  &promotion.Signal{},
	}); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue promotion signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
