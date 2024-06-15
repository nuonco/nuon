package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

// @ID GetAppConfigs
// @Summary	get app configs
// @Description.markdown	get_app_configs.md
// @Param			app_id	path	string	true	"app ID"
// @Tags			apps
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	[]app.AppConfig
// @Router			/v1/apps/{app_id}/configs [get]
func (s *service) GetAppConfigs(ctx *gin.Context) {
	org, err := middlewares.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	cfgs, err := s.getAppConfigs(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, cfgs)
}

func (s *service) getAppConfigs(ctx context.Context, orgID, appID string) ([]app.AppConfig, error) {
	cfgs := make([]app.AppConfig, 0)

	res := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Where(app.AppConfig{
			OrgID: orgID,
			AppID: appID,
		}).
		Order("created_at desc").
		Find(&cfgs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app configs: %w", res.Error)
	}

	return cfgs, nil
}
