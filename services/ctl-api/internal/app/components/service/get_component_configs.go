package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
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
func (s *service) GetAppComponentConfigs(ctx *gin.Context) (any, error) {
	return s.GetComponentConfigs(ctx)
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
func (s *service) GetComponentConfigs(ctx *gin.Context) (any, error) {
	cmpID := ctx.Param("component_id")

	return s.getComponentConfigs(ctx, cmpID)
}

func (s *service) getComponentConfigs(ctx *gin.Context, cmpID string) ([]app.ComponentConfigConnection, error) {
	access, err := require.ReadAccessFromContext(ctx.Request.Context())
	if err != nil {
		return nil, err
	}

	cmp := app.Component{}
	res := access.DB().WithContext(ctx.Request.Context()).
		// preload configs
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.
				Where(&app.ComponentConfigConnection{OrgID: access.OrgID()}).
				Scopes(scopes.WithOffsetPagination).
				Order(views.TableOrViewName(db, &app.ComponentConfigConnection{}, ".created_at DESC")).
				Order(views.TableOrViewName(db, &app.ComponentConfigConnection{}, ".id DESC"))
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
		Preload("ComponentConfigs.ExternalImageComponentConfig.GCPGARImageConfig").
		Preload("ComponentConfigs.ExternalImageComponentConfig.AzureACRImageConfig").

		// preload all job configs
		Preload("ComponentConfigs.JobComponentConfig").

		// preload all kubernetes configs
		Preload("ComponentConfigs.KubernetesManifestComponentConfig").

		// preload all pulumi configs
		Preload("ComponentConfigs.PulumiComponentConfig").
		Preload("ComponentConfigs.PulumiComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.PulumiComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.PulumiComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Where(&app.Component{ID: cmpID, OrgID: access.OrgID()}).First(&cmp)
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
