package service

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *OIDCFederationTestSuite) createPolicy(idp *fakeIDP, conditions map[string]string) app.OIDCTrustPolicy {
	s.deps.Service.cfg.OIDCFederationAllowInsecureIssuers = true

	rr := s.makeRequest(http.MethodPost, "/v1/oidc/trust-policies", CreateOIDCTrustPolicyRequest{
		Name:            "ci",
		IssuerURL:       idp.issuer(),
		Audience:        "nuon-test",
		ClaimConditions: conditions,
	})
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

	var policy app.OIDCTrustPolicy
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &policy))
	return policy
}

func (s *OIDCFederationTestSuite) TestFederationDisabled() {
	idp := newFakeIDP(s.T())
	policy := s.createPolicy(idp, map[string]string{"sub": "repo:acme/app:ref:refs/heads/main"})
	token := idp.signToken(s.T(), jwt.MapClaims{"sub": "repo:acme/app:ref:refs/heads/main"})

	s.deps.Service.cfg.OIDCFederationEnabled = false
	defer func() { s.deps.Service.cfg.OIDCFederationEnabled = true }()

	rr := s.makeRequest(http.MethodPost, "/v1/oidc/token", ExchangeOIDCTokenRequest{OrgID: s.testOrg.ID, Token: token})
	require.Equal(s.T(), http.StatusUnauthorized, rr.Code, rr.Body.String())

	rr = s.makeRequest(http.MethodGet, "/v1/oidc/trust-policies", nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	require.Contains(s.T(), rr.Body.String(), "not enabled")

	rr = s.makeRequest(http.MethodDelete, "/v1/oidc/trust-policies/"+policy.ID, nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
}

func (s *OIDCFederationTestSuite) TestCreatePolicyCreatesServiceAccount() {
	idp := newFakeIDP(s.T())
	policy := s.createPolicy(idp, map[string]string{"sub": "repo:acme/app:*:*"})

	require.Equal(s.T(), s.testOrg.ID, policy.OrgID)
	require.Equal(s.T(), string(app.RoleTypeOrgReadOnly), policy.Role)
	require.Equal(s.T(), app.OIDCTrustPolicyDefaultTokenDuration, policy.TokenDurationSeconds)
	require.True(s.T(), policy.Enabled)
	require.NotEmpty(s.T(), policy.ServiceAccountID)

	var acct app.Account
	require.NoError(s.T(), s.deps.DB.Preload("Roles").Preload("Roles.Org").Where("id = ?", policy.ServiceAccountID).First(&acct).Error)
	require.Equal(s.T(), app.AccountTypeService, acct.AccountType)
	require.Len(s.T(), acct.Roles, 1)
	require.Equal(s.T(), app.RoleTypeOrgReadOnly, acct.Roles[0].RoleType)
}

func (s *OIDCFederationTestSuite) TestCreatePolicyValidation() {
	s.deps.Service.cfg.OIDCFederationAllowInsecureIssuers = false

	cases := []struct {
		name string
		req  CreateOIDCTrustPolicyRequest
	}{
		{
			name: "missing sub condition",
			req: CreateOIDCTrustPolicyRequest{
				Name: "ci", IssuerURL: "https://idp.example.com", Audience: "aud",
				ClaimConditions: map[string]string{"repository_owner": "acme"},
			},
		},
		{
			name: "http issuer rejected",
			req: CreateOIDCTrustPolicyRequest{
				Name: "ci", IssuerURL: "http://idp.example.com", Audience: "aud",
				ClaimConditions: map[string]string{"sub": "repo:acme/app:*:*"},
			},
		},
		{
			name: "invalid role",
			req: CreateOIDCTrustPolicyRequest{
				Name: "ci", IssuerURL: "https://idp.example.com", Audience: "aud",
				ClaimConditions: map[string]string{"sub": "repo:acme/app:*:*"},
				Role:            "runner",
			},
		},
		{
			name: "excessive token duration",
			req: CreateOIDCTrustPolicyRequest{
				Name: "ci", IssuerURL: "https://idp.example.com", Audience: "aud",
				ClaimConditions:      map[string]string{"sub": "repo:acme/app:*:*"},
				TokenDurationSeconds: app.OIDCTrustPolicyMaxTokenDuration + 1,
			},
		},
	}

	for _, tc := range cases {
		rr := s.makeRequest(http.MethodPost, "/v1/oidc/trust-policies", tc.req)
		require.Equal(s.T(), http.StatusBadRequest, rr.Code, "%s: %s", tc.name, rr.Body.String())
	}
}

func (s *OIDCFederationTestSuite) TestCreatePolicyAllowsAdminRole() {
	idp := newFakeIDP(s.T())
	s.deps.Service.cfg.OIDCFederationAllowInsecureIssuers = true

	rr := s.makeRequest(http.MethodPost, "/v1/oidc/trust-policies", CreateOIDCTrustPolicyRequest{
		Name: "admin", IssuerURL: idp.issuer(), Audience: "nuon-test",
		ClaimConditions: map[string]string{"sub": "repo:acme/app:*:*"},
		Role:            string(app.RoleTypeOrgAdmin),
	})
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

	var policy app.OIDCTrustPolicy
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &policy))
	require.Equal(s.T(), string(app.RoleTypeOrgAdmin), policy.Role)

	var acct app.Account
	require.NoError(s.T(), s.deps.DB.Preload("Roles").Preload("Roles.Org").Where(app.Account{ID: policy.ServiceAccountID}).First(&acct).Error)
	require.Len(s.T(), acct.Roles, 1)
	require.Equal(s.T(), app.RoleTypeOrgAdmin, acct.Roles[0].RoleType)
}

