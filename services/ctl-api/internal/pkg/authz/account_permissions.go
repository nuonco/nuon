package authz

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

type PermissionSource struct {
	RoleID     string
	PolicyID   string
	Permission permissions.Permission
	Grant      *permissions.ScopedPermissions
}

type AccountPermissions struct {
	Sources map[string]PermissionSource
}

func (c *Client) LoadAccountPermissions(ctx context.Context, account *app.Account) (*AccountPermissions, error) {
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
	loaded := &AccountPermissions{Sources: make(map[string]PermissionSource)}
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
					loaded.add(grant.Scope.OrgID, PermissionSource{
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
				loaded.add(object, PermissionSource{RoleID: role.ID, PolicyID: policy.ID, Permission: verb})
			}
		}
	}
	return loaded, nil
}

func (p *AccountPermissions) add(object string, source PermissionSource) {
	if previous, ok := p.Sources[object]; ok && previous.Permission == permissions.PermissionAll {
		return
	}
	p.Sources[object] = source
}

func (p *AccountPermissions) Grants() []permissions.ScopedPermissions {
	grants := make([]permissions.ScopedPermissions, 0)
	for _, source := range p.Sources {
		if source.Grant != nil {
			grants = append(grants, *source.Grant)
		}
	}
	return grants
}

func (p *AccountPermissions) Legacy() permissions.Set {
	legacy := permissions.Set(permissions.NewSet())
	for object, source := range p.Sources {
		if source.Grant == nil {
			legacy[object] = source.Permission
		}
	}
	return legacy
}
