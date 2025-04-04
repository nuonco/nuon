package service

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
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
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}

	s.evClient.Send(ctx, "general", &signals.Signal{
		Type: signals.OperationTerminateEventLoops,
	})
	s.evClient.Send(ctx, "general", &signals.Signal{
		Type: signals.OperationSeed,
	})
}
