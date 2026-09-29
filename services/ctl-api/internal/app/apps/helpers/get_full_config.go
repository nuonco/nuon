package helpers

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/Masterminds/semver/v3"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

func (h *Helpers) GetFullAppConfig(ctx context.Context, appConfigID string, skipAdditionalChecks bool) (*app.AppConfig, error) {
	appCfg := app.AppConfig{}
	res := h.db.WithContext(ctx).
		Where(app.AppConfig{
			ID: appConfigID,
		}).
		Scopes(
			PreloadAppSecretsConfig,
			PreloadAppBreakGlassConfig,
			PreloadAppConfigPermissionsConfig,
			PreloadAppConfigPolicyConfig,
			PreloadAppOperationRoleConfig,

			PreloadAppConfigRunnerConfig,
			PreloadAppConfigSandboxConfig,
			PreloadAppConfigInputConfig,
			PreloadAppConfigStackConfig,
			PreloadAppConfigKubernetesContextsConfig,

			PreloadAppConfigComponentConfigConnections,

			PreloadAppActionWorkflowConfigs,
		).
		First(&appCfg)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app config")
	}
	if appCfg.Status == app.AppConfigStatusError {
		return nil, stderr.ErrUser{
			Description: fmt.Sprintf("app config %s is in an error state", appCfg.ID),
			Err:         fmt.Errorf("app config %s is in an error state", appCfg.ID),
		}
	}

	if !skipAdditionalChecks {
		versionAllowed, err := h.CliVerisionAllowed(ctx, appCfg.CLIVersion)
		if err != nil {
			return nil, errors.Wrap(err, "unable to check cli version")
		}

		if !versionAllowed {
			return nil, fmt.Errorf("resync config with a newer cli version: app_config_id %s", appCfg.ID)
		}
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
		var boundedCfgIDs []string
		res = h.db.WithContext(ctx).Raw(fmt.Sprintf(`
			SELECT DISTINCT ON (component_id) id
			FROM %s
			WHERE component_id IN ? AND app_config_version <= ?
			ORDER BY component_id, app_config_version DESC`,
			views.CurrentViewName(h.db, &app.ComponentConfigConnection{})),
			missingComponentIds, appCfg.Version,
		).Scan(&boundedCfgIDs)
		if res.Error != nil {
			return nil, errors.Wrap(res.Error, "unable to resolve component configs at config version")
		}

		res = h.db.WithContext(ctx).
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
			Where("id IN ?", boundedCfgIDs).
			Find(&missingComponents)
		if res.Error != nil {
			return nil, errors.Wrap(res.Error, "unable to get missing component configs")
		}
		if len(missingComponents) > 0 {
			appCfg.ComponentConfigConnections = append(appCfg.ComponentConfigConnections, missingComponents...)
		}
	}

	if len(appCfg.ComponentConfigConnections) != len(appCfg.ComponentIDs) {
		ctxLogger := cctx.GetLogger(ctx, h.l)
		ctxLogger.Warn("app config is missing component configs",
			zap.String("app_config.status", string(appCfg.Status)),
			zap.String("app_config_id", appCfg.ID),
			zap.Int("expected_component_configs", len(appCfg.ComponentIDs)),
			zap.Int("found_component_configs", len(appCfg.ComponentConfigConnections)),
		)

		if !skipAdditionalChecks && appCfg.Status == app.AppConfigStatusActive {
			return nil, errors.New("app config references a component-id which has a config that could not be found")
		}
	}

	return &appCfg, nil
}

func (h *Helpers) CliVerisionAllowed(ctx context.Context, version string) (bool, error) {
	if h.cfg.MinCLIVersion == "" {
		return true, nil
	}

	if version == "development" {
		return true, nil
	}

	if commitSHARegex.MatchString(version) {
		return true, nil
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return false, err
	}

	constraint, err := semver.NewConstraint(">= " + h.cfg.MinCLIVersion)
	if err != nil {
		return false, err
	}

	return constraint.Check(v), nil
}

var commitSHARegex = regexp.MustCompile(`^[0-9a-f]{7}$`)
