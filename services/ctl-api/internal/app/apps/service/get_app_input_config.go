package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetAppInputConfig
// @Summary				get app input config
// @Description.markdown	get_app_input_config.md
// @Param					app_id	path	string	true	"app ID"
// @Param					input_config_id	path	string	true	"input config ID"
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
// @Success				200	{object}	app.AppInputConfig
// @Router					/v1/apps/{app_id}/input-configs/{input_config_id} [get]
func (s *service) GetAppInputConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	inputConfigID := ctx.Param("input_config_id")
	app, err := s.findApp(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to find app"))
		return
	}

	cfg, err := s.getAppInputConfig(ctx, app.ID, inputConfigID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get app input config"))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) getAppInputConfig(ctx context.Context, appID, inputConfigID string) (*app.AppInputConfig, error) {
	var appInputConfig app.AppInputConfig

	if res := s.db.WithContext(ctx).
		Where(app.AppInputConfig{
			AppID: appID,
			ID:    inputConfigID,
		}).
		Preload("AppInputs").
		Preload("AppInputs.AppInputGroup").
		Preload("AppInputGroups.AppInputs").
		First(&appInputConfig); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get input config")
	}

	return &appInputConfig, nil
}
