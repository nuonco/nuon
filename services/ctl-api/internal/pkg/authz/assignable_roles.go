package authz

import (
	"context"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (c *Client) AssignableRoles(ctx context.Context, orgID, roleContext string) ([]app.Role, error) {
	var roles []app.Role
	res := c.db.WithContext(ctx).
		Where(app.Role{OrgID: generics.NewNullString(orgID)}).
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

	return nil, fmt.Errorf("invalid role %q: must be one of %s", roleType, strings.Join(types, ", "))
}
