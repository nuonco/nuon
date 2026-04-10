package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @ID						GetAppPolicyConfig
// @Summary				get app policy config
// @Description			get a single app policy config by ID
// @Param					app_id				path	string	true	"app ID"
// @Param					policy_config_id	path	string	true	"app policy config ID"
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
// @Success				200	{object}	app.AppPolicyConfig
// @Router					/v1/apps/{app_id}/policy-config/{policy_config_id} [get]
func (s *service) GetAppPolicyConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	policyConfigID := ctx.Param("policy_config_id")

	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.getAppPolicyConfig(ctx, currentApp.ID, policyConfigID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get policy config"))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) getAppPolicyConfig(ctx context.Context, appID, policyConfigID string) (*app.AppPolicyConfig, error) {
	var policyConfig app.AppPolicyConfig

	res := s.db.WithContext(ctx).
		Where("app_id = ? AND id = ?", appID, policyConfigID).
		First(&policyConfig)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app policy config")
	}

	return &policyConfig, nil
}
