package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetAppComponentConfig
// @Summary					get a config for a component
// @Description.markdown	get_component_config.md
// @Param					app_id						path	string	true	"app ID"
// @Param					component_id				path	string	true	"component ID"
// @Param					config_id					path	string	true	"config ID"
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
// @Success				200	{object}		app.ComponentConfigConnection
// @Router					/v1/apps/{app_id}/components/{component_id}/configs/{config_id} [GET]
func (s *service) GetAppComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")
	cfgID := ctx.Param("config_id")

	cfg, err := s.getComponentConfig(ctx, cmpID, cfgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component configs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

// @ID						GetComponentConfig
// @Summary					get a config for a component
// @Description.markdown	get_component_config.md
// @Param					component_id				path	string	true	"component ID"
// @Param					config_id					path	string	true	"config ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}		app.ComponentConfigConnection
// @Router					/v1/components/{component_id}/configs/{config_id} [GET]
func (s *service) GetComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")
	cfgID := ctx.Param("config_id")

	cfg, err := s.getComponentConfig(ctx, cmpID, cfgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component configs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) getComponentConfig(ctx *gin.Context, cmpID, cfgID string) (*app.ComponentConfigConnection, error) {
	cfg := app.ComponentConfigConnection{}
	res := s.db.WithContext(ctx).
		// preload all terraform configs
		Preload("TerraformModuleComponentConfig").
		Preload("TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
		Preload("TerraformModuleComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all helm configs
		Preload("HelmComponentConfig").
		Preload("HelmComponentConfig.PublicGitVCSConfig").
		Preload("HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("HelmComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all docker configs
		Preload("DockerBuildComponentConfig").
		Preload("DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("DockerBuildComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all external image configs
		Preload("ExternalImageComponentConfig").
		Preload("ExternalImageComponentConfig.AWSECRImageConfig").

		// preload all job configs
		Preload("JobComponentConfig").

		// preload all job configs
		Preload("KubernetesManifestComponentConfig").
		First(&cfg, "id = ? and component_id = ?", cfgID, cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component config connection: %w", res.Error)
	}

	return &cfg, nil
}
