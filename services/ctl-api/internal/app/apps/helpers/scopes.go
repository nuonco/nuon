package helpers

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

// why: ActiveAppConfigs restricts a query to the app's configs that finished syncing.
// A pending, syncing or errored config has no component_ids and no config records, so it must never
// stand in as the app's current config.
//
// Ordering must be chained on the query builder by the caller, NOT added here: an Order clause
// inside a scope is dropped at execution time, which silently turns "newest active config" into
// First()'s primary-key ordering (i.e. the config with the smallest ID).
func ActiveAppConfigs(appID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Where(app.AppConfig{AppID: appID}).
			Where(
				views.TableOrViewName(db, &app.AppConfig{}, ".status_v2 ->> 'status' = ?"),
				string(app.AppConfigStatusActive),
			)
	}
}

func PreloadAppSecretsConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("SecretsConfig").
		Preload("SecretsConfig.Secrets").
		Preload("SecretsConfig.Secrets.KubernetesSyncTargets")
}

func PreloadAppBreakGlassConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("BreakGlassConfig").
		Preload("BreakGlassConfig.Roles").
		Preload("BreakGlassConfig.Roles.Policies")
}

func PreloadAppOperationRoleConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("OperationRoleConfig").
		Preload("OperationRoleConfig.Rules")
}

func PreloadAppConfigStackConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("StackConfig")
}

func PreloadAppConfigPermissionsConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("PermissionsConfig").
		Preload("PermissionsConfig.Roles").
		Preload("PermissionsConfig.Roles.Policies").
		Preload("PermissionsConfig.NamedPolicies")
}

func PreloadAppConfigPolicyConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("PoliciesConfig").
		Preload("PoliciesConfig.Policies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC, id ASC")
		})
}

func PreloadAppConfigInputConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("InputConfig").
		Preload("InputConfig.AppInputGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_groups.index ASC")
		}).
		Preload("InputConfig.AppInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_inputs.index ASC")
		})
}

func PreloadAppConfigSandboxConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("SandboxConfig").
		Preload("SandboxConfig.PublicGitVCSConfig").
		Preload("SandboxConfig.ConnectedGithubVCSConfig").
		Preload("SandboxConfig.ConnectedGithubVCSConfig.VCSConnection")
}

func PreloadAppConfigKubernetesContextsConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("KubernetesContextsConfig").
		Preload("KubernetesContextsConfig.Contexts")
}

func PreloadAppConfigRunnerConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("RunnerConfig")
}

func PreloadAppActionWorkflowConfigs(db *gorm.DB) *gorm.DB {
	return db.
		Preload("ActionWorkflowConfigs").
		Preload("ActionWorkflowConfigs.Triggers")
}

func PreloadAppConfigComponentConfigConnections(db *gorm.DB) *gorm.DB {
	return db.
		Preload("ComponentConfigConnections.Component").
		Preload("ComponentConfigConnections.TerraformModuleComponentConfig").
		Preload("ComponentConfigConnections.TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnections.TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnections.HelmComponentConfig").
		Preload("ComponentConfigConnections.HelmComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnections.HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnections.DockerBuildComponentConfig").
		Preload("ComponentConfigConnections.DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnections.DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnections.ExternalImageComponentConfig").
		Preload("ComponentConfigConnections.JobComponentConfig").
		Preload("ComponentConfigConnections.KubernetesManifestComponentConfig").
		Preload("ComponentConfigConnections.PulumiComponentConfig").
		Preload("ComponentConfigConnections.PulumiComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnections.PulumiComponentConfig.ConnectedGithubVCSConfig")
}

func PreloadComponentConfigConnections(db *gorm.DB) *gorm.DB {
	return db.
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
		Preload("ComponentConfigs.PulumiComponentConfig.ConnectedGithubVCSConfig")
}
