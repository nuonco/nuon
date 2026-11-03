package service

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAppComponentBuild
// @Summary				get a build for a component
// @Description.markdown	get_component_build.md
// @Param					app_id			path	string	true	"app ID"
// @Param					component_id	path	string	true	"component ID"
// @Param					build_id		path	string	true	"build ID"
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
// @Router					/v1/apps/{app_id}/components/{component_id}/builds/{build_id} [GET]
func (s *service) GetAppComponentBuild(ctx *gin.Context) (any, error) {
	return s.GetComponentBuild(ctx)
}

// @ID						GetComponentBuild
// @Summary				get a build for a component
// @Description.markdown	get_component_build.md
// @Param					component_id	path	string	true	"component ID"
// @Param					build_id		path	string	true	"build ID"
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
// @Router					/v1/components/{component_id}/builds/{build_id} [GET]
func (s *service) GetComponentBuild(ctx *gin.Context) (any, error) {
	cmpID := ctx.Param("component_id")
	bldID := ctx.Param("build_id")

	return s.getComponentBuild(ctx.Request.Context(), cmpID, bldID)
}

func (s *service) getComponentBuild(ctx context.Context, cmpID, bldID string) (*app.ComponentBuild, error) {
	access, err := require.ReadAccessFromContext(ctx)
	if err != nil {
		return nil, err
	}
	query, err := access.Query()
	if err != nil {
		return nil, err
	}
	if cmpID != "" {
		query = query.Where("component_builds.component_config_connection_id IN (?)", access.DB().Model(&app.ComponentConfigConnection{}).
			Scopes(scopes.WithDisableViews).Select("id").Where(&app.ComponentConfigConnection{ComponentID: cmpID, OrgID: access.OrgID()}))
	}

	var bld app.ComponentBuild

	// query the build in a way where it will _only_ be returned if it belongs to the component id in question
	res := query.WithContext(ctx).
		Preload("CreatedBy").
		Preload("VCSConnectionCommit").
		Preload("ComponentConfigConnection").
		Preload("ComponentConfigConnection.Component").
		Preload("LogStream").
		Preload("QueueSignal").
		Preload("AppBranchRun").
		Preload("AppBranchRun.Preview").
		Preload("AppBranchRun.VCSConnectionCommit").
		Preload("AppBranchRun.AppBranch").
		Where(&app.ComponentBuild{ID: bldID}).First(&bld)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component build: %w", res.Error)
	}
	if err := s.hydrateBuildRunnerJobs(ctx, &bld); err != nil {
		return nil, err
	}
	jobCompositeError, err := runnershelpers.GetLatestJobCompositeError(ctx, access.DB().Where(app.RunnerJob{
		OrgID: access.OrgID(), Group: app.RunnerJobGroupBuild, Operation: app.RunnerJobOperationTypeBuild,
	}), runnershelpers.GetLatestJobCompositeErrorRequest{
		OwnerID:   bld.ID,
		OwnerType: "component_builds",
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get component build composite error: %w", err)
	}
	if jobCompositeError != nil || bld.RunnerJob.ID != "" {
		bld.CompositeError = jobCompositeError
	}

	return &bld, nil
}
