package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

type AdminReprovisionRunnerRequest struct{}

// @ID						AdminReprovisionRunner
// @Summary					reprovision a runner, but keep it in the database
// @Description.markdown	reprovision_runner.md
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req			body	AdminReprovisionRunnerRequest	true	"Input"
// @Param					runner_id	path	string							true	"runner ID to reprovision"
// @Produce					json
// @Success					200	{string}	ok
// @Router					/v1/runners/{runner_id}/reprovision [POST]
// @Deprecated 				true
func (s *service) AdminReprovisionRunner(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner: %w", err))
		return
	}

	s.evClient.Send(ctx, runner.ID, &signals.Signal{
		Type: signals.OperationReprovision,
	})
	ctx.JSON(http.StatusOK, true)
}
