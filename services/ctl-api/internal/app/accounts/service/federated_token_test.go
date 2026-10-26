package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *AccountsServiceTestSuite) makeOrgAdmin() *app.Account {
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	require.NoError(s.T(), s.service.AccountsService.authzClient.CreateOrgRoles(ctx, s.testOrg.ID))
	require.NoError(s.T(), s.service.AccountsService.authzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, s.testOrg.ID, s.testAcc.ID))

	admin, err := s.service.AccountsService.acctClient.FetchAccount(ctx, s.testAcc.ID)
	require.NoError(s.T(), err)
	return admin
}

func (s *AccountsServiceTestSuite) routerWithTokenType(admin *app.Account, tokenType app.TokenType) *gin.Engine {
	router := tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: admin,
		AdditionalMiddlewares: []gin.HandlerFunc{
			func(c *gin.Context) {
				cctx.SetTokenTypeGinContext(c, tokenType)
				c.Next()
			},
		},
	})
	require.NoError(s.T(), s.service.AccountsService.RegisterPublicRoutes(router))
	return router
}

func (s *AccountsServiceTestSuite) countStaticTokens() int64 {
	var count int64
	require.NoError(s.T(), s.service.DB.Model(&app.Token{}).
		Where(app.Token{OrgID: s.testOrg.ID, TokenType: app.TokenTypeStatic}).
		Count(&count).Error)
	return count
}

func (s *AccountsServiceTestSuite) countServiceAccounts() int64 {
	var count int64
	require.NoError(s.T(), s.service.DB.Model(&app.Account{}).
		Where(app.Account{AccountType: app.AccountTypeService}).
		Count(&count).Error)
	return count
}

func serveJSON(router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	jsonBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func (s *AccountsServiceTestSuite) TestCreateStaticTokenRejectsFederatedToken() {
	admin := s.makeOrgAdmin()

	testCases := map[string]struct {
		tokenType  app.TokenType
		identity   string
		wantStatus int
		wantMinted bool
		wantNewSA  bool
	}{
		"federated service account token": {
			tokenType:  app.TokenTypeFederated,
			identity:   TokenIdentityServiceAccount,
			wantStatus: http.StatusForbidden,
		},
		"federated personal token": {
			tokenType:  app.TokenTypeFederated,
			identity:   TokenIdentityPersonal,
			wantStatus: http.StatusForbidden,
		},
		"user service account token": {
			tokenType:  app.TokenTypeNuon,
			identity:   TokenIdentityServiceAccount,
			wantStatus: http.StatusCreated,
			wantMinted: true,
			wantNewSA:  true,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			router := s.routerWithTokenType(admin, tc.tokenType)
			tokensBefore := s.countStaticTokens()
			accountsBefore := s.countServiceAccounts()

			rr := serveJSON(router, http.MethodPost, "/v1/account/static-token", CreateStaticTokenRequest{
				Name:          "ci",
				Duration:      "87600h",
				TokenIdentity: tc.identity,
			})
			require.Equal(s.T(), tc.wantStatus, rr.Code, rr.Body.String())

			if tc.wantStatus == http.StatusForbidden {
				require.Contains(s.T(), rr.Body.String(), "federated tokens cannot create API tokens")
			}
			require.Equal(s.T(), tc.wantMinted, s.countStaticTokens() > tokensBefore)
			require.Equal(s.T(), tc.wantNewSA, s.countServiceAccounts() > accountsBefore)
		})
	}
}

func (s *AccountsServiceTestSuite) TestCreateServiceAccountTokenRejectsFederatedToken() {
	admin := s.makeOrgAdmin()

	testCases := map[string]struct {
		tokenType  app.TokenType
		wantStatus int
	}{
		"federated token": {tokenType: app.TokenTypeFederated, wantStatus: http.StatusForbidden},
		"user token":      {tokenType: app.TokenTypeNuon, wantStatus: http.StatusCreated},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			router := s.routerWithTokenType(admin, tc.tokenType)
			target, err := s.service.AccountsService.createTokenServiceAccount(cctx.SetAccountContext(s.ctx, s.testAcc), s.testOrg.ID, app.RoleTypeOrgAdmin)
			require.NoError(s.T(), err)
			tokensBefore := s.countStaticTokens()

			rr := serveJSON(router, http.MethodPost, fmt.Sprintf("/v1/service-accounts/%s/tokens", target.ID), CreateServiceAccountTokenRequest{
				Duration: "87600h",
			})
			require.Equal(s.T(), tc.wantStatus, rr.Code, rr.Body.String())
			require.Equal(s.T(), tc.wantStatus == http.StatusCreated, s.countStaticTokens() > tokensBefore)
		})
	}
}
