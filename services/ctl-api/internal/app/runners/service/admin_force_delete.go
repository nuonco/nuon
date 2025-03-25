package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

type AdminForceDeleteRunnerRequest struct{}

//	@ID						AdminForceDeleteRunner
//	@Summary				force delete a runner
//	@Description.markdown	force_delete_runner.md
//	@Tags					runners/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					req			body	AdminForceDeleteRunnerRequest	true	"Input"
//	@Param					runner_id	path	string							true	"runner ID to force delete"
//	@Produce				json
//	@Success				200	{string}	ok
//	@Router					/v1/runners/{runner_id}/force-delete [POST]
func (s *service) AdminForceDeleteRunner(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner: %w", err))
		return
	}

	s.evClient.Send(ctx, runner.ID, &signals.Signal{
		Type: signals.OperationForceDelete,
	})
	ctx.JSON(http.StatusOK, true)
}
