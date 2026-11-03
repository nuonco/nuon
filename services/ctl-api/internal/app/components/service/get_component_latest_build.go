package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAppComponentLatestBuild
// @Summary				get latest build for a component
// @Description.markdown	get_component_latest_build.md
// @Param					app_id			path	string	true	"app ID"
// @Param					component_id	path	string	true	"component ID"
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
// @Success				200	{object}	app.ComponentBuild
// @Router					/v1/apps/{app_id}/components/{component_id}/builds/latest [GET]
func (s *service) GetAppComponentLatestBuild(ctx *gin.Context) (any, error) {
	return s.GetComponentLatestBuild(ctx)
}

// @ID						GetComponentLatestBuild
// @Summary				get latest build for a component
// @Description.markdown	get_component_latest_build.md
// @Param					component_id	path	string	true	"component ID"
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
// @Success				200	{object}	app.ComponentBuild
// @Router					/v1/components/{component_id}/builds/latest [GET]
func (s *service) GetComponentLatestBuild(ctx *gin.Context) (any, error) {
	cmpID := ctx.Param("component_id")

	return s.getComponentLatestBuild(ctx, cmpID)
}

func (s *service) getComponentLatestBuild(ctx *gin.Context, cmpID string) (*app.ComponentBuild, error) {
	access, err := require.ReadAccessFromContext(ctx.Request.Context())
	if err != nil {
		return nil, err
	}
	query, err := access.Query()
	if err != nil {
		return nil, err
	}

	configs := access.DB().Model(&app.ComponentConfigConnection{}).
		Scopes(scopes.WithOverrideTable(app.LatestComponentConfigConnectionsViewName)).
		Select("id").Where(&app.ComponentConfigConnection{ComponentID: cmpID, OrgID: access.OrgID()})
	var build app.ComponentBuild
	res := withoutPreviewBuilds(query.WithContext(ctx.Request.Context())).
		Where("component_builds.component_config_connection_id IN (?)", configs).
		Preload("ComponentConfigConnection.Component").
		Preload("VCSConnectionCommit").
		Preload("AppBranchRun.VCSConnectionCommit").
		Preload("AppBranchRun.AppBranch").
		Preload("AppBranchRun.Preview").
		Order("component_builds.created_at DESC, component_builds.id DESC").
		First(&build)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get latest component build: %w", res.Error)
	}
	return &build, nil
}

func withoutPreviewBuilds(db *gorm.DB) *gorm.DB {
	return db.
		Joins("LEFT JOIN app_branch_runs latest_build_runs ON latest_build_runs.id = component_builds.app_branch_run_id").
		Joins("LEFT JOIN app_branch_run_previews latest_build_previews ON latest_build_previews.app_branch_run_id = latest_build_runs.id AND latest_build_previews.deleted_at = 0").
		Where(clause.Or(
			clause.Eq{
				Column: clause.Column{Table: "component_builds", Name: "app_branch_run_id"},
				Value:  nil,
			},
			clause.And(
				clause.Eq{
					Column: clause.Column{Table: "latest_build_previews", Name: "id"},
					Value:  nil,
				},
				clause.Neq{
					Column: clause.Column{Table: "latest_build_runs", Name: "run_type"},
					Value:  app.AppBranchRunTypeGitPreview,
				},
				clause.Eq{
					Column: clause.Column{Table: "latest_build_runs", Name: "plan_only"},
					Value:  false,
				},
			),
		))
}
