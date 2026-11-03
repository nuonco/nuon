package service

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) authorizeTrustPolicy(tx *gorm.DB, policy *app.OIDCTrustPolicy, reauthorize bool) (*app.Role, error) {
	ctx := tx.Statement.Context
	credential, ok := cctx.VerifiedCredentialFromContext(ctx)
	if !ok {
		return nil, genericExchangeError()
	}
	var source app.Token
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where(app.Token{ID: credential.TokenID, AccountID: credential.AccountID}).Take(&source).Error; err != nil {
		return nil, err
	}
	if source.TokenType == app.TokenTypeFederated {
		return nil, stderr.ErrAuthorization{Err: fmt.Errorf("federated credentials cannot delegate federation trust"), Description: "federated credentials cannot manage federation trust"}
	}
	client := authz.New(authz.Params{DB: tx})
	role, err := client.ResolveDelegableRoleAtScope(ctx, policy.Scope, app.RoleType(policy.Role), app.RoleContextTrustPolicy)
	if err != nil {
		return nil, err
	}
	document, err := permissions.DecodePolicy(role.Version, role.Policies[0].Document)
	if err != nil {
		return nil, err
	}
	ceiling, err := document.Compile(policy.OrgID, policy.Scope)
	if err != nil {
		return nil, err
	}
	if !reauthorize {
		approved, err := policyLimits(policy)
		if err != nil {
			return nil, err
		}
		target, err := client.ResolveTarget(ctx, policy.Scope)
		if err != nil {
			return nil, err
		}
		if !permissions.ContainsGrant(ceiling, target, approved.Ceiling[0].Permissions) {
			return nil, stderr.NewInvalidRequest(fmt.Errorf("trust policy requires explicit role and scope reauthorization"))
		}
		return role, nil
	}
	policy.Authorization, err = (permissions.CredentialLimits{
		Profile: permissions.ProfileServiceAccount, API: permissions.APICtl, Ceiling: ceiling,
	}).MarshalJSON()
	return role, err
}

func (s *service) lockTrustPolicy(tx *gorm.DB, orgID, policyID string) (*app.OIDCTrustPolicy, error) {
	if policyID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	if _, err := authz.New(authz.Params{DB: tx.Clauses(clause.Locking{Strength: "UPDATE"})}).ResolveTarget(tx.Statement.Context,
		permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}); err != nil {
		return nil, err
	}
	var policy app.OIDCTrustPolicy
	if err := tx.Where(app.OIDCTrustPolicy{ID: policyID, OrgID: orgID}).Take(&policy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, stderr.ErrNotFound{Err: fmt.Errorf("trust policy not found")}
		}
		return nil, err
	}
	if policy.ServiceAccountID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var subject app.Account
	if err := tx.Unscoped().Select("id").Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(app.Account{ID: policy.ServiceAccountID}).Find(&subject).Error; err != nil {
		return nil, err
	}
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.OIDCTrustPolicy{ID: policyID, OrgID: orgID}).Take(&policy).Error; err != nil {
		return nil, err
	}
	return &policy, nil
}

func policyLimits(policy *app.OIDCTrustPolicy) (permissions.CredentialLimits, error) {
	limits, err := permissions.DecodeCredentialLimits(policy.Authorization)
	if err != nil {
		return permissions.CredentialLimits{}, err
	}
	if _, err := authz.RoleBindingScope(policy.OrgID, &policy.Scope); err != nil {
		return permissions.CredentialLimits{}, err
	}
	if limits.Profile != permissions.ProfileServiceAccount || limits.API != permissions.APICtl || len(limits.Ceiling) != 1 ||
		limits.Ceiling[0].Scope != policy.Scope || len(limits.Ceiling[0].Permissions) == 0 {
		return permissions.CredentialLimits{}, fmt.Errorf("trust policy requires approved customer credential limits")
	}
	return limits, nil
}

func (s *service) federatedLimits(tx *gorm.DB, policy *app.OIDCTrustPolicy) (permissions.CredentialLimits, error) {
	limits, err := policyLimits(policy)
	if err != nil {
		return permissions.CredentialLimits{}, err
	}
	if err := requireFederationSubject(tx, policy.ServiceAccountID); err != nil {
		return permissions.CredentialLimits{}, err
	}
	client := authz.New(authz.Params{DB: tx})
	grants, err := client.LoadScopedGrants(tx.Statement.Context, policy.ServiceAccountID, policy.OrgID)
	if err != nil {
		return permissions.CredentialLimits{}, err
	}
	if grants.API != permissions.APICtl {
		return permissions.CredentialLimits{}, fmt.Errorf("federation requires customer subject authority")
	}
	candidates := append([]permissions.ScopedPermissions{}, limits.Ceiling...)
	candidates = append(candidates, grants.Grants...)
	ceiling := make([]permissions.ScopedPermissions, 0, len(candidates))
	seen := make(map[permissions.Scope]bool)
	for _, candidate := range candidates {
		if seen[candidate.Scope] {
			continue
		}
		seen[candidate.Scope] = true
		target, err := authz.New(authz.Params{DB: tx.Clauses(clause.Locking{Strength: "SHARE"})}).ResolveTarget(tx.Statement.Context, candidate.Scope)
		if err != nil {
			return permissions.CredentialLimits{}, err
		}
		allowed := make(permissions.PermissionSet)
		for permission := range limits.Ceiling[0].Permissions {
			requested, err := permissions.NewPermissionSet(permission)
			if err != nil {
				return permissions.CredentialLimits{}, err
			}
			if permissions.ContainsGrant(limits.Ceiling, target, requested) && permissions.ContainsGrant(grants.Grants, target, requested) {
				allowed[permission] = struct{}{}
			}
		}
		if len(allowed) != 0 {
			ceiling = append(ceiling, permissions.ScopedPermissions{Scope: candidate.Scope, Permissions: allowed})
		}
	}
	if len(ceiling) == 0 {
		return permissions.CredentialLimits{}, fmt.Errorf("subject has no authority within the approved federation limits")
	}
	limits.Ceiling = ceiling
	return limits, nil
}

func requireFederationSubject(tx *gorm.DB, accountID string) error {
	if err := authz.RequireUserManaged(tx, accountID); err != nil {
		return err
	}
	return tx.Select("id").Where(app.Account{ID: accountID, AccountType: app.AccountTypeService}).Take(&app.Account{}).Error
}
