package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

//	@ID						AdminGetRunnerSettings
//	@Summary				get a runner settings
//	@Description.markdown	admin_get_runner_settings.md
//	@Param					runner_id	path	string	true	"runner ID"
//	@Tags					runners/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Produce				json
//	@Success				200	{object}	app.RunnerGroupSettings
//	@Router					/v1/runners/{runner_id}/settings [GET]
func (s *service) AdminGetRunnerSettings(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runner.RunnerGroup.Settings)
}
