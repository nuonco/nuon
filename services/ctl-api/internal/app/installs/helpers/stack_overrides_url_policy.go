package helpers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/validate"
	"github.com/nuonco/nuon/pkg/urlpolicy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/validation"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (h *Helpers) CheckStackOverridesURLPolicy(ctx context.Context, installID string, vpcURL, runnerURL *string, stacks []config.CustomNestedStack) error {
	if !h.stackOverridesRestricted(vpcURL, runnerURL, stacks) {
		return nil
	}

	var install app.Install
	if err := h.db.WithContext(ctx).Scopes(scopes.WithDisableViews).
		Select("id", "app_id").
		Where(app.Install{ID: installID}).
		Take(&install).Error; err != nil {
		return fmt.Errorf("unable to get install for stack overrides check: %w", err)
	}
	return h.CheckAppStackOverridesURLPolicy(ctx, install.AppID, vpcURL, runnerURL, stacks)
}

func (h *Helpers) CheckAppStackOverridesURLPolicy(ctx context.Context, appID string, vpcURL, runnerURL *string, stacks []config.CustomNestedStack) error {
	if !h.stackOverridesRestricted(vpcURL, runnerURL, stacks) {
		return nil
	}

	var stackConfig app.AppStackConfig
	err := h.db.WithContext(ctx).
		Select("id", "type").
		Where(app.AppStackConfig{AppID: appID}).
		Order("created_at DESC").
		Take(&stackConfig).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("unable to get app stack config for stack overrides check: %w", err)
	}
	urls := stackOverrideURLs(vpcURL, runnerURL, stacks)
	urls.Type = string(stackConfig.Type)

	return validation.CheckURLPolicy(h.cfg.URLPolicy, validate.StackURLPolicyViolations(h.cfg.URLPolicy, "stack_overrides", urls))
}

func (h *Helpers) stackOverridesRestricted(vpcURL, runnerURL *string, stacks []config.CustomNestedStack) bool {
	if h.cfg == nil || !h.cfg.URLPolicy.Restricted(urlpolicy.CategoryStackTemplates) {
		return false
	}
	return len(validate.StackURLRefs("stack_overrides", stackOverrideURLs(vpcURL, runnerURL, stacks))) > 0
}

func stackOverrideURLs(vpcURL, runnerURL *string, stacks []config.CustomNestedStack) validate.StackURLs {
	urls := validate.StackURLs{CustomNestedStacks: stacks}
	if vpcURL != nil {
		urls.VPCNestedTemplateURL = *vpcURL
	}
	if runnerURL != nil {
		urls.RunnerNestedTemplateURL = *runnerURL
	}
	return urls
}
