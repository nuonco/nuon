package authz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type PrincipalGrants struct {
	API    permissions.API
	Grants []permissions.ScopedPermissions
}

func (c *Client) LoadScopedGrants(ctx context.Context, accountID, orgID string) (PrincipalGrants, error) {
	var result PrincipalGrants
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = loadScopedGrants(tx, accountID, orgID)
		return err
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return PrincipalGrants{}, err
	}
	return result, nil
}

func loadScopedGrants(db *gorm.DB, accountID, orgID string) (PrincipalGrants, error) {
	if accountID == "" {
		return PrincipalGrants{}, fmt.Errorf("account ID is required")
	}
	if _, err := resolveTarget(db, permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}); err != nil {
		return PrincipalGrants{}, err
	}
	var account app.Account
	if err := db.Select("id", "account_type").Where(app.Account{ID: accountID}).Take(&account).Error; err != nil {
		return PrincipalGrants{}, fmt.Errorf("load grant subject: %w", err)
	}
	switch account.AccountType {
	case app.AccountTypeAuth, app.AccountTypeAuth0, app.AccountTypeService:
	default:
		return PrincipalGrants{}, fmt.Errorf("unsupported grant subject type %q", account.AccountType)
	}
	var managed app.ManagedServiceAccount
	if err := db.Unscoped().Where(app.ManagedServiceAccount{AccountID: accountID}).Find(&managed).Error; err != nil {
		return PrincipalGrants{}, fmt.Errorf("load managed account ownership: %w", err)
	}
	if managed.AccountID != "" && (managed.DeletedAt != 0 || managed.OrgID != orgID || managed.InstanceKey == "" || account.AccountType != app.AccountTypeService) {
		return PrincipalGrants{}, fmt.Errorf("invalid managed account tenant or lifecycle")
	}
	var bindings []app.AccountRole
	if err := db.Preload("Role.Policies").Where(app.AccountRole{AccountID: accountID}).Find(&bindings).Error; err != nil {
		return PrincipalGrants{}, fmt.Errorf("load scoped role bindings: %w", err)
	}
	var result PrincipalGrants
	for _, binding := range bindings {
		role := binding.Role
		if account.AccountType != app.AccountTypeService && binding.OrgID.ValueString() != orgID && role.OrgID.ValueString() != orgID {
			continue
		}
		if binding.OrgID.ValueString() != orgID || role.OrgID.ValueString() != orgID || role.ID == "" || binding.RoleID != role.ID {
			return PrincipalGrants{}, fmt.Errorf("role binding does not belong to the selected tenant")
		}
		if role.Version != app.RoleVersionV3 || len(role.Policies) != 1 {
			return PrincipalGrants{}, fmt.Errorf("scoped grants require a v3 role with one policy")
		}
		stored := role.Policies[0]
		if stored.OrgID.ValueString() != orgID || stored.RoleID != role.ID || len(stored.Permissions) != 0 {
			return PrincipalGrants{}, fmt.Errorf("policy must belong to its role tenant and contain no legacy authority")
		}
		policy, err := permissions.DecodePolicy(role.Version, stored.Document)
		if err != nil {
			return PrincipalGrants{}, fmt.Errorf("decode role policy: %w", err)
		}
		scope := permissions.Scope{Kind: binding.ScopeKind, OrgID: orgID, ID: binding.ScopeID}
		grants, err := policy.Compile(orgID, scope)
		if err != nil {
			return PrincipalGrants{}, err
		}
		switch policy.Kind() {
		case permissions.PolicyKindBundle:
			if managed.AccountID != "" || role.CatalogRevision <= 0 {
				return PrincipalGrants{}, fmt.Errorf("customer bundle requires a catalog revision and a non-workload subject")
			}
			if _, err := resolveTarget(db, scope); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					missing, checkErr := bindingTargetMissing(db, scope)
					if checkErr != nil {
						return PrincipalGrants{}, checkErr
					}
					if missing {
						continue
					}
				}
				return PrincipalGrants{}, err
			}
		case permissions.PolicyKindResourceRules:
			if managed.AccountID == "" || managed.PrivateRoleID == nil || *managed.PrivateRoleID != role.ID || role.Managed || len(role.Contexts) != 0 {
				return PrincipalGrants{}, fmt.Errorf("resource policy requires its managed account's private role")
			}
			if err := validateManagedGrants(db, managed, grants); err != nil {
				return PrincipalGrants{}, err
			}
		}
		if result.API != "" && result.API != policy.API() {
			return PrincipalGrants{}, fmt.Errorf("account has scoped roles for different APIs")
		}
		result.API = policy.API()
		result.Grants = append(result.Grants, grants...)
	}
	if result.API == "" {
		return PrincipalGrants{}, fmt.Errorf("account has no valid role binding in the selected tenant")
	}
	return result, nil
}