func (s *OIDCFederationTestSuite) TestCreatePolicyRejectsDeprecatedBuilderRole() {
	idp := newFakeIDP(s.T())
	s.deps.Service.cfg.OIDCFederationAllowInsecureIssuers = true

	rr := s.makeRequest(http.MethodPost, "/v1/oidc/trust-policies", CreateOIDCTrustPolicyRequest{
		Name: "builder", IssuerURL: idp.issuer(), Audience: "nuon-test",
		ClaimConditions: map[string]string{"sub": "repo:acme/app:*:*"},
		Role:            string(app.RoleTypeOrgBuilder),
	})
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
}

func (s *OIDCFederationTestSuite) TestCRUDRequiresOrgAdmin() {
	s.demoteTestAccount()

	rr := s.makeRequest(http.MethodGet, "/v1/oidc/trust-policies", nil)
	require.Equal(s.T(), http.StatusForbidden, rr.Code, rr.Body.String())

	rr = s.makeRequest(http.MethodPost, "/v1/oidc/trust-policies", CreateOIDCTrustPolicyRequest{
		Name: "ci", IssuerURL: "https://idp.example.com", Audience: "aud",
		ClaimConditions: map[string]string{"sub": "repo:acme/app:*:*"},
	})
	require.Equal(s.T(), http.StatusForbidden, rr.Code, rr.Body.String())
}

func (s *OIDCFederationTestSuite) TestUpdatePolicyRoleUpdatesServiceAccount() {
	idp := newFakeIDP(s.T())
	policy := s.createPolicy(idp, map[string]string{"sub": "repo:acme/app:*:*"})

	rr := s.makeRequest(http.MethodPatch, "/v1/oidc/trust-policies/"+policy.ID, UpdateOIDCTrustPolicyRequest{
		Role: string(app.RoleTypeOrgAdmin),
	})
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var acct app.Account
	require.NoError(s.T(), s.deps.DB.Preload("Roles").Preload("Roles.Org").Where("id = ?", policy.ServiceAccountID).First(&acct).Error)
	require.Len(s.T(), acct.Roles, 1)
	require.Equal(s.T(), app.RoleTypeOrgAdmin, acct.Roles[0].RoleType)
}

func (s *OIDCFederationTestSuite) TestDeletePolicyDeletesServiceAccount() {
	idp := newFakeIDP(s.T())
	policy := s.createPolicy(idp, map[string]string{"sub": "repo:acme/app:*:*"})

	rr := s.makeRequest(http.MethodDelete, "/v1/oidc/trust-policies/"+policy.ID, nil)
	require.Equal(s.T(), http.StatusNoContent, rr.Code, rr.Body.String())

	var count int64
	s.deps.DB.Model(&app.OIDCTrustPolicy{}).Where("id = ?", policy.ID).Count(&count)
	require.Zero(s.T(), count)

	s.deps.DB.Model(&app.Account{}).Where("id = ?", policy.ServiceAccountID).Count(&count)
	require.Zero(s.T(), count)
}

