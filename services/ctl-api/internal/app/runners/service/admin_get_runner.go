package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

//	@ID						AdminGetRunner
//	@Summary				get a runner
//	@Description.markdown	get_runner.md
//	@Tags					runners/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					runner_id	path	string	true	"runner ID to fetc"
//	@Produce				json
//	@Success				200	{object}	app.Runner
//	@Router					/v1/runners/{runner_id} [GET]
func (s *service) AdminGetRunner(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runner)
}
