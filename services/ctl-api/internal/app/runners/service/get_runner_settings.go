package service

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	"gorm.io/gorm"
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
	installTable := plugins.TableName(s.db, app.Install{})
	if s.telemetryRelayEndpoint != "" && settings.VendorTelemetryEnabled && runner.RunnerGroup.Type == app.RunnerGroupTypeInstall && runner.RunnerGroup.OwnerType == installTable {
		// A projection avoids model AfterQuery hooks, which also run with SkipHooks.
		var install struct {
			Name    string
			Labels  labels.Labels
			AppName string `gorm:"column:App__name"`
		}
		err := s.db.WithContext(ctx).
			Model(&app.Install{}).
			Scopes(scopes.WithDisableViews).
			Select(installTable+".name", installTable+".labels").
			Joins("App", s.db.Select("name")).
			Where(app.Install{ID: runner.RunnerGroup.OwnerID, OrgID: runner.OrgID}).
			Take(&install).Error
		if err == nil {
			settings.TelemetryRelayEndpoint = s.telemetryRelayEndpoint
			settings.VendorTelemetryResourceAttributes = map[string]string{
				"nuon.org.name":     runner.Org.Name,
				"nuon.app.name":     install.AppName,
				"nuon.install.name": install.Name,
			}
			for key, value := range install.Labels {
				settings.VendorTelemetryResourceAttributes["nuon.install.labels."+key] = value
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			settings.VendorTelemetryEnabled = false
			s.l.Warn("vendor telemetry disabled: owner install not found in runner org",
				zap.String("runner_id", runner.ID),
				zap.String("owner_id", runner.RunnerGroup.OwnerID),
				zap.String("org_id", runner.OrgID),
			)
		} else {
			ctx.Error(err)
			return
		}
	} else {
		settings.VendorTelemetryEnabled = false
	}
	ctx.JSON(http.StatusOK, settings)
}
