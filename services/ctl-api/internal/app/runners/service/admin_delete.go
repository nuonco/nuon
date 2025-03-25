package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

type AdminDeleteRunnerRequest struct{}

//	@ID						AdminDeleteRunner
//	@Summary				delete a runner database
//	@Description.markdown	delete_runner.md
//	@Tags					runners/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					req			body	AdminDeleteRunnerRequest	true	"Input"
//	@Param					runner_id	path	string						true	"runner ID to delete"
//	@Produce				json
//	@Success				200	{string}	ok
//	@Router					/v1/runners/{runner_id}/delete [POST]
func (s *service) AdminDeleteRunner(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner: %w", err))
		return
	}

	s.evClient.Send(ctx, runner.ID, &signals.Signal{
		Type: signals.OperationDelete,
	})
	ctx.JSON(http.StatusOK, true)
}
