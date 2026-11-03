package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcverify"
)

type CreateOIDCTrustPolicyRequest struct {
	// human-friendly name to identify the policy
	Name string `json:"name" validate:"required"`

	// exact `iss` claim value; also used for OIDC discovery + JWKS fetching
	IssuerURL string `json:"issuer_url" validate:"required"`

	// expected `aud` claim value
	Audience string `json:"audience" validate:"required"`

	// map of claim name -> pattern; all must match. A `sub` condition is
	// required. Patterns are exact strings or globs where `*` cannot cross
	// `:` segments.
	ClaimConditions map[string]string `json:"claim_conditions" validate:"required"`

	Role  string             `json:"role" validate:"required"`
	Scope *permissions.Scope `json:"scope" validate:"required"`

	// lifetime of exchanged tokens in seconds. defaults to 3600, max 86400.
	TokenDurationSeconds int `json:"token_duration_seconds"`
}

type UpdateOIDCTrustPolicyRequest struct {
	Name                 string             `json:"name"`
	Enabled              *bool              `json:"enabled" swaggertype:"boolean" extensions:"x-nullable"`
	IssuerURL            string             `json:"issuer_url"`
	Audience             string             `json:"audience"`
	ClaimConditions      map[string]string  `json:"claim_conditions"`
	Role                 string             `json:"role"`
	Scope                *permissions.Scope `json:"scope"`
	TokenDurationSeconds int                `json:"token_duration_seconds"`
}

func parseTrustPolicyTokenDuration(raw int) (int, error) {
	if raw == 0 {
		return app.OIDCTrustPolicyDefaultTokenDuration, nil
	}
	if raw < 0 {
		return 0, fmt.Errorf("token duration must be positive")
	}
	if raw > app.OIDCTrustPolicyMaxTokenDuration {
		return 0, fmt.Errorf("token duration cannot exceed %d seconds", app.OIDCTrustPolicyMaxTokenDuration)
	}

	return raw, nil
}

func (s *service) validateIssuerURL(ctx context.Context, raw string) error {
	return oidcverify.ValidateIssuerURL(ctx, raw, s.cfg.OIDCFederationAllowInsecureIssuers)
}

// @ID						CreateOIDCTrustPolicy
// @Summary				create an OIDC trust policy
// @Description			Creates an OIDC workload identity trust policy for your current org. OIDC tokens matching the policy's issuer, audience, and claim conditions can be exchanged for short-lived Nuon API tokens. Each policy gets a dedicated service account with the configured role.
// @Param					req	body	CreateOIDCTrustPolicyRequest	true	"Input"
// @Tags					oidc_federation
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Produce				json
// @Success				201	{object}	app.OIDCTrustPolicy
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Router					/v1/oidc/trust-policies [POST]
func (s *service) CreateOIDCTrustPolicy(ctx *gin.Context) (require.PreparedMutation, error) {
	var req CreateOIDCTrustPolicyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return require.PreparedMutation{}, stderr.NewInvalidRequest(err)
	}
	if err := s.v.Struct(req); err != nil {
		return require.PreparedMutation{}, stderr.NewInvalidRequest(err)
	}

	duration, err := parseTrustPolicyTokenDuration(req.TokenDurationSeconds)
	if err != nil {
		return require.PreparedMutation{}, stderr.NewInvalidRequest(err)
	}

	org, err := s.requirePolicyAdministration(ctx)
	if err != nil {
		return require.PreparedMutation{}, err
	}

	if err := s.validateIssuerURL(ctx, req.IssuerURL); err != nil {
		return require.PreparedMutation{}, stderr.NewInvalidRequest(err)
	}

	if err := s.verifier.Discover(ctx, req.IssuerURL); err != nil {
		return require.PreparedMutation{}, stderr.NewInvalidRequest(fmt.Errorf("issuer discovery failed: %w", err))
	}

	if err := validateClaimConditions(req.ClaimConditions); err != nil {
		return require.PreparedMutation{}, stderr.NewInvalidRequest(err)
	}

	scope, err := authz.RoleBindingScope(org.ID, req.Scope)
	if err != nil {
		return require.PreparedMutation{}, stderr.NewInvalidRequest(err)
	}

	caller, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return require.PreparedMutation{}, err
	}

	return require.PreparedMutation{Mutate: func(ctx *gin.Context) (require.Response, error) {
		access, err := require.MutationAccessFromContext(ctx)
		if err != nil {
			return require.Response{}, err
		}
		tx := access.DB()
		policy := app.OIDCTrustPolicy{
			CreatedByID:          caller.ID,
			OrgID:                access.Target().ID,
			Name:                 req.Name,
			Enabled:              true,
			IssuerURL:            req.IssuerURL,
			Audience:             req.Audience,
			ClaimConditions:      req.ClaimConditions,
			Role:                 req.Role,
			Scope:                scope,
			TokenDurationSeconds: duration,
		}
		role, err := s.authorizeTrustPolicy(tx, &policy, true)
		if err != nil {
			return require.Response{}, err
		}
		acct, err := s.createPolicyServiceAccount(tx, scope, role.ID)
		if err != nil {
			return require.Response{}, err
		}
		policy.ServiceAccountID = acct.ID
		if err := tx.Create(&policy).Error; err != nil {
			return require.Response{}, fmt.Errorf("unable to create trust policy: %w", err)
		}
		return require.Response{Status: http.StatusCreated, Body: policy}, nil
	}}, nil
}

