package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm"
)

// @ID GetComponentLatestConfig
// @Summary	get latest config for a component
// @Description.markdown	get_component_latest_config.md
// @Param			component_id	path	string	true	"component ID"
// @Tags			components
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Accept			json
// @Produce		json
// @Success		200	{object}	app.ComponentConfigConnection
// @Router			/v1/components/{component_id}/configs/latest [GET]
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

		// preload all external image configs
		Preload("ComponentConfigs.ExternalImageComponentConfig").

		// preload all job configs
		Preload("ComponentConfigs.JobComponentConfig").

		// get config by build ID
		First(&cmp, "id = ?", cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get most recent component config: %w", res.Error)
	}

	if len(cmp.ComponentConfigs) < 1 {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("no component config found for component"),
			Description: "please make sure at least one component config has been created",
		}
	}

	return &cmp.ComponentConfigs[0], nil
}
