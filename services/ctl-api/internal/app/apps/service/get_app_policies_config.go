package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetAppPoliciesConfig
// @Summary				get app policies config
// @Description.markdown	get_app_policies_config.md
// @Param					app_id	path	string	true	"app ID"
// @Param policies_config_id path string	true	"input policies config ID"
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
// @Success				200	{object}	app.AppPoliciesConfig
// @Router					/v1/apps/{app_id}/policies-configs/{policies_config_id} [get]
func (s *service) GetAppPoliciesConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	appPoliciesConfigID := ctx.Param("policies_config_id")

	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.getAppPoliciesConfig(ctx, currentApp.ID, appPoliciesConfigID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get policies config"))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) getAppPoliciesConfig(ctx context.Context, appID, appPoliciesConfigID string) (*app.AppPoliciesConfig, error) {
	var appPoliciesCfg app.AppPoliciesConfig

	res := s.db.WithContext(ctx).
		Where(app.AppPoliciesConfig{
			AppID: appID,
			ID:    appPoliciesConfigID,
		}).
		Preload("Policies").
		Order("created_at desc").
		Limit(1).
		First(&appPoliciesCfg)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app policies config")
	}

	return &appPoliciesCfg, nil
}
