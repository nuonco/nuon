package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *TokensServiceTestSuite) makeOrgAdmin() *app.Account {
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	require.NoError(s.T(), s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID))
	require.NoError(s.T(), s.service.AuthzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, s.testOrg.ID, s.testAcc.ID))

	admin, err := s.service.AcctClient.FindAccount(ctx, s.testAcc.ID)
	require.NoError(s.T(), err)
	return admin
}

func (s *TokensServiceTestSuite) useAdminRouter() {
	admin := s.makeOrgAdmin()
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: admin,
	})
	require.NoError(s.T(), s.service.TokensService.RegisterPublicRoutes(s.router))
}

func (s *TokensServiceTestSuite) createOrgServiceAccount(orgID string, roleType app.RoleType) *app.Account {
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)

	acct := s.service.Seeder.CreateServiceAccount(ctx, s.T(), domains.NewAccountID())
	require.NoError(s.T(), s.service.AuthzClient.AddAccountOrgRole(ctx, roleType, orgID, acct.ID))

	return acct
}

func (s *TokensServiceTestSuite) createOtherOrg() *app.Org {
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)

	org := s.service.Seeder.CreateOrg(ctx, s.T())
	require.NoError(s.T(), s.service.AuthzClient.CreateOrgRoles(ctx, org.ID))

	return org
}

func (s *TokensServiceTestSuite) seedToken(orgID, accountID, name string) *app.Token {
	token := &app.Token{
		CreatedByID: s.testAcc.ID,
		Name:        name,
		Role:        string(app.RoleTypeOrgReadOnly),
		OrgID:       orgID,
		Token:       domains.NewUserTokenID(),
		TokenType:   app.TokenTypeStatic,
		ExpiresAt:   time.Now().Add(time.Hour),
		IssuedAt:    time.Now(),
		Issuer:      accountID,
		AccountID:   accountID,
	}
	require.NoError(s.T(), s.service.DB.Create(token).Error)
	return token
}

func (s *TokensServiceTestSuite) listTokens(path string) []app.Token {
	return tokensFromResponses(s.listTokenResponses(path))
}

func (s *TokensServiceTestSuite) listTokenResponses(path string) []TokenResponse {
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var tokens []TokenResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &tokens))
	return tokens
}

func tokenIDs(tokens []app.Token) []string {
	ids := make([]string, 0, len(tokens))
	for _, token := range tokens {
		ids = append(ids, token.ID)
	}
	return ids
}

func tokensFromResponses(tokens []TokenResponse) []app.Token {
	out := make([]app.Token, 0, len(tokens))
	for _, t := range tokens {
		out = append(out, t.Token)
	}
	return out
}

func (s *TokensServiceTestSuite) ensureOrgRoles() {
	require.NoError(s.T(), s.service.AuthzClient.CreateOrgRoles(cctx.SetAccountContext(s.ctx, s.testAcc), s.testOrg.ID))
}

