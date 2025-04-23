package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
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
		).
		First(&appCfg)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app config")
	}

	return &appCfg, nil
}
