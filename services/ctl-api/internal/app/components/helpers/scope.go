package helpers

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
)

// component config connections
func PreloadLatestConfig(db *gorm.DB) *gorm.DB {
	return db.
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.
				Table(views.DefaultViewName(db,
					&app.ComponentConfigConnection{}, 1)).
				Order("created_at DESC").Limit(1)
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
		Preload("ComponentConfigs.JobComponentConfig")
}

// component config connections
func PreloadComponentBuildConfig(db *gorm.DB) *gorm.DB {
	return db.
		// preload all terraform configs
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig").
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all helm configs
		Preload("ComponentConfigConnection.HelmComponentConfig").
		Preload("ComponentConfigConnection.HelmComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.HelmComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all docker configs
		Preload("ComponentConfigConnection.DockerBuildComponentConfig").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all external image configs
		Preload("ComponentConfigConnection.ExternalImageComponentConfig").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.AWSECRImageConfig")
}