func (s *TokensServiceTestSuite) requireNonAdminForbidden(method, path string, body any) {
	rr := s.makeRequest(method, path, body)
	if rr.Code != http.StatusForbidden {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusForbidden, rr.Code)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func (s *TokensServiceTestSuite) TestListTokens() {
	s.useAdminRouter()

	acct := s.createOrgServiceAccount(s.testOrg.ID, app.RoleTypeOrgReadOnly)
	otherAcct := s.createOrgServiceAccount(s.testOrg.ID, app.RoleTypeOrgReadOnly)

	deployToken := s.seedToken(s.testOrg.ID, acct.ID, "Acme-Deploy")
	ciToken := s.seedToken(s.testOrg.ID, acct.ID, "ci")
	otherToken := s.seedToken(s.testOrg.ID, otherAcct.ID, "other-deploy")
	personalToken := s.seedToken(s.testOrg.ID, s.testAcc.ID, "personal-deploy")

	otherOrg := s.createOtherOrg()
	otherOrgAcct := s.createOrgServiceAccount(otherOrg.ID, app.RoleTypeOrgReadOnly)
	otherOrgToken := s.seedToken(otherOrg.ID, otherOrgAcct.ID, "other-org-deploy")

	s.Run("returns tokens across service accounts, excluding personal and other-org tokens", func() {
		ids := tokenIDs(s.listTokens("/v1/tokens"))

		assert.ElementsMatch(s.T(), []string{deployToken.ID, ciToken.ID, otherToken.ID}, ids)
		assert.NotContains(s.T(), ids, personalToken.ID)
		assert.NotContains(s.T(), ids, otherOrgToken.ID)
	})

	s.Run("filters by account_id", func() {
		ids := tokenIDs(s.listTokens(fmt.Sprintf("/v1/tokens?account_id=%s", acct.ID)))
		assert.ElementsMatch(s.T(), []string{deployToken.ID, ciToken.ID}, ids)
	})

	s.Run("q matches name case-insensitively", func() {
		ids := tokenIDs(s.listTokens("/v1/tokens?q=DEPLOY"))
		assert.ElementsMatch(s.T(), []string{deployToken.ID, otherToken.ID}, ids)
	})

	s.Run("q matches exact ID", func() {
		ids := tokenIDs(s.listTokens(fmt.Sprintf("/v1/tokens?q=%s", ciToken.ID)))
		assert.Equal(s.T(), []string{ciToken.ID}, ids)
	})

	s.Run("q combined with account_id", func() {
		ids := tokenIDs(s.listTokens(fmt.Sprintf("/v1/tokens?q=deploy&account_id=%s", otherAcct.ID)))
		assert.Equal(s.T(), []string{otherToken.ID}, ids)
	})

	s.Run("limit and offset page through results", func() {
		rr := s.makeRequest(http.MethodGet, "/v1/tokens?limit=2&offset=0", nil)
		require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())
		assert.Equal(s.T(), "true", rr.Header().Get("X-Nuon-Page-Next"))

		var first []app.Token
		require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &first))
		require.Len(s.T(), first, 2)

		rr = s.makeRequest(http.MethodGet, "/v1/tokens?limit=2&offset=2", nil)
		require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())
		assert.Equal(s.T(), "false", rr.Header().Get("X-Nuon-Page-Next"))

		var second []app.Token
		require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &second))
		require.Len(s.T(), second, 1)

		assert.ElementsMatch(s.T(), []string{deployToken.ID, ciToken.ID, otherToken.ID}, append(tokenIDs(first), tokenIDs(second)...))
	})

	s.Run("secrets are not returned", func() {
		rr := s.makeRequest(http.MethodGet, "/v1/tokens", nil)
		require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

		body := rr.Body.String()
		assert.NotContains(s.T(), body, deployToken.Token)
		assert.NotContains(s.T(), body, ciToken.Token)
		assert.NotContains(s.T(), body, otherToken.Token)
	})

	s.Run("account_id that is not a service account returns 404", func() {
		rr := s.makeRequest(http.MethodGet, fmt.Sprintf("/v1/tokens?account_id=%s", s.testAcc.ID), nil)
		require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())
	})

	s.Run("service account in another org returns 403", func() {
		rr := s.makeRequest(http.MethodGet, fmt.Sprintf("/v1/tokens?account_id=%s", otherOrgAcct.ID), nil)
		require.Equal(s.T(), http.StatusForbidden, rr.Code, rr.Body.String())
	})

	s.Run("includes the owning account's name, email, and current org role", func() {
		resp := s.listTokenResponses(fmt.Sprintf("/v1/tokens?account_id=%s", acct.ID))
		require.NotEmpty(s.T(), resp)

		for _, t := range resp {
			assert.Equal(s.T(), acct.Name, t.AccountName)
			assert.Equal(s.T(), acct.Email, t.AccountEmail)
			assert.Equal(s.T(), string(app.RoleTypeOrgReadOnly), t.AccountRole)
		}
	})

	s.Run("account_role reflects a role change after token creation, not the stale token role", func() {
		roleAcct := s.createOrgServiceAccount(s.testOrg.ID, app.RoleTypeOrgReadOnly)
		roleToken := s.seedToken(s.testOrg.ID, roleAcct.ID, "role-change")
		require.Equal(s.T(), string(app.RoleTypeOrgReadOnly), roleToken.Role)

		ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
		require.NoError(s.T(), authz.New(authz.Params{DB: s.service.DB}).SetAccountOrgRole(ctx, s.testOrg.ID, roleAcct.ID, app.RoleTypeOrgAdmin))

		resp := s.listTokenResponses(fmt.Sprintf("/v1/tokens?account_id=%s", roleAcct.ID))
		require.Len(s.T(), resp, 1)
		assert.Equal(s.T(), string(app.RoleTypeOrgAdmin), resp[0].AccountRole)
	})
}

