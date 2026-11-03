package authz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// AssignableRoles returns the org's roles offerable on the given assignment
// surface (see app.RoleContext*), ordered by role type.
func (c *Client) AssignableRoles(ctx context.Context, orgID, roleContext string) ([]app.Role, error) {
	var roles []app.Role
	res := c.db.WithContext(ctx).
		Preload("Policies").
		Where(app.Role{OrgID: generics.NewNullString(orgID), Version: app.RoleVersionV3}).
		Order("role_type").
		Find(&roles)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to list roles for org %s: %w", orgID, res.Error)
	}

	assignable := make([]app.Role, 0, len(roles))
	for _, role := range roles {
		if role.AllowsContext(roleContext) {
			assignable = append(assignable, role)
		}
	}
	return assignable, nil
}

// ResolveAssignableRole validates that roleType is offerable on the given
// surface in the org and returns its role row. The error lists the roles that
// are assignable there.
func (c *Client) ResolveAssignableRole(ctx context.Context, orgID string, roleType app.RoleType, roleContext string) (*app.Role, error) {
	assignable, err := c.AssignableRoles(ctx, orgID, roleContext)
	if err != nil {
		return nil, err
	}

	types := make([]string, 0, len(assignable))
	for i := range assignable {
		if assignable[i].RoleType == roleType {
			return &assignable[i], nil
		}
		types = append(types, string(assignable[i].RoleType))
	}

	return nil, stderr.NewInvalidRequest(fmt.Errorf("invalid role %q: must be one of %s", roleType, strings.Join(types, ", ")))
}

func (c *Client) ResolveDelegableRole(ctx context.Context, orgID string, roleType app.RoleType, roleContext string) (*app.Role, error) {
	return c.ResolveDelegableRoleAtScope(ctx, permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}, roleType, roleContext)
}

func RoleBindingScope(orgID string, requested *permissions.Scope) (permissions.Scope, error) {
	scope := permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}
	if requested != nil {
		scope = *requested
		if scope.OrgID == "" {
			scope.OrgID = orgID
		}
	}
	if scope.OrgID != orgID {
		return permissions.Scope{}, fmt.Errorf("role scope must belong to the selected org")
	}
	if err := validateBindingScope(scope); err != nil {
		return permissions.Scope{}, err
	}
	return scope, nil
}

func (c *Client) ResolveDelegableRoleAtScope(ctx context.Context, scope permissions.Scope, roleType app.RoleType, roleContext string) (*app.Role, error) {
	if err := validateBindingScope(scope); err != nil {
		return nil, err
	}
	orgID := scope.OrgID
	role, err := c.ResolveAssignableRole(ctx, orgID, roleType, roleContext)
	if err != nil {
		return nil, err
	}
	policy, err := decodeScopedRole(*role, orgID)
	if err != nil {
		return nil, err
	}
	if policy.Kind() != permissions.PolicyKindBundle || policy.API() != permissions.APICtl || role.CatalogRevision <= 0 {
		return nil, fmt.Errorf("delegation requires a customer catalog bundle")
	}
	credential, ok := cctx.VerifiedCredentialFromContext(ctx)
	if !ok || !customerCredentialAllowed(credential) {
		return nil, fmt.Errorf("verified customer credential required for delegation")
	}
	delegated, err := policy.Compile(orgID, scope)
	if err != nil {
		return nil, stderr.NewInvalidRequest(err)
	}
	decision, err := authorize(c.db.WithContext(ctx), credential, permissions.OrgWrite,
		permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}, time.Now())
	if err != nil {
		return nil, err
	}
	if !decision.Allowed {
		return nil, stderr.ErrAuthorization{Err: fmt.Errorf("delegation denied"), Description: "access denied"}
	}
	target, err := resolveTarget(c.db.WithContext(ctx), scope)
	if err != nil {
		return nil, err
	}
	grants, err := loadScopedGrants(c.db.WithContext(ctx), credential.AccountID, orgID)
	if err != nil {
		return nil, err
	}
	ceiling, reason, err := loadCustomerCredential(c.db.WithContext(ctx), credential, orgID, time.Now())
	if err != nil {
		return nil, err
	}
	for _, grant := range delegated {
		if !permissions.ContainsGrant(grants.Grants, target, grant.Permissions) ||
			!permissions.ContainsGrant(credential.Ceiling, target, grant.Permissions) || reason != ReasonAllowed ||
			!permissions.ContainsGrant(ceiling, target, grant.Permissions) {
			return nil, stderr.ErrAuthorization{Err: fmt.Errorf("role exceeds delegable authority"), Description: "role exceeds your credential's permissions"}
		}
	}
	return role, nil
}