// @ID						ListOIDCTrustPolicies
// @Summary				list your org's OIDC trust policies
// @Description			Lists the OIDC workload identity trust policies for your current org.
// @Tags					oidc_federation
// @Security				APIKey
// @Security				OrgID
// @Produce				json
// @Success				200	{array}	app.OIDCTrustPolicy
// @Failure				403	{object}	stderr.ErrResponse
// @Router					/v1/oidc/trust-policies [GET]
func (s *service) ListOIDCTrustPolicies(ctx *gin.Context) (any, error) {
	access, err := require.ReadAccessFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireFederationEnabled(); err != nil {
		return nil, err
	}

	var policies []app.OIDCTrustPolicy
	res := access.DB().
		Where(app.OIDCTrustPolicy{OrgID: access.OrgID()}).
		Order("created_at DESC, id DESC").
		Find(&policies)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to list trust policies: %w", res.Error)
	}

	return policies, nil
}

// @ID						GetOIDCTrustPolicy
// @Summary				get an OIDC trust policy
// @Description			Returns an OIDC workload identity trust policy belonging to your current org.
// @Param					policy_id	path	string	true	"policy ID"
// @Tags					oidc_federation
// @Security				APIKey
// @Security				OrgID
// @Produce				json
// @Success				200	{object}	app.OIDCTrustPolicy
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Router					/v1/oidc/trust-policies/{policy_id} [GET]
func (s *service) GetOIDCTrustPolicy(ctx *gin.Context) (any, error) {
	access, err := require.ReadAccessFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireFederationEnabled(); err != nil {
		return nil, err
	}

	var policy app.OIDCTrustPolicy
	err = access.DB().Where(app.OIDCTrustPolicy{ID: ctx.Param("policy_id"), OrgID: access.OrgID()}).Take(&policy).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, stderr.ErrNotFound{Err: fmt.Errorf("trust policy not found")}
	}
	if err != nil {
		return nil, fmt.Errorf("unable to look up trust policy: %w", err)
	}
	return &policy, nil
}

