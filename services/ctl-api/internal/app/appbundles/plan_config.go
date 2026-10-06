package appbundle

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	appbundle "github.com/nuonco/nuon/pkg/appbundle"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func exportAppConfigJSON(ctx context.Context, db *gorm.DB, appConfigID string) (json.RawMessage, error) {
	var raw string
	err := db.WithContext(ctx).Raw(`
SELECT (row_to_json(app_config)::jsonb || jsonb_build_object(
	'sandbox',
	(
		SELECT row_to_json(sandbox)::jsonb
		FROM app_sandbox_configs AS sandbox
		WHERE sandbox.app_config_id = app_config.id AND sandbox.deleted_at = 0
		ORDER BY sandbox.created_at DESC
		LIMIT 1
	)
))::text
FROM app_configs AS app_config
WHERE id = ? AND deleted_at = 0`, appConfigID).Scan(&raw).Error
	if err != nil {
		return nil, fmt.Errorf("serialize app config %s: %w", appConfigID, err)
	}
	if raw == "" {
		return nil, fmt.Errorf("app config %s not found", appConfigID)
	}
	var probe struct {
		Sandbox json.RawMessage `json:"sandbox"`
	}
	if err := json.Unmarshal([]byte(raw), &probe); err != nil {
		return nil, fmt.Errorf("parse app config %s: %w", appConfigID, err)
	}
	if len(probe.Sandbox) == 0 || string(probe.Sandbox) == "null" {
		return nil, fmt.Errorf("app config %s has no sandbox config; offline replay of sandbox jobs requires one", appConfigID)
	}

	connections, err := exportComponentConfigConnections(ctx, db, appConfigID)
	if err != nil {
		return nil, err
	}
	var merged map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &merged); err != nil {
		return nil, fmt.Errorf("parse app config %s: %w", appConfigID, err)
	}
	connectionsJSON, err := json.Marshal(connections)
	if err != nil {
		return nil, fmt.Errorf("serialize component config connections for %s: %w", appConfigID, err)
	}
	merged["component_config_connections"] = connectionsJSON
	out, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("serialize app config %s: %w", appConfigID, err)
	}
	return out, nil
}

func exportComponentConfigConnections(ctx context.Context, db *gorm.DB, appConfigID string) ([]app.ComponentConfigConnection, error) {
	var appCfg app.AppConfig
	err := db.WithContext(ctx).
		Where(app.AppConfig{ID: appConfigID}).
		Scopes(appshelpers.PreloadAppConfigComponentConfigConnections).
		First(&appCfg).Error
	if err != nil {
		return nil, fmt.Errorf("load component configs for app config %s: %w", appConfigID, err)
	}

	seen := map[string]bool{}
	for _, connection := range appCfg.ComponentConfigConnections {
		seen[connection.ComponentID] = true
	}
	var missing []string
	for _, componentID := range appCfg.ComponentIDs {
		if !seen[componentID] {
			missing = append(missing, componentID)
		}
	}
	connections := appCfg.ComponentConfigConnections
	if len(missing) > 0 {
		var latest []app.ComponentConfigConnection
		err = db.WithContext(ctx).
			Scopes(
				scopes.WithDisableViews,
				scopes.WithOverrideTable("component_config_connections_latest_configs_view"),
			).
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
			Preload("PulumiComponentConfig").
			Preload("PulumiComponentConfig.PublicGitVCSConfig").
			Preload("PulumiComponentConfig.ConnectedGithubVCSConfig").
			Where("component_id IN ?", missing).
			Find(&latest).Error
		if err != nil {
			return nil, fmt.Errorf("load latest component configs for app config %s: %w", appConfigID, err)
		}
		connections = append(connections, latest...)
	}
	if len(connections) != len(appCfg.ComponentIDs) {
		return nil, fmt.Errorf("app config %s: found %d component configs, expected %d", appConfigID, len(connections), len(appCfg.ComponentIDs))
	}
	return connections, nil
}

func componentSpecs(installComponents []app.InstallComponent, connections []app.ComponentConfigConnection) []appbundle.ComponentSpec {
	helmByComponent := map[string]*app.HelmComponentConfig{}
	for i := range connections {
		if cfg := connections[i].HelmComponentConfig; cfg != nil {
			helmByComponent[connections[i].ComponentID] = cfg
		}
	}
	specs := make([]appbundle.ComponentSpec, 0, len(installComponents))
	for _, ic := range installComponents {
		spec := appbundle.ComponentSpec{
			InstallComponentID: ic.ID,
			ComponentID:        ic.ComponentID,
			ComponentName:      ic.Component.Name,
			ComponentType:      string(ic.Component.Type),
		}
		if cfg, ok := helmByComponent[ic.ComponentID]; ok {
			spec.HelmReleaseName = cfg.ChartName
			spec.HelmNamespace = cfg.Namespace.ValueString()
		}
		specs = append(specs, spec)
	}
	return specs
}

func exportInputSpecs(ctx context.Context, db *gorm.DB, appConfigID string) ([]appbundle.InputSpec, error) {
	var config app.AppInputConfig
	err := db.WithContext(ctx).
		Preload("AppInputs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order(`app_inputs."index", app_inputs.created_at, app_inputs.id`)
		}).
		Where(app.AppInputConfig{AppConfigID: appConfigID}).
		First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("load inputs for app config %s: %w", appConfigID, err)
	}
	specs := make([]appbundle.InputSpec, 0, len(config.AppInputs))
	for _, in := range config.AppInputs {
		specs = append(specs, appbundle.InputSpec{
			Name:        in.Name,
			Type:        string(in.Type),
			Description: in.Description,
			Secret:      in.Sensitive,
			Required:    in.Required,
			Default:     in.Default,
		})
	}
	return specs, nil
}
