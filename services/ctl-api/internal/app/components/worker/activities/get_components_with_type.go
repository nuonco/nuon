package activities

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

type GetComponentsWithType struct {
	IDs []string `validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 300s
// @max-retries 1
func (a *Activities) GetComponentsWithType(ctx context.Context, req GetComponentsWithType) ([]app.Component, error) {
	comps := make([]app.Component, 0)

	res := a.db.WithContext(ctx).Model(&app.Component{}).Where("ID IN ?", req.IDs).
		Preload("ComponentConfigs").
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(a.db, &app.ComponentConfigConnection{}, ".created_at DESC")).Limit(1)
		}).
		Preload("ComponentConfigs.TerraformModuleComponentConfig").
		Preload("ComponentConfigs.TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.HelmComponentConfig").
		Preload("ComponentConfigs.HelmComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.ExternalImageComponentConfig").
		Preload("ComponentConfigs.JobComponentConfig").
		Preload("ComponentConfigs.KubernetesManifestComponentConfig").
		Preload("ComponentConfigs.PulumiComponentConfig").
		Preload("ComponentConfigs.PulumiComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.PulumiComponentConfig.ConnectedGithubVCSConfig").
		Find(&comps)
	if res.Error != nil {
		return nil, res.Error
	}

	return comps, nil
}
