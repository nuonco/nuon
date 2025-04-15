package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetLatestAppPoliciesConfig
// @Summary				get latest app policies config
// @Description.markdown	get_latest_app_policies_config.md
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
// @Success				200	{object}	app.AppPoliciesConfig
// @Router					/v1/apps/{app_id}/latest-app-policies-config [get]
func (s *service) GetLatestAppPoliciesConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.getLatestAppPoliciesConfig(ctx, currentApp.ID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get policies config"))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) getLatestAppPoliciesConfig(ctx context.Context, appID string) (*app.AppPoliciesConfig, error) {
	var appPoliciesCfg app.AppPoliciesConfig

	res := s.db.WithContext(ctx).
		Where(app.AppPoliciesConfig{
			AppID: appID,
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
