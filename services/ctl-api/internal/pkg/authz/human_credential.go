package authz

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (c *Client) HumanCredentialLimits(ctx context.Context, accountID string) (permissions.CredentialLimits, error) {
	limits := permissions.CredentialLimits{Profile: permissions.ProfileHuman, API: permissions.APICtl, Ceiling: []permissions.ScopedPermissions{}}
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		account, err := lockRoleSubject(tx, accountID)
		if err != nil {
			return err
		}
		if account.AccountType != app.AccountTypeAuth && account.AccountType != app.AccountTypeAuth0 {
			return fmt.Errorf("human credential requires a human account")
		}
		var bindings []app.AccountRole
		if err := tx.Where(app.AccountRole{AccountID: accountID}).Find(&bindings).Error; err != nil {
			return err
		}
		var orgIDs []string
		for _, binding := range bindings {
			if binding.OrgID.Empty() {
				return fmt.Errorf("human credential requires tenant-owned grants")
			}
			orgIDs = append(orgIDs, binding.OrgID.ValueString())
		}
		slices.Sort(orgIDs)
		for _, orgID := range slices.Compact(orgIDs) {
			missing, err := bindingTargetMissing(tx, permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID})
			if err != nil {
				return err
			}
			if missing {
				continue
			}
			grants, err := loadScopedGrants(tx, accountID, orgID)
			if errors.Is(err, errNoScopedGrants) {
				continue
			}
			if err != nil {
				return err
			}
			if grants.API != permissions.APICtl {
				return fmt.Errorf("human credential requires customer API grants")
			}
			limits.Ceiling = append(limits.Ceiling, grants.Grants...)
		}
		return limits.Validate()
	})
	return limits, err
}

func (c *Client) ConstrainHumanCredential(ctx context.Context, accountID string, ceiling permissions.CredentialLimits) (permissions.CredentialLimits, error) {
	result := permissions.CredentialLimits{Profile: permissions.ProfileHuman, API: permissions.APICtl, Ceiling: []permissions.ScopedPermissions{}}
	if err := ceiling.Validate(); err != nil {
		return result, err
	}
	if ceiling.Profile != result.Profile || ceiling.API != result.API {
		return result, fmt.Errorf("human credential requires human customer API limits")
	}
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		live, err := New(Params{DB: tx}).HumanCredentialLimits(ctx, accountID)
		if err != nil {
			return err
		}
		seen := make(map[permissions.Scope]bool)
		for _, grant := range append(slices.Clone(live.Ceiling), ceiling.Ceiling...) {
			if seen[grant.Scope] {
				continue
			}
			seen[grant.Scope] = true
			if err := validateBindingScope(grant.Scope); err != nil {
				return err
			}
			if !slices.ContainsFunc(live.Ceiling, func(current permissions.ScopedPermissions) bool {
				return current.Scope.OrgID == grant.Scope.OrgID
			}) {
				continue
			}
			target, err := resolveTarget(tx, grant.Scope)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				missing, checkErr := bindingTargetMissing(tx, grant.Scope)
				if checkErr != nil {
					return checkErr
				}
				if missing {
					continue
				}
			}
			if err != nil {
				return err
			}
			set := permissions.PermissionSet{}
			for _, allowed := range ceiling.Ceiling {
				for permission := range allowed.Permissions {
					one := permissions.PermissionSet{permission: {}}
					if permissions.ContainsGrant(live.Ceiling, target, one) && permissions.ContainsGrant(ceiling.Ceiling, target, one) {
						set[permission] = struct{}{}
					}
				}
			}
			if len(set) != 0 {
				result.Ceiling = append(result.Ceiling, permissions.ScopedPermissions{Scope: grant.Scope, Permissions: set})
			}
		}
		return result.Validate()
	})
	return result, err
}
