package service

import (
	"errors"
	"io"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type SeedRequest struct{}

// @ID Seed
// @Summary				seed
// @Description.markdown	seed.md
// @Param					req	body	SeedRequest	true	"Input"
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/general/seed [post]
func (s *service) Seed(ctx *gin.Context) {
	var req RestartGeneralEventLoopRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	s.evClient.Send(ctx, "general", &signals.Signal{
		Type: signals.OperationTerminateEventLoops,
	})
	s.evClient.Send(ctx, "general", &signals.Signal{
		Type: signals.OperationSeed,
	})
}
