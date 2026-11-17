package authz

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (c *Client) MigrateWorkloadGrantPolicies(ctx context.Context, orgID string, apply bool) (int, error) {
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
		types := append(slices.Clone(app.PrivateRoleTypes), app.RoleTypeRunner)
		if err := tx.Preload("Policies").Where(app.Role{OrgID: generics.NewNullString(orgID)}).
			Where("role_type IN ?", types).Find(&roles).Error; err != nil {
			return err
		}
		roleIDs := make([]string, 0, len(roles))
		runners := 0
		for _, role := range roles {
			roleIDs = append(roleIDs, role.ID)
			if role.RoleType == app.RoleTypeRunner {
				runners++
			}
		}
		if runners > 1 {
			return fmt.Errorf("org has multiple managed runner roles")
		}
		var bindings []app.AccountRole
		if err := tx.Where("role_id IN ?", roleIDs).Find(&bindings).Error; err != nil {
			return err
		}
		holders := make(map[string][]string)
		accounts := make([]string, 0, len(bindings))
		for _, binding := range bindings {
			if binding.OrgID.ValueString() != orgID || binding.AccountID == "" || slices.Contains(holders[binding.RoleID], binding.AccountID) {
				return fmt.Errorf("workload role %s has an inconsistent or duplicate binding", binding.RoleID)
			}
			holders[binding.RoleID] = append(holders[binding.RoleID], binding.AccountID)
			accounts = append(accounts, binding.AccountID)
		}
		var allBindings []app.AccountRole
		if err := tx.Where(app.AccountRole{OrgID: generics.NewNullString(orgID)}).
			Where("account_id IN ?", accounts).Find(&allBindings).Error; err != nil {
			return err
		}
		for _, binding := range allBindings {
			if !slices.Contains(holders[binding.RoleID], binding.AccountID) {
				return fmt.Errorf("account %s mixes workload and other org roles; requires explicit effective-access conversion", binding.AccountID)
			}
		}
		type boundary struct {
			api    permissions.API
			grants bool
		}
		boundaries := make(map[string]boundary)
		for _, role := range roles {
			if len(holders[role.ID]) == 0 {
				continue
			}
			api, _, err := APIFromRole(&role)
			if err != nil {
				return err
			}
			current := boundary{api: api, grants: len(role.Policies) == 1 && role.Policies[0].HasGrantDocument()}
			for _, accountID := range holders[role.ID] {
				if previous, exists := boundaries[accountID]; exists &&
					(previous.grants != current.grants || (!current.grants && previous.api != current.api)) {
					return fmt.Errorf("account %s mixes policy formats or API restrictions; requires explicit effective-access conversion", accountID)
				}
				boundaries[accountID] = current
			}
		}
		for _, accountID := range accounts {
			var account app.Account
			if err := tx.Where(app.Account{ID: accountID, AccountType: app.AccountTypeService}).Take(&account).Error; err != nil {
				return fmt.Errorf("workload role holder must be a live service account: %w", err)
			}
		}
		documents := make(map[string][]byte)
		for _, role := range roles {
			if role.RoleType == app.RoleTypeRunner {
				if !role.Managed || len(role.Contexts) != 0 {
					return fmt.Errorf("runner requires its managed held-only role")
				}
			} else {
				if len(holders[role.ID]) == 0 {
					continue
				}
				if len(holders[role.ID]) != 1 {
					return fmt.Errorf("private role %s is shared", role.ID)
				}
				if err := validatePrivateRole(tx, holders[role.ID][0], orgID, &role); err != nil {
					return err
				}
			}
			if len(role.Policies) != 1 || role.Policies[0].RoleID != role.ID || role.Policies[0].OrgID.ValueString() != orgID {
				return fmt.Errorf("workload role %s requires one consistently owned policy", role.ID)
			}
			if role.Policies[0].HasGrantDocument() {
				if _, err := decodeWorkloadGrant(&role); err != nil {
					return err
				}
				continue
			}
			if role.Version != app.RoleVersionV3 || role.Policies[0].Document == nil {
				return fmt.Errorf("workload role %s requires v3 authority before conversion", role.ID)
			}
			policy, err := workloadGrantPolicy(&role)
			if err != nil {
				return fmt.Errorf("convert workload role %s: %w", role.ID, err)
			}
			if (role.RoleType == app.RoleTypeRunnerInstall || role.RoleType == app.RoleTypeRunnerJob) && policy.API != permissions.APIRunner {
				return fmt.Errorf("runner role %s without the runner API restriction requires a cross-listener access recipe", role.ID)
			}
			document, err := json.Marshal(policy)
			if err != nil {
				return err
			}
			documents[role.Policies[0].ID] = document
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
