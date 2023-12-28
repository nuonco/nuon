package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetComponentConfigs
// @Summary	get all configs for a component
// @Description.markdown	get_component_configs.md
// @Param			component_id	path	string	true	"component ID"
// @Tags			components
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.ComponentConfigConnection
// @Router			/v1/components/{component_id}/configs [GET]
func (s *service) GetComponentConfigs(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	component, err := s.getComponentConfigs(ctx, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component configs: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, component)
}

func (s *service) getComponentConfigs(ctx context.Context, cmpID string) ([]app.ComponentConfigConnection, error) {
	var cfgs []app.ComponentConfigConnection
	res := s.db.Model(&app.ComponentConfigConnection{
		ComponentID: cmpID,
	}).
		// preload all terraform configs
		Preload("TerraformModuleComponentConfig").
		Preload("TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("TerraformModuleComponentConfig.ConnectedGithubVCSConfig").

		// preload all helm configs
		Preload("HelmComponentConfig").
		Preload("HelmComponentConfig.PublicGitVCSConfig").
		Preload("HelmComponentConfig.ConnectedGithubVCSConfig").

		// preload all docker configs
		Preload("DockerBuildComponentConfig").
		Preload("DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("DockerBuildComponentConfig.ConnectedGithubVCSConfig").

		// preload all external image configs
		Preload("ExternalImageComponentConfig").

		// preload all job configs
		Preload("JobComponentConfig").

		// order by created at
		Order("created_at DESC").

		// find all configs
		Find(&cfgs, "component_id = ?", cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to load component configs: %w", res.Error)
	}

	return cfgs, nil
}
