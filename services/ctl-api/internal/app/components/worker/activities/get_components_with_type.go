package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"gorm.io/gorm"
)

type GetComponentsWithType struct {
	IDs []string
}

// @temporal-gen activity
func (a *Activities) GetComponentsWithType(ctx context.Context, req GetComponentsWithType) ([]app.Component, error) {
	comps := make([]app.Component, 0)
	res := a.db.WithContext(ctx).Model(&app.Component{}).Where("ID IN ?", req.IDs).
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.
				Table(views.DefaultViewName(db,
					&app.ComponentConfigConnection{}, 1)).
				Order("created_at DESC")
		}).

		// preload all terraform configs
		Preload("ComponentConfigs.TerraformModuleComponentConfig").
		Preload("ComponentConfigs.TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.TerraformModuleComponentConfig.ConnectedGithubVCSConfig").

		// preload all helm configs
		Preload("ComponentConfigs.HelmComponentConfig").
		Preload("ComponentConfigs.HelmComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.HelmComponentConfig.ConnectedGithubVCSConfig").

		// preload all docker configs
		Preload("ComponentConfigs.DockerBuildComponentConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig.ConnectedGithubVCSConfig").

		// preload all external image configs
		Preload("ComponentConfigs.ExternalImageComponentConfig").

		// preload all job configs
		Preload("ComponentConfigs.JobComponentConfig").
		Find(&comps)
	if res.Error != nil {
		return nil, res.Error
	}

	return comps, nil
}
