package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
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
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(errors.Wrap(err, "unable to promote"))
		return
	}

	s.evClient.Send(ctx, "general", &signals.Signal{
		Type: signals.OperationTerminateEventLoops,
	})

	ctx.JSON(http.StatusCreated, "ok")
}
