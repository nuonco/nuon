package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
	"gorm.io/gorm"
)

// @ID						GetAppInputConfigs
// @Summary				get app input configs
// @Description.markdown	get_app_input_configs.md
// @Param					app_id						path	string	true	"app ID"
// @Param					offset						query	int		false	"offset of jobs to return"	Default(0)
// @Param					limit						query	int		false	"limit of jobs to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
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
// @Success				200	{object}	[]app.AppInputConfig
// @Router					/v1/apps/{app_id}/input-configs [get]
func (s *service) GetAppInputConfigs(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	cfgs, err := s.findAppInputConfigs(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	ctx.JSON(http.StatusOK, cfgs)
}

func (s *service) findAppInputConfigs(ctx *gin.Context, orgID, appID string) ([]app.AppInputConfig, error) {
	app := app.App{}
	res := s.db.WithContext(ctx).
		Preload("Org").
		Preload("Components").
		Preload("AppInputConfigs", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOffsetPagination).
				Order("app_input_configs.created_at DESC")
		}).
		Preload("AppInputConfigs.AppInputs").
		Preload("AppInputConfigs.AppInputs.AppInputGroup").
		Preload("AppInputConfigs.AppInputGroups.AppInputs").
		Where("name = ? AND org_id = ?", appID, orgID).
		Or("id = ?", appID).
		First(&app)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	cfgs, err := db.HandlePaginatedResponse(ctx, app.AppInputConfigs)
	if err != nil {
		return nil, fmt.Errorf("unable to get app input configs: %w", err)
	}

	app.AppInputConfigs = cfgs

	return app.AppInputConfigs, nil
}
