package service

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAppComponentBuilds
// @Summary				get builds for components
// @Description.markdown	get_component_builds.md
// @Param					app_id						path	string	true	"app id to filter by"
// @Param					component_id				path	string	true	"component id to filter by"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.ComponentBuild
// @Router					/v1/apps/{app_id}/components/{component_id}/builds [GET]
func (s *service) GetAppComponentBuilds(ctx *gin.Context) (any, error) {
	return s.getComponentBuilds(ctx, ctx.Param("component_id"), ctx.Param("app_id"))
}

// @ID						GetComponentBuilds
// @Summary				get builds for components
// @Description.markdown	get_component_builds.md
// @Param					component_id				query	string	false	"component id to filter by"
// @Param					app_id						query	string	false	"app id to filter by"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated			true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.ComponentBuild
// @Router					/v1/builds [GET]
func (s *service) GetComponentBuilds(ctx *gin.Context) (any, error) {
	cmpID := ctx.Query("component_id")
	appID := ctx.Query("app_id")
	if cmpID == "" && appID == "" {
		return nil, stderr.ErrUser{Err: fmt.Errorf("component id or app id must be passed in"), Description: "component id or app id must be passed in"}
	}

	return s.getComponentBuilds(ctx, cmpID, appID)
}

func (s *service) getComponentBuilds(ctx *gin.Context, cmpID, appID string) ([]app.ComponentBuild, error) {
	access, err := require.ReadAccessFromContext(ctx.Request.Context())
	if err != nil {
		return nil, err
	}
	query, err := access.Query()
	if err != nil {
		return nil, err
	}
	configs := access.DB().WithContext(ctx.Request.Context()).Model(&app.ComponentConfigConnection{}).
		Scopes(scopes.WithDisableViews).Select("id").Where(&app.ComponentConfigConnection{OrgID: access.OrgID()})
	if cmpID != "" {
		configs = configs.Where(&app.ComponentConfigConnection{ComponentID: cmpID})
	}
	if appID != "" {
		configs = configs.Where("component_id IN (?)", access.DB().Model(&app.Component{}).Select("id").Where(&app.Component{AppID: appID, OrgID: access.OrgID()}))
	}

	blds := []app.ComponentBuild{}

	res := query.WithContext(ctx.Request.Context()).
		Where("component_builds.component_config_connection_id IN (?)", configs).
		Scopes(scopes.WithOffsetPagination).
		Preload("CreatedBy").
		Preload("ComponentConfigConnection").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig").
		Preload("VCSConnectionCommit").
		Preload("AppBranchRun.VCSConnectionCommit").
		Preload("AppBranchRun.AppBranch").
		Preload("AppBranchRun.Preview").
		Preload("ComponentConfigConnection.Component").
		Order("component_builds.created_at DESC, component_builds.id DESC").
		Find(&blds)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get builds: %w", res.Error)
	}

	blds, err = db.HandlePaginatedResponse(ctx, blds)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return blds, nil
}
