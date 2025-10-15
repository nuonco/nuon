package helpers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

func (h *Helpers) GetFullAppConfig(ctx context.Context, appConfigID string) (*app.AppConfig, error) {
	appCfg := app.AppConfig{}
	res := h.db.WithContext(ctx).
		Where(app.AppConfig{
			ID: appConfigID,
		}).
		Scopes(
			// permissions
			PreloadAppSecretsConfig,
			PreloadAppBreakGlassConfig,
			PreloadAppConfigPermissionsConfig,
			PreloadAppConfigPolicyConfig,

			// basics
			PreloadAppConfigRunnerConfig,
			PreloadAppConfigSandboxConfig,
			PreloadAppConfigInputConfig,
			PreloadAppConfigStackConfig,

			// components
			PreloadAppConfigComponentConfigConnections,

			// actions
			PreloadAppActionWorkflowConfigs,
		).
		First(&appCfg)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app config")
	}
	if appCfg.Status != app.AppConfigStatusActive {
		return nil, fmt.Errorf("app config %s is in an error state", appCfg.ID)
	}

	missingComponentIds := make([]string, 0)
	componentsByID := make(map[string]bool)
	for _, componentCfg := range appCfg.ComponentConfigConnections {
		if _, ok := componentsByID[componentCfg.ComponentID]; !ok {
			componentsByID[componentCfg.ComponentID] = true
		}
	}

	for _, componentID := range appCfg.ComponentIDs {
		if _, ok := componentsByID[componentID]; !ok {
			missingComponentIds = append(missingComponentIds, componentID)
		}
	}

	if len(missingComponentIds) > 0 {
		missingComponents := []app.ComponentConfigConnection{}
		res = h.db.WithContext(ctx).
			Scopes(
				scopes.WithDisableViews,
				scopes.WithOverrideTable("component_config_connections_latest_configs_view"),
			).
			// preload the component this belongs too
			Preload("Component").

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

			// preload all kubernetes config
			Preload("KubernetesManifestComponentConfig").
			Where("component_id IN ?", missingComponentIds).
			Find(&missingComponents)
		if res.Error != nil {
			return nil, errors.Wrap(res.Error, "unable to get missing component configs")
		}
		if len(missingComponents) > 0 {
			appCfg.ComponentConfigConnections = append(appCfg.ComponentConfigConnections, missingComponents...)
		}
	}

	if len(appCfg.ComponentConfigConnections) != len(appCfg.ComponentIDs) {
		return nil, errors.New("an app config references a component-id which has a config that could not be found")
	}

	return &appCfg, nil
}
