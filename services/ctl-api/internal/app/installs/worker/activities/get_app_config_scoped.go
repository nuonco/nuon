package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
)

// @temporal-gen-v2 activity
// @as-wrapper
// @by-field ID
func (a *Activities) getAppConfigForStack(ctx context.Context, ID string) (*app.AppConfig, error) {
	cfg, err := a.appsHelpers.GetScopedAppConfig(ctx, ID, appshelpers.StackRenderAppConfigSections...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config for stack")
	}

	return cfg, nil
}

// @temporal-gen-v2 activity
// @as-wrapper
// @by-field ID
func (a *Activities) getAppConfigForStackAwait(ctx context.Context, ID string) (*app.AppConfig, error) {
	cfg, err := a.appsHelpers.GetScopedAppConfig(ctx, ID, appshelpers.StackAwaitAppConfigSections...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config for stack await")
	}

	return cfg, nil
}

// @temporal-gen-v2 activity
// @as-wrapper
// @by-field ID
func (a *Activities) getAppConfigInputSection(ctx context.Context, ID string) (*app.AppConfig, error) {
	cfg, err := a.appsHelpers.GetScopedAppConfig(ctx, ID, appshelpers.SectionInput)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config input section")
	}

	return cfg, nil
}
