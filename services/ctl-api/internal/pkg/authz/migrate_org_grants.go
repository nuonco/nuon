package authz

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (c *Client) MigrateOrgGrantPolicies(ctx context.Context, orgID string, apply bool) (int, error) {
	converted := 0
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if apply {
			if err := tx.Exec("LOCK TABLE roles, policies, account_roles IN SHARE ROW EXCLUSIVE MODE").Error; err != nil {
				return err
			}
		}
		var org app.Org
		if err := tx.Where(app.Org{ID: orgID}, "id").First(&org).Error; err != nil {
			return err
		}
		var roles []app.Role
		if err := tx.Preload("Policies").Where(app.Role{OrgID: generics.NewNullString(orgID)}).
			Where("role_type IN ?", []app.RoleType{app.RoleTypeOrgAdmin, app.RoleTypeOrgReadOnly}).Find(&roles).Error; err != nil {
			return err
		}
		if len(roles) != 2 || roles[0].RoleType == roles[1].RoleType {
			return fmt.Errorf("org requires exactly one managed Admin and Reader role before conversion")
		}
		eligible := make(map[string]bool, len(roles))
		documents := make(map[string][]byte, len(roles))
		for _, role := range roles {
			if !role.Managed || len(role.Policies) != 1 {
				return fmt.Errorf("role %s is not a standard managed role", role.ID)
			}
			policy := role.Policies[0]
			if policy.RoleID != role.ID || policy.OrgID.ValueString() != orgID {
				return fmt.Errorf("role %s has inconsistent policy ownership", role.ID)
			}
			eligible[role.ID] = true
			if isStandardOrgGrant(role) {
				continue
			}
			if role.Version != app.RoleVersionV3 || policy.Document == nil || policy.HasGrantDocument() {
				return fmt.Errorf("role %s requires standard v3 authority before conversion", role.ID)
			}
			projection, err := policy.PermissionProjection()
			if err != nil {
				return err
			}
			var api permissions.API
			keys := 1
			if value, present := projection[permissions.APIPermissionKey]; present {
				if value == nil || *value != string(permissions.APICtl) {
					return fmt.Errorf("role %s differs from its baseline API restriction", role.ID)
				}
				api = permissions.APICtl
				keys++
			}
			verb := permissions.PermissionAll
			if role.RoleType == app.RoleTypeOrgReadOnly {
				verb = permissions.PermissionRead
			}
			if len(projection) != keys || projection[orgID] == nil || *projection[orgID] != string(verb) {
				return fmt.Errorf("role %s differs from its baseline authority", role.ID)
			}
			document, err := orgGrantPolicy(orgID, role.RoleType, api)
			if err != nil {
				return err
			}
			documents[policy.ID] = document
		}
		roleIDs := []string{roles[0].ID, roles[1].ID}
		var bindings []app.AccountRole
		if err := tx.Where("role_id IN ?", roleIDs).Find(&bindings).Error; err != nil {
			return err
		}
		accounts := make([]string, 0, len(bindings))
		seen := make(map[[2]string]bool)
		for _, binding := range bindings {
			key := [2]string{binding.AccountID, binding.RoleID}
			if binding.OrgID.ValueString() != orgID || binding.AccountID == "" || seen[key] {
				return fmt.Errorf("role %s has an inconsistent or duplicate binding", binding.RoleID)
			}
			seen[key] = true
			accounts = append(accounts, binding.AccountID)
		}
		var allBindings []app.AccountRole
		if err := tx.Where(app.AccountRole{OrgID: generics.NewNullString(orgID)}).Where("account_id IN ?", accounts).Find(&allBindings).Error; err != nil {
			return err
		}
		for _, binding := range allBindings {
			if !eligible[binding.RoleID] {
				return fmt.Errorf("account %s mixes standard and other org roles; requires explicit effective-access conversion", binding.AccountID)
			}
		}
		converted = len(documents)
		if !apply {
			return nil
		}
		for policyID, document := range documents {
			if err := tx.Model(&app.Policy{}).Where(app.Policy{ID: policyID, OrgID: generics.NewNullString(orgID)}).
				UpdateColumns(map[string]any{"document": document, "permissions": nil}).Error; err != nil {
				return err
			}
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: !apply})
	if err != nil {
		return 0, err
	}
	return converted, nil
}
