package helpers

import (
	"context"
	"encoding/json"
	"fmt"

	pkgconfig "github.com/nuonco/nuon/pkg/config"
	configsync "github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"gorm.io/gorm"
)

func (h *Helpers) GetCustomerManagedAppConfig(ctx context.Context, orgID, appID, appConfigID string) (*app.AppConfig, error) {
	var appConfig app.AppConfig
	res := h.db.WithContext(ctx).
		Where(app.AppConfig{ID: appConfigID, OrgID: orgID, AppID: appID}).
		Scopes(
			PreloadAppConfigSandboxConfig,
		).
		Preload("ActionWorkflowConfigs").
		Preload("ActionWorkflowConfigs.ActionWorkflow").
		Preload("ActionWorkflowConfigs.Triggers").
		Preload("ActionWorkflowConfigs.Steps").
		First(&appConfig)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get exact app config for customer-managed bundle: %w", res.Error)
	}

	var state configsync.State
	if err := json.Unmarshal([]byte(appConfig.State), &state); err != nil {
		return nil, fmt.Errorf("decode exact app config state: %w", err)
	}

	connections, err := h.loadComponentConfigConnections(ctx, orgID, state.Components)
	if err != nil {
		return nil, err
	}
	appConfig.ComponentConfigConnections = connections
	appConfig.ActionIDs = appConfig.ActionIDs[:0]
	for _, action := range state.Actions {
		appConfig.ActionIDs = append(appConfig.ActionIDs, action.ID)
	}
	intermediate, err := appConfig.IntermediateConfig.Get(blobstore.WithBlobService(ctx, h.blobSvc))
	if err != nil {
		return nil, fmt.Errorf("load customer-managed runtime config: %w", err)
	}
	var source pkgconfig.AppConfig
	if err := json.Unmarshal([]byte(intermediate), &source); err != nil {
		return nil, fmt.Errorf("decode customer-managed runtime config: %w", err)
	}
	if source.CustomerManaged != nil {
		runtime := &app.AppReleaseRuntime{
			RunnerImageURL: source.CustomerManaged.RunnerImageURL,
			RunnerImageTag: source.CustomerManaged.RunnerImageTag,
			Platforms:      make(map[string]app.AppReleasePlatformRuntime, len(source.CustomerManaged.Platforms)),
		}
		for _, platform := range source.CustomerManaged.Platforms {
			if _, exists := runtime.Platforms[platform.Target]; exists {
				return nil, fmt.Errorf("customer-managed runtime has duplicate platform %q", platform.Target)
			}
			runtime.Platforms[platform.Target] = app.AppReleasePlatformRuntime{
				PortalBinaryURL: platform.PortalBinaryURL,
				RunnerBinaryURL: platform.RunnerBinaryURL,
			}
		}
		appConfig.CustomerManagedRuntime = runtime
	}
	return &appConfig, nil
}

func (h *Helpers) loadComponentConfigConnections(ctx context.Context, orgID string, components []configsync.ComponentState) ([]app.ComponentConfigConnection, error) {
	configIDs := make([]string, 0, len(components))
	for _, c := range components {
		if c.ConfigID != "" {
			configIDs = append(configIDs, c.ConfigID)
		}
	}
	if len(configIDs) == 0 {
		return nil, nil
	}

	var direct []app.ComponentConfigConnection
	if err := h.db.WithContext(ctx).
		Scopes(PreloadComponentConfigConnection).
		Where("org_id = ? AND id IN ?", orgID, configIDs).
		Find(&direct).Error; err != nil {
		return nil, fmt.Errorf("load component config connections: %w", err)
	}

	byConfigID := make(map[string]app.ComponentConfigConnection, len(components))
	for _, c := range direct {
		byConfigID[c.ID] = c
	}

	var missing []configsync.ComponentState
	for _, c := range components {
		if c.ConfigID == "" {
			continue
		}
		if _, ok := byConfigID[c.ConfigID]; !ok {
			missing = append(missing, c)
		}
	}

	for _, c := range missing {
		connectionID, err := h.connectionIDFromTypedComponentConfig(ctx, orgID, c)
		if err != nil {
			return nil, fmt.Errorf("load exact component config %s for customer-managed bundle: %w", c.ConfigID, err)
		}
		var connection app.ComponentConfigConnection
		res := h.db.WithContext(ctx).
			Scopes(PreloadComponentConfigConnection).
			Where(app.ComponentConfigConnection{ID: connectionID, OrgID: orgID, ComponentID: c.ID}).
			First(&connection)
		if res.Error != nil {
			return nil, fmt.Errorf("load exact component config %s for customer-managed bundle: %w", c.ConfigID, res.Error)
		}
		byConfigID[c.ConfigID] = connection
	}

	ordered := make([]app.ComponentConfigConnection, 0, len(components))
	for _, c := range components {
		if c.ConfigID == "" {
			continue
		}
		if conn, ok := byConfigID[c.ConfigID]; ok {
			ordered = append(ordered, conn)
		}
	}
	return ordered, nil
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

func PreloadComponentConfigConnection(db *gorm.DB) *gorm.DB {
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
