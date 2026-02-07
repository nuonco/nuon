package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// GetOrgTestSuite is the test suite for get org endpointi.
type GetOrgTestSuite struct {
	tests.BaseDBTestSuite

	app         *fxtest.App
	service     TestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	orgsService *service
}

func TestGetOrgSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetOrgTestSuite))
}

func (s *GetOrgTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service, &s.orgsService),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *GetOrgTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.orgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetOrgTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetOrgTestSuite) setupTestData() {
	// Create test account
	testAcc := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "test@example.com",
		Subject:     "test-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	// Create test org with account context (required by BeforeCreate hook)
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, testAcc)
	testOrg := &app.Org{
		ID:   domains.NewOrgID(),
		Name: "test-org",
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
}

func (s *GetOrgTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetOrgTestSuite) TestGetOrg() {
	testCases := []struct {
		name           string
		setupFunc      func() *app.Org
		expectedStatus int
		validateFunc   func(*app.Org)
	}{
		{
			name: "successfully returns org with all associations",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:     domains.NewOrgID(),
					Name:   "test-org-full",
					Status: "active",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				// Create a runner group for the org
				runnerGroup := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OwnerID:   org.ID,
					OwnerType: "orgs",
				}
				err = s.service.DB.WithContext(ctx).Create(runnerGroup).Error
				require.NoError(s.T(), err)

				return org
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.Equal(s.T(), "test-org-full", org.Name)
				require.Equal(s.T(), app.OrgStatus("active"), org.Status)
				require.NotNil(s.T(), org.RunnerGroup)
			},
		},
		{
			name: "returns org with vcs connections",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "test-org-vcs",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				// Create a VCS connection
				vcsConn := &app.VCSConnection{
					OrgID:             org.ID,
					GithubInstallID:   "12345",
					GithubAccountID:   "67890",
					GithubAccountName: "test-account",
				}
				err = s.service.DB.WithContext(ctx).Create(vcsConn).Error
				require.NoError(s.T(), err)

				return org
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.Equal(s.T(), "test-org-vcs", org.Name)
				require.NotEmpty(s.T(), org.VCSConnections)
				require.Equal(s.T(), "test-account", org.VCSConnections[0].GithubAccountName)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			org := tc.setupFunc()

			// Update router context to use the test org
			s.router = tests.NewTestRouter(tests.RouterOptions{
				L:       s.service.L,
				DB:      s.service.DB,
				TestOrg: org,
				TestAcc: s.testAcc,
			})
			err := s.orgsService.RegisterPublicRoutes(s.router)
			require.NoError(s.T(), err)

			// Make request
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/current")

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Parse response
			var response app.Org
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)

			// Run validation
			if tc.validateFunc != nil {
				tc.validateFunc(&response)
			}
		})
	}
}

func (s *GetOrgTestSuite) TestGetOrgWithoutOrgContext() {
	// Create router without org context
	router := tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
		// TestOrg intentionally omitted
	})

	err := s.orgsService.RegisterPublicRoutes(router)
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, "/v1/orgs/current", nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should fail without org context
	require.NotEqual(s.T(), http.StatusOK, rr.Code)
}
