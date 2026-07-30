package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const defaultTrustPolicyRole = app.RoleTypeOrgReadOnly

var allowedTrustPolicyRoles = map[app.RoleType]struct{}{
	app.RoleTypeOrgAdmin:    {},
	app.RoleTypeOrgSupport:  {},
	app.RoleTypeOrgReadOnly: {},
}

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

	// org role granted to exchanged tokens. one of org_admin, org_support,
	// org_read_only. defaults to org_read_only.
	Role string `json:"role"`

	// lifetime of exchanged tokens in seconds. defaults to 3600, max 86400.
	TokenDurationSeconds int `json:"token_duration_seconds"`
}

type UpdateOIDCTrustPolicyRequest struct {
	Name                 string            `json:"name"`
	Enabled              *bool             `json:"enabled" swaggertype:"boolean" extensions:"x-nullable"`
	IssuerURL            string            `json:"issuer_url"`
	Audience             string            `json:"audience"`
	ClaimConditions      map[string]string `json:"claim_conditions"`
	Role                 string            `json:"role"`
	TokenDurationSeconds int               `json:"token_duration_seconds"`
}

func parseTrustPolicyRole(raw string) (app.RoleType, error) {
	if raw == "" {
		return defaultTrustPolicyRole, nil
	}

	role := app.RoleType(raw)
	if _, ok := allowedTrustPolicyRoles[role]; !ok {
		return "", fmt.Errorf("invalid role %q: must be one of %q, %q, %q", raw, app.RoleTypeOrgAdmin, app.RoleTypeOrgSupport, app.RoleTypeOrgReadOnly)
	}

	return role, nil
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

func (s *service) validateIssuerURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid issuer URL: %w", err)
	}

	if parsed.Host == "" {
		return fmt.Errorf("issuer URL must include a host")
	}

	switch parsed.Scheme {
	case "https":
		return nil
	case "http":
		if s.cfg.OIDCFederationAllowInsecureIssuers {
			return nil
		}
		return fmt.Errorf("issuer URL must use https")
	default:
		return fmt.Errorf("issuer URL must use https")
	}
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
func (s *service) CreateOIDCTrustPolicy(ctx *gin.Context) {
	var req CreateOIDCTrustPolicyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := s.v.Struct(req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	role, err := parseTrustPolicyRole(req.Role)
	if err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	duration, err := parseTrustPolicyTokenDuration(req.TokenDurationSeconds)
	if err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if err := s.validateIssuerURL(req.IssuerURL); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if err := validateClaimConditions(req.ClaimConditions); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	caller, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	acct, err := s.createPolicyServiceAccount(ctx, org.ID, role)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create service account: %w", err))
		return
	}

	policy := app.OIDCTrustPolicy{
		CreatedByID:          caller.ID,
		OrgID:                org.ID,
		Name:                 req.Name,
		Enabled:              true,
		IssuerURL:            req.IssuerURL,
		Audience:             req.Audience,
		ClaimConditions:      req.ClaimConditions,
		Role:                 string(role),
		TokenDurationSeconds: duration,
		ServiceAccountID:     acct.ID,
	}

	if res := s.db.WithContext(ctx).Create(&policy); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create trust policy: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusCreated, policy)
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
func (s *service) ListOIDCTrustPolicies(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var policies []app.OIDCTrustPolicy
	res := s.db.WithContext(ctx).
		Where(app.OIDCTrustPolicy{OrgID: org.ID}).
		Order("created_at DESC").
		Find(&policies)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list trust policies: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, policies)
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
func (s *service) GetOIDCTrustPolicy(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	policy, err := s.getTrustPolicy(ctx, org.ID, ctx.Param("policy_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, policy)
}

// @ID						UpdateOIDCTrustPolicy
// @Summary				update an OIDC trust policy
// @Description			Updates an OIDC workload identity trust policy belonging to your current org. Changing the role also updates the policy's service account role, which affects tokens already issued under the policy.
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
func (s *service) UpdateOIDCTrustPolicy(ctx *gin.Context) {
	var req UpdateOIDCTrustPolicyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	policy, err := s.getTrustPolicy(ctx, org.ID, ctx.Param("policy_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	if req.Name != "" {
		policy.Name = req.Name
	}
	if req.Enabled != nil {
		policy.Enabled = *req.Enabled
	}
	if req.IssuerURL != "" {
		if err := s.validateIssuerURL(req.IssuerURL); err != nil {
			ctx.Error(stderr.NewInvalidRequest(err))
			return
		}
		policy.IssuerURL = req.IssuerURL
	}
	if req.Audience != "" {
		policy.Audience = req.Audience
	}
	if req.ClaimConditions != nil {
		if err := validateClaimConditions(req.ClaimConditions); err != nil {
			ctx.Error(stderr.NewInvalidRequest(err))
			return
		}
		policy.ClaimConditions = req.ClaimConditions
	}
	if req.TokenDurationSeconds != 0 {
		duration, err := parseTrustPolicyTokenDuration(req.TokenDurationSeconds)
		if err != nil {
			ctx.Error(stderr.NewInvalidRequest(err))
			return
		}
		policy.TokenDurationSeconds = duration
	}

	if req.Role != "" && req.Role != policy.Role {
		role, err := parseTrustPolicyRole(req.Role)
		if err != nil {
			ctx.Error(stderr.NewInvalidRequest(err))
			return
		}

		if err := s.authzClient.SetAccountOrgRole(ctx, org.ID, policy.ServiceAccountID, role); err != nil {
			ctx.Error(fmt.Errorf("unable to update service account role: %w", err))
			return
		}
		policy.Role = string(role)
	}

	if res := s.db.WithContext(ctx).Save(policy); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to update trust policy: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, policy)
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
func (s *service) DeleteOIDCTrustPolicy(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	policy, err := s.getTrustPolicy(ctx, org.ID, ctx.Param("policy_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.db.WithContext(ctx).Delete(policy).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to delete trust policy: %w", err))
		return
	}

	if err := s.deletePolicyServiceAccount(ctx, org.ID, policy.ServiceAccountID); err != nil {
		ctx.Error(fmt.Errorf("unable to delete service account: %w", err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *service) getTrustPolicy(ctx *gin.Context, orgID, policyID string) (*app.OIDCTrustPolicy, error) {
	var policy app.OIDCTrustPolicy
	res := s.db.WithContext(ctx).
		Where(app.OIDCTrustPolicy{
			ID:    policyID,
			OrgID: orgID,
		}).
		First(&policy)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, stderr.ErrNotFound{Err: fmt.Errorf("trust policy not found")}
	}
	if res.Error != nil {
		return nil, fmt.Errorf("unable to look up trust policy: %w", res.Error)
	}

	return &policy, nil
}

func (s *service) requireOrgAdmin(ctx *gin.Context) (*app.Org, error) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return nil, err
	}

	if !s.isOrgAdmin(acct, org.ID) {
		return nil, stderr.ErrAuthorization{
			Err:         fmt.Errorf("only org admins can manage OIDC trust policies"),
			Description: "only org admins can manage OIDC trust policies",
		}
	}

	return org, nil
}

func (s *service) isOrgAdmin(acct *app.Account, orgID string) bool {
	for _, role := range acct.Roles {
		if role.RoleType == app.RoleTypeOrgAdmin && role.OrgID.ValueString() == orgID {
			return true
		}
	}
	return false
}

func (s *service) createPolicyServiceAccount(ctx context.Context, orgID string, roleType app.RoleType) (*app.Account, error) {
	name := fmt.Sprintf("%s-oidc-%s", orgID, domains.NewAccountID())
	email := account.ServiceAccountEmail(name)

	newAcct := app.Account{
		Email:       email,
		Subject:     name,
		AccountType: app.AccountTypeService,
	}
	if res := s.db.WithContext(ctx).Create(&newAcct); res.Error != nil {
		return nil, fmt.Errorf("unable to create service account: %w", res.Error)
	}

	if err := s.authzClient.AddAccountOrgRole(ctx, roleType, orgID, newAcct.ID); err != nil {
		return nil, fmt.Errorf("unable to add org role to service account: %w", err)
	}

	return &newAcct, nil
}

func (s *service) deletePolicyServiceAccount(ctx context.Context, orgID, accountID string) error {
	if err := s.authzClient.RemoveAccountOrgRoles(ctx, orgID, accountID); err != nil {
		return fmt.Errorf("unable to remove service account roles: %w", err)
	}

	if err := s.db.WithContext(ctx).
		Where(app.Account{ID: accountID, AccountType: app.AccountTypeService}).
		Delete(&app.Account{}).Error; err != nil {
		return fmt.Errorf("unable to delete service account: %w", err)
	}

	return nil
}
