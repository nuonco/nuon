package helpers

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

// ValidateInstallRole rejects a runtime role override the install cannot assume.
// Role selection only resolves a name to an ARN once a workflow plans, so
// without this an unknown role is accepted, a workflow is created, and the
// caller learns about the typo from a failed step. Callers that pass an empty
// role fall through to the configured default and are never rejected here.
func (h *Helpers) ValidateInstallRole(ctx context.Context, installID, role string) error {
	if role == "" {
		return nil
	}

	var install app.Install
	if err := h.db.WithContext(ctx).
		Select("id", "name", "app_config_id").
		Where(app.Install{ID: installID}).
		First(&install).Error; err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	stack, err := h.getInstallStack(ctx, install.ID)
	if err != nil {
		return err
	}
	if stack == nil || stack.ID == "" {
		return fmt.Errorf("role %q is not available: install %q has no stack outputs yet", role, install.Name)
	}

	installState, err := h.GetInstallState(ctx, install.ID, false, true)
	if err != nil {
		return fmt.Errorf("unable to get install state: %w", err)
	}

	appCfg, err := h.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, true)
	if err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	roles, err := operationroles.AvailableRoles(appCfg, &stack.InstallStackOutputs, installState)
	if err != nil {
		return fmt.Errorf("unable to list available roles: %w", err)
	}

	// Role names may be templated, and available roles come back rendered.
	rendered, err := operationroles.RenderRoleName(role, installState)
	if err != nil {
		return fmt.Errorf("unable to render role name %q: %w", role, err)
	}
	if _, ok := roles[rendered]; ok {
		return nil
	}

	available := slices.Sorted(maps.Keys(roles))
	if len(available) == 0 {
		return fmt.Errorf("role %q is not available; install %q has no available roles", role, install.Name)
	}
	return fmt.Errorf("role %q is not available; available roles: %s", role, strings.Join(available, ", "))
}
