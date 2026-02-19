package service

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminTerminateEventLoopsRequest struct{}

// @ID						AdminTerminateEventLoops
// @Summary				terminate event loops.
// @Description.markdown terminate_event_loops.md
// @Param					req	body	AdminTerminateEventLoopsRequest	true	"Input"
// @Tags					general/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/general/terminate-event-loops [POST]
func (s *service) AdminTerminateEventLoops(ctx *gin.Context) {
	var req AdminTerminateEventLoopsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	s.evClient.Send(ctx, "general", &signals.Signal{
		Type: signals.OperationTerminateEventLoops,
	})

	ctx.JSON(http.StatusCreated, "ok")
}
