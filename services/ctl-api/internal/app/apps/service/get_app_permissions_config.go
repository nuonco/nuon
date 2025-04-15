package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetAppPermissionsConfig
// @Summary				get app permissions config
// @Description.markdown	get_app_permissions_config.md
// @Param					app_id	path	string	true	"app ID"
// @Param permissions_config_id path string	true	"input config ID"
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.AppPermissionsConfig
// @Router /v1/apps/{app_id}/permissions-configs/{permissions_config_id} [get]
func (s *service) GetAppPermissionsConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	appPermissionsConfigID := ctx.Param("permissions_config_id")

	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.getAppPermissionsConfig(ctx, currentApp.ID, appPermissionsConfigID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get permissions config"))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) getAppPermissionsConfig(ctx context.Context, appID, appPermissionsConfigID string) (*app.AppPermissionsConfig, error) {
	var appPermissionsCfg app.AppPermissionsConfig

	res := s.db.WithContext(ctx).
		Where(app.AppPermissionsConfig{
			AppID: appID,
			ID:    appPermissionsConfigID,
		}).
		Preload("AWSIAMRoles").
		Preload("AWSIAMRoles.AppAWSIAMPolicyConfigs").
		Order("created_at desc").
		Limit(1).
		First(&appPermissionsCfg)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app permissions config")
	}

	return &appPermissionsCfg, nil
}