func (s *OIDCFederationTestSuite) TestExchangeIssuesToken() {
	idp := newFakeIDP(s.T())
	policy := s.createPolicy(idp, map[string]string{"sub": "repo:acme/app:ref:refs/heads/main"})

	token := idp.signToken(s.T(), jwt.MapClaims{"sub": "repo:acme/app:ref:refs/heads/main"})

	rr := s.makeRequest(http.MethodPost, "/v1/oidc/token", ExchangeOIDCTokenRequest{
		OrgID: s.testOrg.ID,
		Token: token,
	})
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var resp ExchangeOIDCTokenResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	require.True(s.T(), resp.Authenticated)
	require.NotEmpty(s.T(), resp.Token)
	require.Equal(s.T(), policy.ID, resp.TrustPolicyID)
	require.Equal(s.T(), string(app.RoleTypeOrgReadOnly), resp.Role)

	var minted app.Token
	require.NoError(s.T(), s.deps.DB.Where(app.Token{Token: resp.Token}).First(&minted).Error)
	require.Equal(s.T(), app.TokenTypeFederated, minted.TokenType)
	require.Equal(s.T(), policy.ServiceAccountID, minted.AccountID)
	require.Equal(s.T(), s.testOrg.ID, minted.OrgID)
	require.WithinDuration(s.T(), minted.IssuedAt.Add(time.Duration(app.OIDCTrustPolicyDefaultTokenDuration)*time.Second), minted.ExpiresAt, 2*time.Second)

	var updated app.OIDCTrustPolicy
	require.NoError(s.T(), s.deps.DB.Where("id = ?", policy.ID).First(&updated).Error)
	require.NotNil(s.T(), updated.LastUsedAt)
}

func (s *OIDCFederationTestSuite) TestExchangeRejections() {
	idp := newFakeIDP(s.T())
	policy := s.createPolicy(idp, map[string]string{"sub": "repo:acme/app:ref:refs/heads/main"})

	matching := jwt.MapClaims{"sub": "repo:acme/app:ref:refs/heads/main"}

	s.Run("unmatched sub", func() {
		token := idp.signToken(s.T(), jwt.MapClaims{"sub": "repo:evil/app:ref:refs/heads/main"})
		rr := s.makeRequest(http.MethodPost, "/v1/oidc/token", ExchangeOIDCTokenRequest{OrgID: s.testOrg.ID, Token: token})
		require.Equal(s.T(), http.StatusUnauthorized, rr.Code, rr.Body.String())
	})

	s.Run("wrong org", func() {
		token := idp.signToken(s.T(), matching)
		rr := s.makeRequest(http.MethodPost, "/v1/oidc/token", ExchangeOIDCTokenRequest{OrgID: "orgdoesnotexist", Token: token})
		require.Equal(s.T(), http.StatusUnauthorized, rr.Code, rr.Body.String())
	})

	s.Run("wrong audience", func() {
		token := idp.signToken(s.T(), jwt.MapClaims{"sub": "repo:acme/app:ref:refs/heads/main", "aud": "someone-else"})
		rr := s.makeRequest(http.MethodPost, "/v1/oidc/token", ExchangeOIDCTokenRequest{OrgID: s.testOrg.ID, Token: token})
		require.Equal(s.T(), http.StatusUnauthorized, rr.Code, rr.Body.String())
	})

	s.Run("missing fields", func() {
		rr := s.makeRequest(http.MethodPost, "/v1/oidc/token", ExchangeOIDCTokenRequest{OrgID: s.testOrg.ID})
		require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	})

	s.Run("disabled policy", func() {
		disabled := false
		rr := s.makeRequest(http.MethodPatch, "/v1/oidc/trust-policies/"+policy.ID, UpdateOIDCTrustPolicyRequest{Enabled: &disabled})
		require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

		token := idp.signToken(s.T(), matching)
		rr = s.makeRequest(http.MethodPost, "/v1/oidc/token", ExchangeOIDCTokenRequest{OrgID: s.testOrg.ID, Token: token})
		require.Equal(s.T(), http.StatusUnauthorized, rr.Code, rr.Body.String())
	})
}
