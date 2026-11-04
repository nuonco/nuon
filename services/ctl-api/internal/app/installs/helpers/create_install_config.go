package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type CreateInstallConfigParams = build.InstallConfigInput

// ValidateStackOverrides validates per-install stack template override fields.
func ValidateStackOverrides(vpcURL, runnerURL *string, stacks []config.CustomNestedStack) error {
	if vpcURL != nil && *vpcURL != "" {
		if err := config.ValidateTemplateURL(*vpcURL, "vpc_nested_template_url"); err != nil {
			return err
		}
	}
	if runnerURL != nil && *runnerURL != "" {
		if err := config.ValidateTemplateURL(*runnerURL, "runner_nested_template_url"); err != nil {
			return err
		}
	}

	seen := make(map[string]bool, len(stacks))
	for i, stack := range stacks {
		if stack.Name == "" {
			return fmt.Errorf("custom_nested_stacks[%d]: name is required", i)
		}
		if seen[stack.Name] {
			return fmt.Errorf("custom_nested_stacks[%d]: duplicate name %q", i, stack.Name)
		}
		seen[stack.Name] = true
		if stack.TemplateURL == "" {
			return fmt.Errorf("custom_nested_stacks[%d] (%s): template_url is required", i, stack.Name)
		}
		if err := config.ValidateTemplateURL(stack.TemplateURL, fmt.Sprintf("custom_nested_stacks[%d].template_url", i)); err != nil {
			return err
		}
	}
	return nil
}

func (h *Helpers) CreateInstallConfig(ctx context.Context, installID string, req *CreateInstallConfigParams) (*app.InstallConfig, error) {
	if err := ValidateStackOverrides(req.VPCNestedTemplateURL, req.RunnerNestedTemplateURL, req.CustomNestedStacks); err != nil {
		return nil, fmt.Errorf("invalid stack overrides: %w", err)
	}
	if err := req.ValidateActionTimeout(); err != nil {
		return nil, err
	}
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var install struct{ ID string }
	if err := h.db.WithContext(ctx).Model(&app.Install{}).Scopes(scopes.WithDisableViews).
		Where(app.Install{ID: installID, OrgID: orgID}).Take(&install).Error; err != nil {
		return nil, fmt.Errorf("get install for config creation: %w", err)
	}

	installConfig := build.InstallConfig(*req, installID)
	if err := h.db.WithContext(ctx).Create(installConfig).Error; err != nil {
		return nil, fmt.Errorf("unable to create install config: %w", err)
	}
	return installConfig, nil
}
