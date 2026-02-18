package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type UpdateAppConfigInstallsRequest struct {
	UpdateAll  bool
	InstallIDs []string
}

func (c *UpdateAppConfigInstallsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	return nil
}

// @ID						UpdateAppConfigInstallsV2
// @Description.markdown	update_app_config_installs.md
// @Tags					apps
// @Accept					json
// @Param					req	body	UpdateAppConfigInstallsRequest	true	"Input"
// @Produce				json
// @Param					app_id			path	string	true	"app ID"
// @Param					config_id	path	string	true	"app config ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{string}	ok
// @Router					/v1/apps/{app_id}/configs/{config_id}/update-installs [POST]
func (s *service) UpdateAppConfigInstallsV2(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	appConfigID := ctx.Param("config_id")

	var req UpdateAppConfigInstallsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	err := s.updateAppConfigInstalls(ctx, appID, appConfigID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update app config installs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, "ok")
}

// @ID						UpdateAppConfigInstalls
// @Description.markdown	update_app_config_installs.md
// @Tags					apps
// @Accept					json
// @Param					req	body	UpdateAppConfigInstallsRequest	true	"Input"
// @Produce				json
// @Param					app_id			path	string	true	"app ID"
// @Param					app_config_id	path	string	true	"app config ID"
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{string}	ok
// @Router					/v1/apps/{app_id}/config/{app_config_id}/update-installs [POST]
func (s *service) UpdateAppConfigInstalls(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	appConfigID := ctx.Param("app_config_id")

	var req UpdateAppConfigInstallsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	err := s.updateAppConfigInstalls(ctx, appID, appConfigID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update app config installs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, "ok")
}

func (s *service) updateAppConfigInstalls(ctx context.Context, appID, appConfigID string, req *UpdateAppConfigInstallsRequest) error {
	var affectedInstalls []app.Install
	query := s.db.WithContext(ctx).
		Model(&app.Install{}).
		Where(app.Install{AppID: appID})

	if !req.UpdateAll {
		query = query.Where("id in ?", req.InstallIDs)
	}

	if err := query.Find(&affectedInstalls).Error; err != nil {
		return stderr.ErrSystem{
			Err:         fmt.Errorf("unable to query affected installs: %w", err),
			Description: "Failed to retrieve installs for migration",
		}
	}

	// install ID -> old app_config_id
	installConfigMap := make(map[string]string)
	for _, install := range affectedInstalls {
		installConfigMap[install.ID] = install.AppConfigID
	}

	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// migrate install inputs BEFORE updating app_config_id
	if err := s.installsHelpers.MigrateInstallInputsToNewConfig(ctx, tx, installConfigMap, appConfigID); err != nil {
		tx.Rollback()
		return stderr.ErrSystem{
			Err:         fmt.Errorf("unable to migrate install inputs: %w", err),
			Description: "Failed to migrate install inputs to new app config",
		}
	}

	// update install app_config_id
	res := tx.Model(&app.Install{}).
		Where(app.Install{AppID: appID})

	if !req.UpdateAll {
		res = res.Where("id in ?", req.InstallIDs)
	}

	if err := res.Updates(app.Install{AppConfigID: appConfigID}).Error; err != nil {
		tx.Rollback()
		return stderr.ErrSystem{
			Err:         fmt.Errorf("unable to update installs: %w", err),
			Description: "Failed to update install app config references",
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return stderr.ErrSystem{
			Err:         fmt.Errorf("unable to commit transaction: %w", err),
			Description: "Failed to commit install updates",
		}
	}

	return nil
}
