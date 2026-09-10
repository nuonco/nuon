package helpers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// AppConfigSection names a preloadable section of an app config.
type AppConfigSection string

const (
	SectionSecrets            AppConfigSection = "secrets"
	SectionBreakGlass         AppConfigSection = "break_glass"
	SectionOperationRole      AppConfigSection = "operation_role"
	SectionPermissions        AppConfigSection = "permissions"
	SectionPolicy             AppConfigSection = "policy"
	SectionRunner             AppConfigSection = "runner"
	SectionSandbox            AppConfigSection = "sandbox"
	SectionInput              AppConfigSection = "input"
	SectionStack              AppConfigSection = "stack"
	SectionKubernetesContexts AppConfigSection = "kubernetes_contexts"
	SectionComponents         AppConfigSection = "components"
	SectionActionWorkflows    AppConfigSection = "action_workflows"
)

var appConfigSectionScopes = map[AppConfigSection]func(*gorm.DB) *gorm.DB{
	SectionSecrets:            PreloadAppSecretsConfig,
	SectionBreakGlass:         PreloadAppBreakGlassConfig,
	SectionOperationRole:      PreloadAppOperationRoleConfig,
	SectionPermissions:        PreloadAppConfigPermissionsConfig,
	SectionPolicy:             PreloadAppConfigPolicyConfig,
	SectionRunner:             PreloadAppConfigRunnerConfig,
	SectionSandbox:            PreloadAppConfigSandboxConfig,
	SectionInput:              PreloadAppConfigInputConfig,
	SectionStack:              PreloadAppConfigStackConfig,
	SectionKubernetesContexts: PreloadAppConfigKubernetesContextsConfig,
	SectionComponents:         PreloadAppConfigComponentConfigConnections,
	SectionActionWorkflows:    PreloadAppActionWorkflowConfigs,
}

// StackRenderAppConfigSections is what internal/pkg/stacks reads.
var StackRenderAppConfigSections = []AppConfigSection{
	SectionInput,
	SectionStack,
	SectionSecrets,
	SectionBreakGlass,
	SectionPermissions,
	SectionOperationRole,
	SectionRunner,
}

// StackAwaitAppConfigSections covers the managed-stack gate plus
// GetFakeSandboxStackData.
var StackAwaitAppConfigSections = []AppConfigSection{
	SectionRunner,
	SectionBreakGlass,
	SectionInput,
	SectionPermissions,
}

// GetScopedAppConfig loads an app config with only the requested sections. It
// skips GetFullAppConfig's CLI-version gate (sync-path only) and its
// missing-component reconciliation, which presupposes the component preload.
func (h *Helpers) GetScopedAppConfig(ctx context.Context, appConfigID string, sections ...AppConfigSection) (*app.AppConfig, error) {
	scopes := make([]func(*gorm.DB) *gorm.DB, 0, len(sections))
	for _, s := range sections {
		scope, ok := appConfigSectionScopes[s]
		if !ok {
			return nil, fmt.Errorf("unknown app config section %q", s)
		}
		scopes = append(scopes, scope)
	}

	appCfg := app.AppConfig{}
	res := h.db.WithContext(ctx).
		Where(app.AppConfig{ID: appConfigID}).
		Scopes(scopes...).
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

	return &appCfg, nil
}
