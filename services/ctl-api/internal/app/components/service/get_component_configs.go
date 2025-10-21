package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
	"gorm.io/gorm"
)

// @ID						GetAppComponentConfigs
// @Summary				get all configs for a component
// @Description.markdown	get_component_configs.md
// @Param					app_id						path	string	true	"app ID"
// @Param					component_id				path	string	true	"component ID"
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
// @Success				200	{array}		app.ComponentConfigConnection
// @Router					/v1/apps/{app_id}/components/{component_id}/configs [GET]
func (s *service) GetAppComponentConfigs(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	cfgs, err := s.getComponentConfigs(ctx, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component configs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, cfgs)
}

// @ID						GetComponentConfigs
// @Summary				get all configs for a component
// @Description.markdown	get_component_configs.md
// @Param					component_id				path	string	true	"component ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
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
// @Success				200	{array}		app.ComponentConfigConnection
// @Router					/v1/components/{component_id}/configs [GET]
func (s *service) GetComponentConfigs(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	cfgs, err := s.getComponentConfigs(ctx, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component configs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, cfgs)
}

func (s *service) getComponentConfigs(ctx *gin.Context, cmpID string) ([]app.ComponentConfigConnection, error) {
	cmp := app.Component{}
	res := s.db.WithContext(ctx).
		// preload configs
		Preload("ComponentConfigs").
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOffsetPagination).
				Order(views.TableOrViewName(s.db, &app.ComponentConfigConnection{}, ".created_at DESC"))
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
		Preload("ComponentConfigs.ExternalImageComponentConfig.AWSECRImageConfig").

		// preload all job configs
		Preload("ComponentConfigs.JobComponentConfig").

		// preload all kubernetes configs
		Preload("ComponentConfigs.KubernetesManifestComponentConfig").
		First(&cmp, "id = ?", cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	cfgs, err := db.HandlePaginatedResponse(ctx, cmp.ComponentConfigs)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	cmp.ComponentConfigs = cfgs

	return cmp.ComponentConfigs, nil
}