// @ID						UpdateOIDCTrustPolicy
// @Summary				update an OIDC trust policy
// @Description			Updates an OIDC workload identity trust policy belonging to your current org. Changing the role also updates the policy's service account role, which affects tokens already issued under the policy. Disabling the policy revokes tokens already issued under it.
// @Param					policy_id	path	string							true	"policy ID"
// @Param					req			body	UpdateOIDCTrustPolicyRequest	true	"Input"
// @Tags					oidc_federation
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Produce				json
// @Success				200	{object}	app.OIDCTrustPolicy
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Router					/v1/oidc/trust-policies/{policy_id} [PATCH]
func (s *service) UpdateOIDCTrustPolicy(ctx *gin.Context) (require.PreparedMutation, error) {
	var req UpdateOIDCTrustPolicyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return require.PreparedMutation{}, stderr.NewInvalidRequest(err)
	}

	if _, err := s.requirePolicyAdministration(ctx); err != nil {
		return require.PreparedMutation{}, err
	}

	if req.IssuerURL != "" {
		if err := s.validateIssuerURL(ctx, req.IssuerURL); err != nil {
			return require.PreparedMutation{}, stderr.NewInvalidRequest(err)
		}
		if err := s.verifier.Discover(ctx, req.IssuerURL); err != nil {
			return require.PreparedMutation{}, stderr.NewInvalidRequest(fmt.Errorf("issuer discovery failed: %w", err))
		}
	}
	if req.ClaimConditions != nil {
		if err := validateClaimConditions(req.ClaimConditions); err != nil {
			return require.PreparedMutation{}, stderr.NewInvalidRequest(err)
		}
	}
	if req.TokenDurationSeconds != 0 {
		if _, err := parseTrustPolicyTokenDuration(req.TokenDurationSeconds); err != nil {
			return require.PreparedMutation{}, stderr.NewInvalidRequest(err)
		}
	}

	return require.PreparedMutation{Mutate: func(ctx *gin.Context) (require.Response, error) {
		access, err := require.MutationAccessFromContext(ctx)
		if err != nil {
			return require.Response{}, err
		}
		tx := access.DB()
		orgID := access.Target().ID
		policy, err := s.lockTrustPolicy(tx, orgID, ctx.Param("policy_id"))
		if err != nil {
			return require.Response{}, err
		}
		oldScope, oldRole := policy.Scope, policy.Role
		disableOnly := req.Enabled != nil && !*req.Enabled && req.Name == "" && req.IssuerURL == "" && req.Audience == "" && req.ClaimConditions == nil && req.Role == "" && req.Scope == nil && req.TokenDurationSeconds == 0
		if !disableOnly {
			if _, err := policyLimits(policy); err != nil && (req.Role == "" || req.Scope == nil) {
				return require.Response{}, stderr.NewInvalidRequest(fmt.Errorf("trust policy requires explicit role and scope reauthorization"))
			}
			if req.Role != "" {
				policy.Role = req.Role
			}
			if req.Scope != nil {
				policy.Scope, err = authz.RoleBindingScope(orgID, req.Scope)
				if err != nil {
					return require.Response{}, stderr.NewInvalidRequest(err)
				}
			}
			role, err := s.authorizeTrustPolicy(tx, policy, req.Role != "" || req.Scope != nil)
			if err != nil {
				return require.Response{}, err
			}
			if oldRole != policy.Role || oldScope != policy.Scope {
				if err := tx.Unscoped().Where(app.AccountRole{
					OrgID: generics.NewNullString(orgID), AccountID: policy.ServiceAccountID,
					ScopeKind: oldScope.Kind, ScopeID: oldScope.ID,
				}).Delete(&app.AccountRole{}).Error; err != nil {
					return require.Response{}, err
				}
				if err := authz.New(authz.Params{DB: tx}).SetAccountRoleAtScope(tx.Statement.Context, policy.ServiceAccountID, role.ID, policy.Scope); err != nil {
					return require.Response{}, err
				}
			}
			if _, err := s.federatedLimits(tx, policy); err != nil {
				return require.Response{}, err
			}
		}
		if req.Name != "" {
			policy.Name = req.Name
		}
		if req.Enabled != nil {
			policy.Enabled = *req.Enabled
		}
		if req.IssuerURL != "" {
			policy.IssuerURL = req.IssuerURL
		}
		if req.Audience != "" {
			policy.Audience = req.Audience
		}
		if req.ClaimConditions != nil {
			policy.ClaimConditions = req.ClaimConditions
		}
		if req.TokenDurationSeconds != 0 {
			policy.TokenDurationSeconds = req.TokenDurationSeconds
		}
		if !policy.Enabled || req.IssuerURL != "" || req.Audience != "" || req.ClaimConditions != nil || req.Role != "" || req.Scope != nil {
			if err := account.RevokeDerivedCredentials(tx, app.TokenSourceTypeOIDCTrustPolicy, []string{policy.ID}); err != nil {
				return require.Response{}, fmt.Errorf("unable to revoke tokens issued under this policy: %w", err)
			}
		}
		if err := tx.Save(policy).Error; err != nil {
			return require.Response{}, fmt.Errorf("unable to update trust policy: %w", err)
		}
		return require.Response{Status: http.StatusOK, Body: policy}, nil
	}}, nil
}

