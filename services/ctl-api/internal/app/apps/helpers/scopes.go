package helpers

import (
	"gorm.io/gorm"
)

// secrets config
func PreloadAppSecretsConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("SecretsConfig").
		Preload("SecretsConfig.Secrets")
}

// break glass config
func PreloadAppBreakGlassConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("BreakGlassConfig").
		Preload("BreakGlassConfig.Roles").
		Preload("BreakGlassConfig.Roles.Policies")
}

// cloudformation stack config
func PreloadAppConfigStackConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("StackConfig")
}

// permissions config
func PreloadAppConfigPermissionsConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("PermissionsConfig").
		Preload("PermissionsConfig.Roles").
		Preload("PermissionsConfig.Roles.Policies")
}

// policies config
func PreloadAppConfigPolicyConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("PoliciesConfig").
		Preload("PoliciesConfig.Policies")
}

// input config
func PreloadAppConfigInputConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("InputConfig").
		Preload("InputConfig.AppInputGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_groups.index ASC")
		}).
		Preload("InputConfig.AppInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_inputs.index ASC")
		})
}

// sandbox config
func PreloadAppConfigSandboxConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("SandboxConfig").
		Preload("SandboxConfig.PublicGitVCSConfig").
		Preload("SandboxConfig.ConnectedGithubVCSConfig")
}

// runner config
func PreloadAppConfigRunnerConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("RunnerConfig")
}

// component config connections
func PreloadAppConfigComponentConfigConnections(db *gorm.DB) *gorm.DB {
	return db.
		// preload the component this belongs too
		Preload("ComponentConfigConnections.Component").

		// preload all terraform configs
		Preload("ComponentConfigConnections.TerraformModuleComponentConfig").
		Preload("ComponentConfigConnections.TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnections.TerraformModuleComponentConfig.ConnectedGithubVCSConfig").

		// preload all helm configs
		Preload("ComponentConfigConnections.HelmComponentConfig").
		Preload("ComponentConfigConnections.HelmComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnections.HelmComponentConfig.ConnectedGithubVCSConfig").

		// preload all docker configs
		Preload("ComponentConfigConnections.DockerBuildComponentConfig").
		Preload("ComponentConfigConnections.DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnections.DockerBuildComponentConfig.ConnectedGithubVCSConfig").

		// preload all external image configs
		Preload("ComponentConfigConnections.ExternalImageComponentConfig").

		// preload all job configs
		Preload("ComponentConfigConnections.JobComponentConfig")
}

// component config connections
func PreloadComponentConfigConnections(db *gorm.DB) *gorm.DB {
	return db.
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
