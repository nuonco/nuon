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

// @ID						GetApps
// @Summary				get all apps for the current org
// @Description.markdown	get_apps.md
// @Param					offset						query	int		false	"offset of jobs to return"	Default(0)
// @Param					q							query	string	false	"search query to filter apps by name"
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
// @Success				200	{array}		app.App
// @Router					/v1/apps [get]
func (s *service) GetApps(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	q := ctx.Query("q")

	apps, err := s.getApps(ctx, org.ID, q)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get apps for %s: %w", org.ID, err))
		return
	}
	ctx.JSON(http.StatusOK, apps)
}

func (s *service) getApps(ctx *gin.Context, orgID, q string) ([]*app.App, error) {
	var apps []*app.App
	org := &app.Org{
		ID: orgID,
	}

	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("AppConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(scopes.WithOverrideTable("app_configs_latest_view_v1"))
		}).
		Preload("AppRunnerConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(scopes.WithOverrideTable("app_runner_configs_latest_view_v1"))
		}).
		Preload("AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(scopes.WithOverrideTable("app_sandbox_configs_latest_view_v1"))
		}).
		Preload("AppSandboxConfigs.PublicGitVCSConfig").
		Preload("AppSandboxConfigs.ConnectedGithubVCSConfig").
		Order("apps.name ASC")

	if q != "" {
		tx = tx.Where("apps.name ILIKE ?", "%"+q+"%")
	}

	err := tx.Model(&org).Association("Apps").Find(&apps)
	if err != nil {
		return nil, fmt.Errorf("unable to get org apps: %w", err)
	}

	apps, err = db.HandlePaginatedResponse(ctx, apps)
	if err != nil {
		return nil, fmt.Errorf("unable to get org apps: %w", err)
	}

	return apps, nil
}
