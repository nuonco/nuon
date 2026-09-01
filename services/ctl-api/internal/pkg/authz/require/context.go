package require

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

// Org returns the selected org ID from context. Empty org means the caller must
// select_org or pass X-Nuon-Org-ID. Use for non-Gin callers (MCP tool handlers).
func Org(ctx context.Context) (string, error) {
	orgID := keys.OrgIDFromContext(ctx)
	if orgID == "" {
		return "", fmt.Errorf("no org selected; call list_orgs then select_org, or pass X-Nuon-Org-ID")
	}
	return orgID, nil
}

// Read requires a selected org and PermissionRead on that org object.
func Read(ctx context.Context) (string, error) {
	orgID, err := Org(ctx)
	if err != nil {
		return "", err
	}
	if err := permission(ctx, orgID, permissions.PermissionRead); err != nil {
		return "", err
	}
	return orgID, nil
}

// Write requires a non-read-only token role, a selected org, and PermissionCreate
// on that org object.
func Write(ctx context.Context) (string, error) {
	if keys.TokenRoleFromContext(ctx) == string(app.RoleTypeOrgReadOnly) {
		return "", fmt.Errorf("this access token is read-only (org_read_only scope) and cannot perform write operations; re-authorize with the org_support or org_admin scope")
	}
	orgID, err := Org(ctx)
	if err != nil {
		return "", err
	}
	if err := permission(ctx, orgID, permissions.PermissionCreate); err != nil {
		return "", err
	}
	return orgID, nil
}

func permission(ctx context.Context, orgID string, perm permissions.Permission) error {
	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}
	if err := acct.AllPermissions.CanPerform(orgID, perm); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}
	return nil
}
