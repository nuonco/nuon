package authz

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (c *Client) LoadAccountLegacyAccess(ctx context.Context, account *app.Account) (*permissions.AccountPermissions, error) {
	if account == nil || account.ID == "" || account.DeletedAt != 0 {
		return nil, fmt.Errorf("permissions require a live account")
	}
	grantOrgs := make(map[string]bool)
	for _, role := range account.Roles {
		for _, policy := range role.Policies {
			if policy.HasGrantDocument() {
				grantOrgs[role.OrgID.ValueString()] = true
			}
		}
	}
	if len(grantOrgs) == 0 {
		return c.LoadLegacyAccountPermissions(ctx, account)
	}
	catalog, err := permissions.DefaultCatalog()
	if err != nil {
		return nil, err
	}
	for orgID := range grantOrgs {
		if _, err := c.LoadAccountGrants(ctx, account.ID, orgID, catalog); err != nil {
			return nil, err
		}
	}
	legacy := *account
	legacy.Roles = make([]app.Role, 0, len(account.Roles))
	for _, role := range account.Roles {
		if !grantOrgs[role.OrgID.ValueString()] {
			legacy.Roles = append(legacy.Roles, role)
		}
	}
	return c.LoadLegacyAccountPermissions(ctx, &legacy)
}

func (c *Client) LoadLegacyAccountPermissions(ctx context.Context, account *app.Account) (*permissions.AccountPermissions, error) {
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
			grants, err := role.CompileLegacyPermissions(bound[0])
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

func (c *Client) AccountUsesGrants(ctx context.Context, accountID, orgID string) (bool, error) {
	if accountID == "" || orgID == "" {
		return false, fmt.Errorf("grant selection requires account and org IDs")
	}
	db := c.db.WithContext(ctx)
	roles := db.Model(&app.AccountRole{}).Select("role_id").Where(app.AccountRole{AccountID: accountID, OrgID: generics.NewNullString(orgID)})
	var policies []app.Policy
	if err := db.Model(&app.Policy{}).Where("role_id IN (?)", roles).Scan(&policies).Error; err != nil {
		return false, err
	}
	for _, policy := range policies {
		if policy.HasGrantDocument() {
			return true, nil
		}
	}
	return false, nil
}

func (c *Client) LoadAccountGrants(ctx context.Context, accountID, orgID string, catalog *permissions.Catalog) ([]*permissions.CompiledPolicy, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account grants require an account ID")
	}
	if err := (permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}).ValidateShape(); err != nil {
		return nil, err
	}
	db := c.db.WithContext(ctx)
	var account app.Account
	if err := db.Select("id").Where(&app.Account{ID: accountID}).First(&account).Error; err != nil {
		return nil, fmt.Errorf("account grants require a live account: %w", err)
	}
	var org app.Org
	result := db.Model(&app.Org{}).Select("id").Where(&app.Org{ID: orgID}).Scan(&org)
	if result.Error != nil {
		return nil, fmt.Errorf("load account grants org: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return nil, fmt.Errorf("account grants require a live org")
	}
	var bindings []app.AccountRole
	if err := db.Where(&app.AccountRole{AccountID: accountID, OrgID: generics.NewNullString(orgID)}).
		Find(&bindings).Error; err != nil {
		return nil, fmt.Errorf("load account grant bindings: %w", err)
	}
	loaded := make([]*permissions.CompiledPolicy, 0, len(bindings))
	if len(bindings) == 0 {
		return loaded, nil
	}
	roleIDs := make([]string, 0, len(bindings))
	seen := make(map[string]bool, len(bindings))
	for _, binding := range bindings {
		if binding.RoleID == "" || seen[binding.RoleID] {
			return nil, fmt.Errorf("account grants require one live binding per role")
		}
		seen[binding.RoleID] = true
		roleIDs = append(roleIDs, binding.RoleID)
	}
	var roles []app.Role
	if err := db.Model(&app.Role{}).Where(map[string]any{"id": roleIDs}).Scan(&roles).Error; err != nil {
		return nil, fmt.Errorf("load grant roles: %w", err)
	}
	byRole := make(map[string]*app.Role, len(roles))
	for i := range roles {
		byRole[roles[i].ID] = &roles[i]
	}
	var policies []app.Policy
	if err := db.Model(&app.Policy{}).Where(map[string]any{"role_id": roleIDs}).Scan(&policies).Error; err != nil {
		return nil, fmt.Errorf("load grant policies: %w", err)
	}
	for _, policy := range policies {
		if role := byRole[policy.RoleID]; role != nil {
			role.Policies = append(role.Policies, policy)
		}
	}
	for _, binding := range bindings {
		role := byRole[binding.RoleID]
		if role == nil {
			return nil, fmt.Errorf("grant binding %s requires a live role", binding.ID)
		}
		binding.Org = &org
		compiled, err := role.CompileGrants(binding, catalog)
		if err != nil {
			return nil, err
		}
		loaded = append(loaded, compiled)
	}
	return loaded, nil
}
