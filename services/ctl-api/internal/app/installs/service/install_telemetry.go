package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type InstallTelemetrySettings struct {
	Enabled bool `json:"enabled"`
}

type UpdateInstallTelemetryRequest struct {
	Enabled *bool `json:"enabled" validate:"required"`
}

func (r *UpdateInstallTelemetryRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
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
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	settings, err := s.updateInstallTelemetrySettings(ctx, org.ID, ctx.Param("install_id"), *req.Enabled)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update install telemetry settings: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (s *service) getInstallTelemetrySettings(ctx context.Context, orgID, installID string) (*InstallTelemetrySettings, error) {
	runnerGroup, err := s.getInstallRunnerGroupForTelemetry(ctx, orgID, installID)
	if err != nil {
		return nil, err
	}
	return &InstallTelemetrySettings{Enabled: runnerGroup.Settings.VendorTelemetryEnabled}, nil
}

func (s *service) updateInstallTelemetrySettings(ctx context.Context, orgID, installID string, enabled bool) (*InstallTelemetrySettings, error) {
	runnerGroup, err := s.getInstallRunnerGroupForTelemetry(ctx, orgID, installID)
	if err != nil {
		return nil, err
	}

	result := s.db.WithContext(ctx).
		Model(&app.RunnerGroupSettings{}).
		Where(app.RunnerGroupSettings{
			ID:            runnerGroup.Settings.ID,
			OrgID:         orgID,
			RunnerGroupID: runnerGroup.ID,
		}).
		Update("vendor_telemetry_enabled", enabled)
	if result.Error != nil {
		return nil, fmt.Errorf("update runner group telemetry settings: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return nil, fmt.Errorf("update runner group telemetry settings: %w", gorm.ErrRecordNotFound)
	}

	return &InstallTelemetrySettings{Enabled: enabled}, nil
}

func (s *service) getInstallRunnerGroupForTelemetry(ctx context.Context, orgID, installID string) (*app.RunnerGroup, error) {
	var runnerGroup app.RunnerGroup
	result := s.db.WithContext(ctx).
		Preload("Settings").
		Where(app.RunnerGroup{
			OrgID:     orgID,
			OwnerID:   installID,
			OwnerType: plugins.TableName(s.db, app.Install{}),
			Type:      app.RunnerGroupTypeInstall,
		}).
		First(&runnerGroup)
	if result.Error != nil {
		return nil, fmt.Errorf("get install runner group: %w", result.Error)
	}
	if runnerGroup.Settings.ID == "" {
		return nil, fmt.Errorf("get install runner group settings: %w", gorm.ErrRecordNotFound)
	}
	return &runnerGroup, nil
}
