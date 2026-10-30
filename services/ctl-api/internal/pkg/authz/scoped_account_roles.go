package authz

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (c *Client) AddAccountRoleAtScope(ctx context.Context, accountID, roleID string, scope permissions.Scope) error {
	return c.mutateScopedAccountRoles(ctx, accountID, scope, func(tx *gorm.DB, account app.Account) error {
		return addScopedAccountRole(tx, account, roleID, scope)
	})
}

func (c *Client) SetAccountRoleAtScope(ctx context.Context, accountID, roleID string, scope permissions.Scope) error {
	return c.mutateScopedAccountRoles(ctx, accountID, scope, func(tx *gorm.DB, account app.Account) error {
		if err := tx.Unscoped().Where(scopedAccountRole(accountID, "", scope)).Delete(&app.AccountRole{}).Error; err != nil {
			return err
		}
		return addScopedAccountRole(tx, account, roleID, scope)
	})
}

func (c *Client) RemoveAccountRoleAtScope(ctx context.Context, accountID, roleID string, scope permissions.Scope) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}
	return c.mutateScopedAccountRoles(ctx, accountID, scope, func(tx *gorm.DB, account app.Account) error {
		if err := tx.Unscoped().Where(scopedAccountRole(accountID, roleID, scope)).Delete(&app.AccountRole{}).Error; err != nil {
			return err
		}
		if scope.Kind == permissions.KindOrg {
			missing, err := bindingTargetMissing(tx.Clauses(clause.Locking{Strength: "SHARE"}).Session(&gorm.Session{}), scope)
			if err != nil || missing {
				return err
			}
		}
		_, err := loadScopedGrants(tx, accountID, scope.OrgID)
		if errors.Is(err, errNoScopedGrants) {
			return nil
		}
		return err
	})
}

func scopedAccountRole(accountID, roleID string, scope permissions.Scope) app.AccountRole {
	return app.AccountRole{
		AccountID: accountID, RoleID: roleID, OrgID: generics.NewNullString(scope.OrgID),
		ScopeKind: scope.Kind, ScopeID: scope.ID,
	}
}

func validateBindingScope(scope permissions.Scope) error {
	if err := scope.ValidateShape(); err != nil {
		return err
	}
	if scope.Kind != permissions.KindOrg && scope.Kind != permissions.KindApp && scope.Kind != permissions.KindInstall {
		return fmt.Errorf("role binding requires an org, app or install scope")
	}
	return nil
}

func lockRoleSubject(tx *gorm.DB, accountID string) (app.Account, error) {
	if accountID == "" {
		return app.Account{}, fmt.Errorf("account ID is required")
	}
	if err := requireRoleMutationIsolation(tx); err != nil {
		return app.Account{}, err
	}
	var account app.Account
	err := tx.Select("id", "account_type").Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(app.Account{ID: accountID}).Take(&account).Error
	return account, err
}

func requireRoleMutationIsolation(tx *gorm.DB) error {
	var isolation string
	if err := tx.Raw("SHOW transaction_isolation").Scan(&isolation).Error; err != nil {
		return err
	}
	if isolation != "read committed" {
		return fmt.Errorf("role mutations require read-committed isolation")
	}
	return nil
}