// @ID						DeleteOIDCTrustPolicy
// @Summary				delete an OIDC trust policy
// @Description			Deletes an OIDC workload identity trust policy belonging to your current org, along with its dedicated service account. Tokens already issued under the policy stop working immediately.
// @Param					policy_id	path	string	true	"policy ID"
// @Tags					oidc_federation
// @Security				APIKey
// @Security				OrgID
// @Produce				json
// @Success				204
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Router					/v1/oidc/trust-policies/{policy_id} [DELETE]
func (s *service) DeleteOIDCTrustPolicy(ctx *gin.Context) (require.PreparedMutation, error) {
	if _, err := s.requirePolicyAdministration(ctx); err != nil {
		return require.PreparedMutation{}, err
	}

	return require.PreparedMutation{Mutate: func(ctx *gin.Context) (require.Response, error) {
		access, err := require.MutationAccessFromContext(ctx)
		if err != nil {
			return require.Response{}, err
		}
		tx := access.DB()
		orgID := access.Target().ID
		policy, err := s.lockTrustPolicy(tx, orgID, ctx.Param("policy_id"))
		if err != nil {
			return require.Response{}, err
		}
		if err := tx.Delete(policy).Error; err != nil {
			return require.Response{}, err
		}
		if err := account.RevokeDerivedCredentials(tx, app.TokenSourceTypeOIDCTrustPolicy, []string{policy.ID}); err != nil {
			return require.Response{}, fmt.Errorf("unable to revoke tokens issued under this policy: %w", err)
		}
		if err := s.deletePolicyServiceAccount(tx, orgID, policy.ServiceAccountID); err != nil {
			return require.Response{}, fmt.Errorf("unable to delete trust policy: %w", err)
		}
		return require.Response{Status: http.StatusNoContent}, nil
	}}, nil
}

// requireFederationEnabled gates the trust policy CRUD surface behind the
// deployment-level config flag. The exchange endpoint checks the flag
// separately and returns its uniform generic 401 instead.
func (s *service) requireFederationEnabled() error {
	if s.cfg.OIDCFederationEnabled {
		return nil
	}

	return stderr.ErrUser{
		Err:         fmt.Errorf("OIDC federation is not enabled"),
		Description: "OIDC federation is not enabled on this control plane; set oidc_federation_enabled in the API configuration",
	}
}

func (s *service) requirePolicyAdministration(ctx *gin.Context) (*app.Org, error) {
	if err := s.requireFederationEnabled(); err != nil {
		return nil, err
	}

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := require.CustomerRead(ctx, s.db, s.l, authz.ReadRequest{
		Permission: permissions.OrgWrite,
		Target:     permissions.Scope{Kind: permissions.KindOrg, OrgID: org.ID, ID: org.ID},
	}); err != nil {
		return nil, err
	}

	return org, nil
}

func (s *service) createPolicyServiceAccount(tx *gorm.DB, scope permissions.Scope, roleID string) (*app.Account, error) {
	name := fmt.Sprintf("%s-oidc-%s", scope.OrgID, domains.NewAccountID())

	newAcct := app.Account{
		Email:       account.ServiceAccountEmail(name),
		Subject:     name,
		AccountType: app.AccountTypeService,
	}
	if err := tx.Create(&newAcct).Error; err != nil {
		return nil, fmt.Errorf("unable to create service account: %w", err)
	}

	if err := authz.New(authz.Params{DB: tx}).AddAccountRoleAtScope(tx.Statement.Context, newAcct.ID, roleID, scope); err != nil {
		return nil, fmt.Errorf("unable to add org role to service account: %w", err)
	}

	return &newAcct, nil
}

func (s *service) deletePolicyServiceAccount(tx *gorm.DB, orgID, accountID string) error {
	if accountID == "" {
		return gorm.ErrRecordNotFound
	}
	var acct app.Account
	if err := tx.Unscoped().Select("id").Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(app.Account{ID: accountID, AccountType: app.AccountTypeService}).Take(&acct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to lock policy service account: %w", err)
	}
	if err := requireFederationSubject(tx.Unscoped().Session(&gorm.Session{}), accountID); err != nil {
		return err
	}
	return account.New(account.Params{DB: tx}).DeleteServiceAccountByID(tx.Statement.Context, accountID)
}
