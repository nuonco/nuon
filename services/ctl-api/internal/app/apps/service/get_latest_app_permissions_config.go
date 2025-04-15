package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetLatestAppPermissionsConfig
// @Summary				get latest app permissions config
// @Description.markdown	get_latest_app_permissions_config.md
// @Param					app_id	path	string	true	"app ID"
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
// @Router					/v1/apps/{app_id}/latest-app-permissions-config [get]
func (s *service) GetLatestAppPermissionsConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.getLatestAppPermissionsConfig(ctx, currentApp.ID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get permissions config"))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) getLatestAppPermissionsConfig(ctx context.Context, appID string) (*app.AppPermissionsConfig, error) {
	var appPermissionsCfg app.AppPermissionsConfig

	res := s.db.WithContext(ctx).
		Where(app.AppPermissionsConfig{
			AppID: appID,
		}).
		Preload("AWSIAMRoles").
		Order("created_at desc").
		Limit(1).
		First(&appPermissionsCfg)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app permissions config")
	}

	return &appPermissionsCfg, nil
}