func bindingTargetMissing(db *gorm.DB, scope permissions.Scope) (bool, error) {
	var query *gorm.DB
	switch scope.Kind {
	case permissions.KindApp:
		query = db.Model(&app.App{}).Where(app.App{ID: scope.ID})
	case permissions.KindInstall:
		query = db.Model(&app.Install{}).Where(app.Install{ID: scope.ID})
	default:
		return false, nil
	}
	var target struct {
		OrgID     string
		DeletedAt soft_delete.DeletedAt
	}
	err := query.Unscoped().Scopes(scopes.WithDisableViews).Select("org_id", "deleted_at").Take(&target).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("check missing binding target: %w", err)
	}
	if target.OrgID != scope.OrgID {
		return false, fmt.Errorf("binding target belongs to a different tenant")
	}
	return target.DeletedAt != 0, nil
}

func validateManagedGrants(db *gorm.DB, managed app.ManagedServiceAccount, grants []permissions.ScopedPermissions) error {
	var owner permissions.ResolvedTarget
	var err error
	switch managed.Purpose {
	case app.ManagedServiceAccountPurposeStack, app.ManagedServiceAccountPurposePhoneHome:
		if managed.OwnerType != plugins.TableName(db, app.InstallStack{}) {
			return fmt.Errorf("stack identity requires a stack owner")
		}
		owner, err = resolveStackOwner(db, managed.OrgID, managed.OwnerID)
	case app.ManagedServiceAccountPurposeTelemetryCollector:
		if managed.OwnerType != plugins.TableName(db, app.Install{}) {
			return fmt.Errorf("install telemetry identity requires an install owner")
		}
		owner, err = resolveTarget(db, permissions.Scope{Kind: permissions.KindInstall, OrgID: managed.OrgID, ID: managed.OwnerID})
	default:
		return fmt.Errorf("ownership validation does not support managed purpose %q", managed.Purpose)
	}
	if err != nil {
		return err
	}
	for _, grant := range grants {
		target, err := resolveTarget(db, grant.Scope)
		if err != nil {
			return err
		}
		if target.InstallID != owner.InstallID || target.AppID != owner.AppID {
			return fmt.Errorf("private rule is outside the managed account's owner")
		}
		switch managed.Purpose {
		case app.ManagedServiceAccountPurposeStack:
			if target.Kind != permissions.KindStack {
				return fmt.Errorf("stack identity requires a stack rule")
			}
		case app.ManagedServiceAccountPurposeTelemetryCollector:
			if target.Kind != permissions.KindTelemetry {
				return fmt.Errorf("telemetry identity requires a telemetry rule")
			}
		case app.ManagedServiceAccountPurposePhoneHome:
			if target.Kind != permissions.KindPhoneHome || target.ID != managed.InstanceKey {
				return fmt.Errorf("phone-home identity requires its exact stack version")
			}
			var version struct{ ID string }
			if err := db.Model(&app.InstallStackVersion{}).Select("id").Where(app.InstallStackVersion{
				ID: target.ID, OrgID: managed.OrgID, InstallStackID: managed.OwnerID,
			}).Take(&version).Error; err != nil {
				return fmt.Errorf("phone-home version is not owned by the managed stack: %w", err)
			}
		}
	}
	return nil
}
