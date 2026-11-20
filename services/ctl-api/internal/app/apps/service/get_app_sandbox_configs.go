package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	applookup "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/lookup"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAppSandboxConfigs
// @Summary				get app sandbox configs
// @Description.markdown	get_app_sandbox_configs.md
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
// @Success				200	{object}	[]app.AppSandboxConfig
// @Router					/v1/apps/{app_id}/sandbox-configs [get]
func (s *service) GetAppSandboxConfigs(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	cfgs, err := s.findAppSandboxConfigs(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	if len(cfgs) < 1 {
		ctx.Error(fmt.Errorf("no app sandbox configs found for app"))
		return
	}

	ctx.JSON(http.StatusOK, cfgs)
}

func (s *service) findAppSandboxConfigs(ctx *gin.Context, orgID, appID string) ([]app.AppSandboxConfig, error) {
	currentApp, err := applookup.ByRef(ctx, s.db, orgID, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app: %w", err)
	}
	res := s.db.WithContext(ctx).
		// sandbox configs
		Preload("AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOffsetPagination).
				Order("app_sandbox_configs.created_at DESC")
		}).
		Preload("AppSandboxConfigs.PublicGitVCSConfig").
		Preload("AppSandboxConfigs.ConnectedGithubVCSConfig").
		Where(&app.App{ID: currentApp.ID, OrgID: currentApp.OrgID}, "ID", "OrgID").
		First(currentApp)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	cfgs, err := db.HandlePaginatedResponse(ctx, currentApp.AppSandboxConfigs)
	if err != nil {
		return nil, fmt.Errorf("unable to get app sandbox configs: %w", err)
	}

	currentApp.AppSandboxConfigs = cfgs

	return currentApp.AppSandboxConfigs, nil
}
