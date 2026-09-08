package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

// @ID						GetRunnerSettings
// @Summary				get runner settings
// @Description.markdown	get_runner_settings.md
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
// @Success				200	{object}	app.RunnerGroupSettings
// @Router					/v1/runners/{runner_id}/settings [get]
func (s *service) GetRunnerSettings(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	settings := runner.RunnerGroup.Settings
	settings.LongPollJobs = true
	if s.telemetryRelayEndpoint != "" && settings.VendorTelemetryEnabled && runner.RunnerGroup.Type == app.RunnerGroupTypeInstall && runner.RunnerGroup.OwnerType == plugins.TableName(s.db, app.Install{}) {
		settings.TelemetryRelayEndpoint = s.telemetryRelayEndpoint
	} else {
		settings.VendorTelemetryEnabled = false
	}
	ctx.JSON(http.StatusOK, settings)
}
