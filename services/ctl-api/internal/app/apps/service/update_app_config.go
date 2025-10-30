package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type UpdateAppConfigRequest struct {
	Status            app.AppConfigStatus `json:"status"`
	StatusDescription string              `json:"status_description"`
	State             string              `json:"state"`
	ComponentIDs      []string            `json:"component_ids"`
}

func (c *UpdateAppConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	return nil
}

// @ID						UpdateAppConflgV2
// @Description.markdown	update_app_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	UpdateAppConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id			path	string	true	"app ID"
// @Param					app_config_id	path	string	true	"app config ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppConfig
// @Router					/v1/apps/{app_id}/config/{app_config_id} [PATCH]
func (s *service) UpdateAppConfigV2(ctx *gin.Context) {
	s.UpdateAppConfig(ctx)
}

// @ID						UpdateAppConfig
// @Description.markdown	update_app_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	UpdateAppConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id			path	string	true	"app ID"
// @Param					app_config_id	path	string	true	"app config ID"
// @Security				APIKey
// @Security				OrgID
// @Deprecation    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppConfig
// @Router					/v1/apps/{app_id}/config/{app_config_id} [PATCH]
func (s *service) UpdateAppConfig(ctx *gin.Context) {
	appConfigID := ctx.Param("app_config_id")

	var req UpdateAppConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	cfg, err := s.updateAppConfig(ctx, appConfigID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app inputs config: %w", err))
		return
	}

	// Update user journey step for first app sync when status becomes active
	if req.Status == app.AppConfigStatusActive {
		user, err := cctx.AccountFromGinContext(ctx)
		if err == nil && cfg != nil {
			// Only update if this is the user's first app sync (app_synced step incomplete)
			if err := s.accountsHelpers.UpdateUserJourneyStepForFirstAppSync(ctx, user.ID, cfg.AppID); err != nil {
				// Log but don't fail the update
				s.l.Warn("failed to update user journey for first app sync",
					zap.String("account_id", user.ID),
					zap.String("app_id", cfg.AppID),
					zap.String("app_config_id", cfg.ID),
					zap.Error(err))
			}
		}
	}

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) updateAppConfig(ctx context.Context, appConfigID string, req *UpdateAppConfigRequest) (*app.AppConfig, error) {
	var cfg app.AppConfig
	if err := s.db.WithContext(ctx).
		Where("id = ?", appConfigID).
		First(&cfg).Error; err != nil {
		return nil, fmt.Errorf("unable to find app config: %w", err)
	}

	appConfig := app.AppConfig{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
		State:             req.State,
	}

	if len(req.ComponentIDs) > 0 {
		appConfig.ComponentIDs = req.ComponentIDs
	}

	res := s.db.WithContext(ctx).
		Model(&cfg).
		Updates(appConfig)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update app config: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return nil, fmt.Errorf("app config not found %s %w", appConfigID, gorm.ErrRecordNotFound)
	}

	return &cfg, nil
}
