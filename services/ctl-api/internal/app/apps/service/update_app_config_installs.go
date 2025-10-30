package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
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
// @Param					app_config_id	path	string	true	"app config ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{string}	ok
// @Router					/v1/apps/{app_id}/config/{app_config_id}/update-installs [POST]
func (s *service) UpdateAppConfigInstallsV2(ctx *gin.Context) {
	s.UpdateAppConfigInstalls(ctx)
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
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
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
	res := s.db.WithContext(ctx).
		Model(&app.Install{}).
		Where(app.Install{
			AppID: appID,
		})

	// if "all" is false, filter by the provided install IDs
	if !req.UpdateAll {
		res.Where("id in ?", req.InstallIDs)
	}

	res.Updates(app.Install{
		AppConfigID: appConfigID,
	})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update installations with new config")
	}

	return nil
}
