package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetAppSecretsConfig
// @Summary				get app secrets config
// @Description.markdown	get_app_secrets_config.md
// @Param					app_id	path	string	true	"app ID"
// @Param app_secrets_config_id path string	true	"app secrets config ID"
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
// @Success				200	{object}	app.AppSecretsConfig
// @Router					/v1/apps/{app_id}/secrets-configs/{app_secrets_config_id} [get]
func (s *service) GetAppSecretsConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	appSecretsConfigID := ctx.Param("secrets_config_id")

	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.getAppSecretsConfig(ctx, currentApp.ID, appSecretsConfigID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get secrets config"))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) getAppSecretsConfig(ctx context.Context, appID, appSecretsConfigID string) (*app.AppSecretsConfig, error) {
	var appSecretsCfg app.AppSecretsConfig

	res := s.db.WithContext(ctx).
		Where(app.AppSecretsConfig{
			AppID: appID,
			ID:    appSecretsConfigID,
		}).
		Preload("Secrets").
		First(&appSecretsCfg)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app secrets config")
	}

	return &appSecretsCfg, nil
}