func (c *Client) mutateScopedAccountRoles(ctx context.Context, accountID string, scope permissions.Scope, mutate func(*gorm.DB, app.Account) error) error {
	if err := validateBindingScope(scope); err != nil {
		return err
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := resolveTarget(tx.Clauses(clause.Locking{Strength: "SHARE"}).Session(&gorm.Session{}), scope); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			missing, checkErr := bindingTargetMissing(tx.Clauses(clause.Locking{Strength: "SHARE"}).Session(&gorm.Session{}), scope)
			if checkErr != nil {
				return checkErr
			}
			if !missing {
				return err
			}
		}
		account, err := lockRoleSubject(tx, accountID)
		if err != nil {
			return err
		}
		switch account.AccountType {
		case app.AccountTypeAuth, app.AccountTypeAuth0, app.AccountTypeService:
		default:
			return fmt.Errorf("unsupported role subject type %q", account.AccountType)
		}
		var managed int64
		if err := tx.Unscoped().Model(&app.ManagedServiceAccount{}).Where(app.ManagedServiceAccount{AccountID: accountID}).Count(&managed).Error; err != nil {
			return err
		}
		if managed != 0 {
			return fmt.Errorf("managed account roles must be changed through their owner")
		}
		var bindings []app.AccountRole
		if err := tx.Preload("Role.Policies").Where(app.AccountRole{AccountID: accountID}).Find(&bindings).Error; err != nil {
			return err
		}
		for _, binding := range bindings {
			policy, err := decodeScopedRole(binding.Role, binding.OrgID.ValueString())
			if err != nil {
				return err
			}
			if policy.Kind() != permissions.PolicyKindBundle || binding.Role.CatalogRevision <= 0 {
				return fmt.Errorf("scoped assignment requires catalog bundles")
			}
			if _, err := policy.Compile(binding.OrgID.ValueString(), permissions.Scope{
				Kind: binding.ScopeKind, OrgID: binding.OrgID.ValueString(), ID: binding.ScopeID,
			}); err != nil {
				return err
			}
			if account.AccountType == app.AccountTypeService && binding.OrgID.ValueString() != scope.OrgID {
				return fmt.Errorf("service account bindings must remain in one tenant")
			}
		}
		return mutate(tx, account)
	})
}

func addScopedAccountRole(tx *gorm.DB, account app.Account, roleID string, scope permissions.Scope) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}
	if _, err := resolveTarget(tx.Clauses(clause.Locking{Strength: "SHARE"}).Session(&gorm.Session{}), scope); err != nil {
		return err
	}
	var role app.Role
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Preload("Policies", func(db *gorm.DB) *gorm.DB {
		return db.Clauses(clause.Locking{Strength: "SHARE"})
	}).Where(app.Role{ID: roleID}).Take(&role).Error; err != nil {
		return err
	}
	policy, err := decodeScopedRole(role, scope.OrgID)
	if err != nil {
		return err
	}
	roleContext := app.RoleContextTeam
	if account.AccountType == app.AccountTypeService {
		roleContext = app.RoleContextServiceAccount
	}
	if policy.Kind() != permissions.PolicyKindBundle || role.CatalogRevision <= 0 || !role.AllowsContext(roleContext) {
		return fmt.Errorf("role is not an assignable catalog bundle for this subject")
	}
	if _, err := policy.Compile(scope.OrgID, scope); err != nil {
		return err
	}
	binding := scopedAccountRole(account.ID, roleID, scope)
	var count int64
	if err := tx.Model(&app.AccountRole{}).Where(binding).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		if err := tx.Create(&binding).Error; err != nil {
			return err
		}
	}
	_, err = loadScopedGrants(tx, account.ID, scope.OrgID)
	return err
}

func (c *Client) RetireAccountRoleScopes(ctx context.Context, scope permissions.Scope) error {
	if err := validateBindingScope(scope); err != nil {
		return err
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := requireRoleMutationIsolation(tx); err != nil {
			return err
		}
		if _, err := resolveTarget(tx.Clauses(clause.Locking{Strength: "UPDATE"}).Session(&gorm.Session{}), scope); err != nil {
			return err
		}
		query := tx.Where(app.AccountRole{OrgID: generics.NewNullString(scope.OrgID)})
		switch scope.Kind {
		case permissions.KindOrg:
			query = query.Where("scope_kind <> '' AND scope_id <> ''")
		case permissions.KindApp:
			installs := tx.Unscoped().Model(&app.Install{}).Scopes(scopes.WithDisableViews).
				Select("id").Where(app.Install{OrgID: scope.OrgID, AppID: scope.ID})
			query = query.Where(tx.Where(app.AccountRole{ScopeKind: scope.Kind, ScopeID: scope.ID}).
				Or(tx.Where(app.AccountRole{ScopeKind: permissions.KindInstall}).Where("scope_id IN (?)", installs)))
		case permissions.KindInstall:
			query = query.Where(app.AccountRole{ScopeKind: scope.Kind, ScopeID: scope.ID})
		}
		return query.Unscoped().Delete(&app.AccountRole{}).Error
	})
}
