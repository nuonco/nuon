package helpers

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

func PreloadLatestConfig(db *gorm.DB) *gorm.DB {
	return db.
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.
				Table(views.CurrentViewName(db,
					&app.ComponentConfigConnection{})).
				Order("created_at DESC").Limit(1)
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
		Preload("ComponentConfigs.KubernetesManifestComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.KubernetesManifestComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.PulumiComponentConfig").
		Preload("ComponentConfigs.PulumiComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.PulumiComponentConfig.ConnectedGithubVCSConfig")
}

func PreloadComponentBuildConfig(db *gorm.DB) *gorm.DB {
	return db.
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig").
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("ComponentConfigConnection.HelmComponentConfig").
		Preload("ComponentConfigConnection.HelmComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.HelmComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.AWSECRImageConfig").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.GCPGARImageConfig").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.AzureACRImageConfig").
		Preload("ComponentConfigConnection.KubernetesManifestComponentConfig").
		Preload("ComponentConfigConnection.KubernetesManifestComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.KubernetesManifestComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.KubernetesManifestComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("ComponentConfigConnection.PulumiComponentConfig").
		Preload("ComponentConfigConnection.PulumiComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.PulumiComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.PulumiComponentConfig.ConnectedGithubVCSConfig.VCSConnection")
}
