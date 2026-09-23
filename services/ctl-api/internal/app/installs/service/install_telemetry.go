package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type InstallTelemetrySettings struct {
	Enabled    bool  `json:"enabled"`
	Override   *bool `json:"override" extensions:"x-nullable"`
	OrgDefault bool  `json:"org_default"`
}

type UpdateInstallTelemetryRequest struct {
	Enabled *bool `json:"enabled" extensions:"x-nullable,!x-omitempty"`
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

// @ID UpdateInstallTelemetrySettings
// @Summary Update an install's telemetry settings
// @Tags installs
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param install_id path string true "Install ID"
// @Param req body UpdateInstallTelemetryRequest true "Input"
// @Failure 400 {object} stderr.ErrResponse
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Success 200 {object} InstallTelemetrySettings
// @Router /v1/installs/{install_id}/telemetry [patch]
func (s *service) UpdateInstallTelemetrySettings(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req UpdateInstallTelemetryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	patch := cctx.PatcherFromContext(ctx)
	if patch == nil || !slices.Contains(patch.SelectFields, "enabled") {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("provide enabled as true, false, or null")))
		return
	}

	if err := s.helpers.SetInstallTelemetry(ctx, ctx.Param("install_id"), req.Enabled); err != nil {
		ctx.Error(fmt.Errorf("unable to update install telemetry settings: %w", err))
		return
	}
	settings, err := s.getInstallTelemetrySettings(ctx, org.ID, ctx.Param("install_id"))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update install telemetry settings: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (s *service) getInstallTelemetrySettings(ctx context.Context, orgID, installID string) (*InstallTelemetrySettings, error) {
	var install app.Install
	if err := s.db.WithContext(ctx).Preload("InstallConfig").Preload("Org").
		Where(app.Install{ID: installID, OrgID: orgID}).First(&install).Error; err != nil {
		return nil, err
	}
	settings := &InstallTelemetrySettings{
		Enabled:    install.InstallConfig.IsTelemetryEnabled(install.Org.Telemetry.Enabled),
		OrgDefault: install.Org.Telemetry.Enabled,
	}
	if install.InstallConfig != nil {
		settings.Override = install.InstallConfig.TelemetryEnabled
	}
	return settings, nil
}
