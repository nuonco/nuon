package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetAppStackConfig
// @Summary				get app stack config
// @Description.markdown	get_app_stack_config.md
// @Param		app_id	path	string	true	"app ID"
// @Param config_id path string	true	"app stack config ID"
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
// @Success				200	{object}	app.AppStackConfig
// @Router /v1/apps/{app_id}/stack-configs/{config_id} [get]
func (s *service) GetAppStackConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	configID := ctx.Param("config_id")

	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.getAppStackConfig(ctx, currentApp.ID, configID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get cloudformation stack config"))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) getAppStackConfig(ctx context.Context, appID, configID string) (*app.AppStackConfig, error) {
	var cfg app.AppStackConfig

	res := s.db.WithContext(ctx).
		Where(app.AppStackConfig{
			AppID: appID,
			ID:    configID,
		}).
		First(&cfg)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app cloudformation stack config")
	}

	return &cfg, nil
}
