package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

// requireWriteScope blocks write MCP tools when the authenticating token is
// read-only (OAuth scope org_read_only). Tokens without an explicit role
// (non-OAuth tokens) are not restricted here and fall back to the account's
// normal RBAC.
func requireWriteScope(ctx context.Context) error {
	if keys.TokenRoleFromContext(ctx) == string(app.RoleTypeOrgReadOnly) {
		return fmt.Errorf("this access token is read-only (org_read_only scope) and cannot perform write operations; re-authorize with the org_support or org_admin scope")
	}
	return nil
}