func (s *TokensServiceTestSuite) TestListTokensRequiresOrgAdmin() {
	s.requireNonAdminForbidden(http.MethodGet, "/v1/tokens", nil)
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func (s *TokensServiceTestSuite) TestCreateToken() {
	s.useAdminRouter()

	acct := s.createOrgServiceAccount(s.testOrg.ID, app.RoleTypeOrgReadOnly)

	s.Run("creates a static token and returns the secret", func() {
		rr := s.makeRequest(http.MethodPost, "/v1/tokens", CreateTokenRequest{
			AccountID: acct.ID,
			Name:      "acme-ci",
			Duration:  "24h",
		})
		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var resp CreateTokenResponse
		require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.NotEmpty(s.T(), resp.ID)
		assert.NotEmpty(s.T(), resp.Token)
		assert.Equal(s.T(), acct.ID, resp.AccountID)
		assert.Equal(s.T(), "acme-ci", resp.Name)
		assert.WithinDuration(s.T(), time.Now().Add(24*time.Hour), resp.ExpiresAt, time.Minute)
		assert.NotZero(s.T(), resp.CreatedAt)

		var token app.Token
		require.NoError(s.T(), s.service.DB.Where(app.Token{ID: resp.ID}).First(&token).Error)
		assert.Equal(s.T(), app.TokenTypeStatic, token.TokenType)
		assert.Equal(s.T(), s.testOrg.ID, token.OrgID)
		assert.Equal(s.T(), "acme-ci", token.Name)
		assert.Equal(s.T(), acct.ID, token.AccountID)
		assert.Equal(s.T(), resp.Token, token.Token)
	})

	s.Run("missing account_id returns 400", func() {
		rr := s.makeRequest(http.MethodPost, "/v1/tokens", CreateTokenRequest{Name: "acme-ci"})
		require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	})

	s.Run("account that is not a service account returns 404", func() {
		rr := s.makeRequest(http.MethodPost, "/v1/tokens", CreateTokenRequest{AccountID: s.testAcc.ID})
		require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())
	})

	s.Run("service account in another org returns 403", func() {
		otherOrgAcct := s.createOrgServiceAccount(s.createOtherOrg().ID, app.RoleTypeOrgReadOnly)

		rr := s.makeRequest(http.MethodPost, "/v1/tokens", CreateTokenRequest{AccountID: otherOrgAcct.ID})
		require.Equal(s.T(), http.StatusForbidden, rr.Code, rr.Body.String())
	})

	s.Run("invalidate revokes existing tokens", func() {
		target := s.createOrgServiceAccount(s.testOrg.ID, app.RoleTypeOrgReadOnly)
		existing := s.seedToken(s.testOrg.ID, target.ID, "old")

		rr := s.makeRequest(http.MethodPost, "/v1/tokens", CreateTokenRequest{
			AccountID:  target.ID,
			Invalidate: true,
		})
		require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

		var resp CreateTokenResponse
		require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))

		err := s.service.DB.Where(app.Token{ID: existing.ID}).First(&app.Token{}).Error
		assert.ErrorIs(s.T(), err, gorm.ErrRecordNotFound)
		require.NoError(s.T(), s.service.DB.Where(app.Token{ID: resp.ID}).First(&app.Token{}).Error)
	})
}

func (s *TokensServiceTestSuite) TestCreateTokenRequiresOrgAdmin() {
	s.ensureOrgRoles()
	acct := s.createOrgServiceAccount(s.testOrg.ID, app.RoleTypeOrgReadOnly)
	s.requireNonAdminForbidden(http.MethodPost, "/v1/tokens", CreateTokenRequest{AccountID: acct.ID})
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func (s *TokensServiceTestSuite) TestDeleteToken() {
	s.useAdminRouter()

	acct := s.createOrgServiceAccount(s.testOrg.ID, app.RoleTypeOrgReadOnly)
	token := s.seedToken(s.testOrg.ID, acct.ID, "acme-ci")

	s.Run("revokes the token and keeps the service account", func() {
		rr := s.makeRequest(http.MethodDelete, fmt.Sprintf("/v1/tokens/%s", token.ID), nil)
		require.Equal(s.T(), http.StatusNoContent, rr.Code, rr.Body.String())

		err := s.service.DB.Where(app.Token{ID: token.ID}).First(&app.Token{}).Error
		assert.ErrorIs(s.T(), err, gorm.ErrRecordNotFound)

		require.NoError(s.T(), s.service.DB.Where(app.Account{ID: acct.ID}).First(&app.Account{}).Error)
	})

	s.Run("personal token returns 404 and is not deleted", func() {
		personal := s.seedToken(s.testOrg.ID, s.testAcc.ID, "personal")

		rr := s.makeRequest(http.MethodDelete, fmt.Sprintf("/v1/tokens/%s", personal.ID), nil)
		require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())

		require.NoError(s.T(), s.service.DB.Where(app.Token{ID: personal.ID}).First(&app.Token{}).Error)
	})

	s.Run("other org token returns 404 and is not deleted", func() {
		otherOrg := s.createOtherOrg()
		otherOrgAcct := s.createOrgServiceAccount(otherOrg.ID, app.RoleTypeOrgReadOnly)
		otherOrgToken := s.seedToken(otherOrg.ID, otherOrgAcct.ID, "other-org")

		rr := s.makeRequest(http.MethodDelete, fmt.Sprintf("/v1/tokens/%s", otherOrgToken.ID), nil)
		require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())

		require.NoError(s.T(), s.service.DB.Where(app.Token{ID: otherOrgToken.ID}).First(&app.Token{}).Error)
	})
}

func (s *TokensServiceTestSuite) TestDeleteTokenRequiresOrgAdmin() {
	s.ensureOrgRoles()
	acct := s.createOrgServiceAccount(s.testOrg.ID, app.RoleTypeOrgReadOnly)
	token := s.seedToken(s.testOrg.ID, acct.ID, "acme-ci")

	s.requireNonAdminForbidden(http.MethodDelete, fmt.Sprintf("/v1/tokens/%s", token.ID), nil)

	require.NoError(s.T(), s.service.DB.Where(app.Token{ID: token.ID}).First(&app.Token{}).Error)
}
