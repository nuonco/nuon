package authz

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/generics"
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
			sources, err := permissions.LegacySources(grants)
			if err != nil {
				return nil, err
			}
			loaded.Grants = append(loaded.Grants, grants...)
			for object, source := range sources {
				source.RoleID, source.PolicyID = role.ID, role.Policies[0].ID
				loaded.Add(object, source)
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

func (c *Client) LoadNativeAccountPermissions(ctx context.Context, accountID, orgID string, catalog *permissions.Catalog) ([]*permissions.CompiledPolicy, error) {
	if accountID == "" {
		return nil, fmt.Errorf("native permissions require an account ID")
	}
	if err := (permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}).ValidateShape(); err != nil {
		return nil, err
	}
	db := c.db.WithContext(ctx)
	var account app.Account
	if err := db.Select("id").Where(&app.Account{ID: accountID}).First(&account).Error; err != nil {
		return nil, fmt.Errorf("native permissions require a live account: %w", err)
	}
	var org app.Org
	result := db.Model(&app.Org{}).Select("id").Where(&app.Org{ID: orgID}).Scan(&org)
	if result.Error != nil {
		return nil, fmt.Errorf("load native permissions org: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return nil, fmt.Errorf("native permissions require a live org")
	}
	var bindings []app.AccountRole
	if err := db.Where(&app.AccountRole{AccountID: accountID, OrgID: generics.NewNullString(orgID)}).
		Find(&bindings).Error; err != nil {
		return nil, fmt.Errorf("load native account bindings: %w", err)
	}
	loaded := make([]*permissions.CompiledPolicy, 0, len(bindings))
	if len(bindings) == 0 {
		return loaded, nil
	}
	roleIDs := make([]string, 0, len(bindings))
	seen := make(map[string]bool, len(bindings))
	for _, binding := range bindings {
		if binding.RoleID == "" || seen[binding.RoleID] {
			return nil, fmt.Errorf("native permissions require one live binding per role")
		}
		seen[binding.RoleID] = true
		roleIDs = append(roleIDs, binding.RoleID)
	}
	var roles []app.Role
	if err := db.Model(&app.Role{}).Where(map[string]any{"id": roleIDs}).Scan(&roles).Error; err != nil {
		return nil, fmt.Errorf("load native roles: %w", err)
	}
	byRole := make(map[string]*app.Role, len(roles))
	for i := range roles {
		byRole[roles[i].ID] = &roles[i]
	}
	var policies []app.Policy
	if err := db.Model(&app.Policy{}).Where(map[string]any{"role_id": roleIDs}).Scan(&policies).Error; err != nil {
		return nil, fmt.Errorf("load native policies: %w", err)
	}
	for _, policy := range policies {
		if role := byRole[policy.RoleID]; role != nil {
			role.Policies = append(role.Policies, policy)
		}
	}
	for _, binding := range bindings {
		role := byRole[binding.RoleID]
		if role == nil {
			return nil, fmt.Errorf("native binding %s requires a live role", binding.ID)
		}
		binding.Org = &org
		compiled, err := role.CompileNativePermissions(binding, catalog)
		if err != nil {
			return nil, err
		}
		loaded = append(loaded, compiled)
	}
	return loaded, nil
}
