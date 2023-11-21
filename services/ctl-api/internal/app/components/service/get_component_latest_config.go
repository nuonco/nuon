package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

//	@BasePath	/v1/components
//
// Get latest config for a component
//
//	@Summary	get latest config for a component
//	@Schemes
//	@Description	get latest config for a component
//	@Param			component_id	path	string	true	"component ID"
//	@Tags			components
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	app.ComponentConfigConnection
//	@Router			/v1/components/{component_id}/configs/latest [GET]
func (s *service) GetComponentLatestConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	cfg, err := s.getComponentLatestConfig(ctx, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component configs: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) getComponentLatestConfig(ctx *gin.Context, cmpID string) (*app.ComponentConfigConnection, error) {
	cmp := app.Component{}

	res := s.db.WithContext(ctx).Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
		return db.Order("component_config_connections.created_at DESC").Limit(1)
	}).
		// preload all terraform configs
		Preload("ComponentConfigs.TerraformModuleComponentConfig").
		Preload("ComponentConfigs.TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.TerraformModuleComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all helm configs
		Preload("ComponentConfigs.HelmComponentConfig").
		Preload("ComponentConfigs.HelmComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.HelmComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all docker configs
		Preload("ComponentConfigs.DockerBuildComponentConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("ComponentConfigs.DockerBuildComponentConfig.BasicDeployConfig").

		// preload all external image configs
		Preload("ComponentConfigs.ExternalImageComponentConfig").
		Preload("ComponentConfigs.ExternalImageComponentConfig.BasicDeployConfig").

		// preload all job configs
		Preload("ComponentConfigs.JobComponentConfig").

		// get config by build ID
		First(&cmp, "id = ?", cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get most recent component config: %w", res.Error)
	}

	if len(cmp.ComponentConfigs) < 1 {
		return nil, fmt.Errorf("no component config found for component")
	}

	return &cmp.ComponentConfigs[0], nil
}
