package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RunnerPublicSettings struct {
	BinaryVersion string `json:"binary_version"`
}

// @ID						GetRunnerPublicSettings
// @summary				get runner public settings
// @Description.markdown	get_runner_public_settings.md
// @Param					runner_id	path	string	true	"runner ID"
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	RunnerPublicSettings
// @Router					/v1/runners/{runner_id}/public-settings [get]
func (s *service) GetRunnerPublicSettings(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	settings := RunnerPublicSettings{
		BinaryVersion: runner.RunnerGroup.Settings.BinaryVersion,
	}

	ctx.JSON(http.StatusOK, settings)
}
