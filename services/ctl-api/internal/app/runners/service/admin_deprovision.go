package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

type AdminDeprovisionRunnerRequest struct{}

//	@ID						AdminDeprovisionRunner
//	@Summary				deprovision a runner, but keep it in the database
//	@Description.markdown	deprovision_runner.md
//	@Tags					runners/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					req			body	AdminDeprovisionRunnerRequest	true	"Input"
//	@Param					runner_id	path	string							true	"runner ID to deprovision"
//	@Produce				json
//	@Success				201	{string}	ok
//	@Router					/v1/runners/{runner_id}/deprovision [POST]
func (s *service) AdminDeprovisionRunner(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner: %w", err))
		return
	}

	s.evClient.Send(ctx, runner.ID, &signals.Signal{
		Type: signals.OperationDeprovision,
	})
	ctx.JSON(http.StatusOK, true)
}
