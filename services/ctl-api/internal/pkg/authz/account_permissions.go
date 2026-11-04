package authz

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (c *Client) LoadAccountPermissions(ctx context.Context, account *app.Account) (*permissions.AccountPermissions, error) {
	if account == nil || account.ID == "" || account.DeletedAt != 0 {
		return nil, fmt.Errorf("permissions require a live account")
	}
	var bindings []app.AccountRole
	for _, role := range account.Roles {
		if role.HasPolicyDocument() {
			if err := c.db.WithContext(ctx).
				Preload("Org").
				Where(app.AccountRole{AccountID: account.ID}).
				Find(&bindings).Error; err != nil {
				return nil, fmt.Errorf("load account bindings: %w", err)
			}
			break
		}
	}
	byRole := make(map[string][]app.AccountRole, len(bindings))
	for _, binding := range bindings {
		byRole[binding.RoleID] = append(byRole[binding.RoleID], binding)
	}
	loaded := &permissions.AccountPermissions{Sources: make(map[string]permissions.PermissionSource)}
	for _, role := range account.Roles {
		if role.HasPolicyDocument() {
			bound := byRole[role.ID]
			if len(bound) != 1 || bound[0].AccountID != account.ID {
				return nil, fmt.Errorf("role %s requires one live account binding", role.ID)
			}
			grants, err := role.CompilePermissions(bound[0])
			if err != nil {
				return nil, err
			}
			for _, grant := range grants {
				verb, err := permissions.LegacyOrgPermission(grant.Permissions)
				if err != nil {
					return nil, err
				}
				if verb != permissions.PermissionUnknown {
					loaded.Add(grant.Scope.OrgID, permissions.PermissionSource{
						RoleID: role.ID, PolicyID: role.Policies[0].ID, Permission: verb, Grant: &grant,
					})
				}
			}
			continue
		}
		for _, policy := range role.Policies {
			legacy := permissions.Set(permissions.NewSet())
			_ = legacy.Add(policy.Permissions)
			for object, verb := range legacy {
				loaded.Add(object, permissions.PermissionSource{RoleID: role.ID, PolicyID: policy.ID, Permission: verb})
			}
		}
	}
	return loaded, nil
}
