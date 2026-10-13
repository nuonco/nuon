package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type InstallTelemetrySettings struct {
	Enabled         bool  `json:"enabled"`
	Override        *bool `json:"override" extensions:"x-nullable"`
	OrgDefault      bool  `json:"org_default"`
	RelayConfigured bool  `json:"relay_configured"`
}

// @ID GetInstallTelemetrySettings
// @Summary Get an install's telemetry settings
// @Tags installs
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param install_id path string true "Install ID"
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Success 200 {object} InstallTelemetrySettings
// @Router /v1/installs/{install_id}/telemetry [get]
func (s *service) GetInstallTelemetrySettings(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	settings, err := s.getInstallTelemetrySettings(ctx, org.ID, ctx.Param("install_id"))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install telemetry settings: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (s *service) getInstallTelemetrySettings(ctx context.Context, orgID, installID string) (*InstallTelemetrySettings, error) {
	config, err := s.helpers.GetInstallTelemetryConfig(ctx, orgID, installID)
	if err != nil {
		return nil, err
	}
	installConfig := app.InstallConfig{TelemetryEnabled: config.TelemetryEnabled}
	orgTelemetry := app.OrgTelemetrySettings{RelayEndpoint: config.OrgRelayEndpoint}
	settings := &InstallTelemetrySettings{
		Enabled:         installConfig.IsTelemetryEnabled(config.OrgTelemetryEnabled),
		Override:        config.TelemetryEnabled,
		OrgDefault:      config.OrgTelemetryEnabled,
		RelayConfigured: app.ValidateTelemetryRelayEndpoint(orgTelemetry.ResolveRelayEndpoint(s.cfg.TelemetryRelayEndpoint)) == nil,
	}
	return settings, nil
}
