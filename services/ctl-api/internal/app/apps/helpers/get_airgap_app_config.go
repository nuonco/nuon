package helpers

import (
	"context"
	"encoding/json"
	"fmt"

	configsync "github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (h *Helpers) GetAirgapAppConfig(ctx context.Context, orgID, appID, appConfigID string) (*app.AppConfig, error) {
	var appConfig app.AppConfig
	res := h.db.WithContext(ctx).
		Where(app.AppConfig{ID: appConfigID, OrgID: orgID, AppID: appID}).
		Scopes(
			PreloadAppConfigRunnerConfig,
			PreloadAppConfigSandboxConfig,
			PreloadAppConfigStackConfig,
			PreloadAppConfigPermissionsConfig,
			PreloadAppBreakGlassConfig,
			PreloadAppSecretsConfig,
			PreloadAppConfigInputConfig,
			PreloadAppActionWorkflowConfigs,
		).
		First(&appConfig)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get exact app config for air-gap bundle: %w", res.Error)
	}

	var state configsync.State
	if err := json.Unmarshal([]byte(appConfig.State), &state); err != nil {
		return nil, fmt.Errorf("decode exact app config state: %w", err)
	}
	connections := make([]app.ComponentConfigConnection, 0, len(state.Components))
	for _, component := range state.Components {
		if component.ConfigID == "" {
			continue
		}
		var connection app.ComponentConfigConnection
		res = h.db.WithContext(ctx).
			Scopes(preloadComponentConfigConnectionsForAirgap).
			Where(app.ComponentConfigConnection{ID: component.ConfigID, OrgID: orgID, ComponentID: component.ID}).
			First(&connection)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			connectionID, cerr := h.connectionIDFromTypedComponentConfig(ctx, orgID, component)
			if cerr != nil {
				return nil, fmt.Errorf("load exact component config %s for air-gap bundle: %w", component.ConfigID, cerr)
			}
			res = h.db.WithContext(ctx).
				Scopes(preloadComponentConfigConnectionsForAirgap).
				Where(app.ComponentConfigConnection{ID: connectionID, OrgID: orgID, ComponentID: component.ID}).
				First(&connection)
		}
		if res.Error != nil {
			return nil, fmt.Errorf("load exact component config %s for air-gap bundle: %w", component.ConfigID, res.Error)
		}
		connections = append(connections, connection)
	}
	appConfig.ComponentConfigConnections = connections
	appConfig.ActionIDs = appConfig.ActionIDs[:0]
	for _, action := range state.Actions {
		appConfig.ActionIDs = append(appConfig.ActionIDs, action.ID)
	}
	return &appConfig, nil
}

func (h *Helpers) connectionIDFromTypedComponentConfig(ctx context.Context, orgID string, component configsync.ComponentState) (string, error) {
	db := h.db.WithContext(ctx)
	switch component.Type {
	case models.AppComponentTypeTerraformModule:
		var cfg app.TerraformModuleComponentConfig
		if res := db.Where(app.TerraformModuleComponentConfig{ID: component.ConfigID, OrgID: orgID}).First(&cfg); res.Error != nil {
			return "", res.Error
		}
		return cfg.ComponentConfigConnectionID, nil
	case models.AppComponentTypeHelmChart:
		var cfg app.HelmComponentConfig
		if res := db.Where(app.HelmComponentConfig{ID: component.ConfigID, OrgID: orgID}).First(&cfg); res.Error != nil {
			return "", res.Error
		}
		return cfg.ComponentConfigConnectionID, nil
	case models.AppComponentTypeDockerBuild:
		var cfg app.DockerBuildComponentConfig
		if res := db.Where(app.DockerBuildComponentConfig{ID: component.ConfigID, OrgID: orgID}).First(&cfg); res.Error != nil {
			return "", res.Error
		}
		return cfg.ComponentConfigConnectionID, nil
	case models.AppComponentTypeExternalImage:
		var cfg app.ExternalImageComponentConfig
		if res := db.Where(app.ExternalImageComponentConfig{ID: component.ConfigID, OrgID: orgID}).First(&cfg); res.Error != nil {
			return "", res.Error
		}
		return cfg.ComponentConfigConnectionID, nil
	case models.AppComponentTypeJob:
		var cfg app.JobComponentConfig
		if res := db.Where(app.JobComponentConfig{ID: component.ConfigID, OrgID: orgID}).First(&cfg); res.Error != nil {
			return "", res.Error
		}
		return cfg.ComponentConfigConnectionID, nil
	case models.AppComponentTypeKubernetesManifest:
		var cfg app.KubernetesManifestComponentConfig
		if res := db.Where(app.KubernetesManifestComponentConfig{ID: component.ConfigID, OrgID: orgID}).First(&cfg); res.Error != nil {
			return "", res.Error
		}
		return cfg.ComponentConfigConnectionID, nil
	case models.AppComponentTypePulumi:
		var cfg app.PulumiComponentConfig
		if res := db.Where(app.PulumiComponentConfig{ID: component.ConfigID, OrgID: orgID}).First(&cfg); res.Error != nil {
			return "", res.Error
		}
		return cfg.ComponentConfigConnectionID, nil
	}
	return "", fmt.Errorf("unsupported component type %q", component.Type)
}

func preloadComponentConfigConnectionsForAirgap(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Component").
		Preload("TerraformModuleComponentConfig").
		Preload("TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
		Preload("HelmComponentConfig").
		Preload("HelmComponentConfig.PublicGitVCSConfig").
		Preload("HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("DockerBuildComponentConfig").
		Preload("DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("ExternalImageComponentConfig").
		Preload("JobComponentConfig").
		Preload("KubernetesManifestComponentConfig").
		Preload("KubernetesManifestComponentConfig.PublicGitVCSConfig").
		Preload("KubernetesManifestComponentConfig.ConnectedGithubVCSConfig").
		Preload("PulumiComponentConfig").
		Preload("PulumiComponentConfig.PublicGitVCSConfig").
		Preload("PulumiComponentConfig.ConnectedGithubVCSConfig")
}
