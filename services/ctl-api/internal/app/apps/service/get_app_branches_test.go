package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// AppBranchesTestSuite is the testify suite for app branches endpoints.
type AppBranchesTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      TestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	testApp      *app.App
	mockEvClient *tests.MockEventLoopClient
}

func TestAppBranchesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AppBranchesTestSuite))
}

func (s *AppBranchesTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	// Create fake event loop client for testing
	s.mockEvClient = tests.NewMockEventLoopClient()

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			Mocks: &tests.TestMocks{MockEv: s.mockEvClient},

			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AppBranchesTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Reset mock before each test
	s.mockEvClient.Reset()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.AppsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AppBranchesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AppBranchesTestSuite) setupTestData() {
	testAcc := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "test@example.com",
		Subject:     "test-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, testAcc)
	testOrg := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "test-org-" + domains.NewOrgID(),
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg

	testApp := &app.App{
		ID:          domains.NewAppID(),
		Name:        "test-app",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}
	err = s.service.DB.Create(testApp).Error
	require.NoError(s.T(), err)
	s.testApp = testApp
}

func (s *AppBranchesTestSuite) makeGetRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AppBranchesTestSuite) makeRequestWithBody(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

// TestGetAppBranches tests the GET /v1/apps/:app_id/branches endpoint.
func (s *AppBranchesTestSuite) TestGetAppBranches() {
	testCases := []struct {
		name          string
		setupFunc     func() []string
		expectedCount int
		expectedCode  int
	}{
		{
			name: "returns empty array when no branches exist",
			setupFunc: func() []string {
				return []string{}
			},
			expectedCount: 0,
			expectedCode:  http.StatusOK,
		},
		{
			name: "returns branches after creating some",
			setupFunc: func() []string {
				branchIDs := []string{}

				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				ctx = cctx.SetOrgContext(ctx, s.testOrg)

				vcsConn := &app.VCSConnection{
					OrgID:             s.testOrg.ID,
					GithubInstallID:   "test-install-" + domains.NewVCSConnectionID(),
					GithubAccountID:   "test-account",
					GithubAccountName: "test-account-name",
				}
				err := s.service.DB.WithContext(ctx).Create(vcsConn).Error
				require.NoError(s.T(), err)

				vcsConfig := &app.ConnectedGithubVCSConfig{
					OrgID:           s.testOrg.ID,
					VCSConnectionID: vcsConn.ID,
					Repo:            "test-repo",
				}
				err = s.service.DB.WithContext(ctx).Create(vcsConfig).Error
				require.NoError(s.T(), err)

				for i := 0; i < 3; i++ {
					branch := &app.AppBranch{
						OrgID:                      s.testOrg.ID,
						AppID:                      s.testApp.ID,
						Name:                       domains.NewAppBranchID(),
						ConnectedGithubVCSConfigID: vcsConfig.ID,
					}
					err := s.service.DB.WithContext(ctx).Create(branch).Error
					require.NoError(s.T(), err)
					branchIDs = append(branchIDs, branch.ID)
				}

				s.T().Cleanup(func() {
					for _, branchID := range branchIDs {
						s.service.DB.Unscoped().Delete(&app.AppBranch{}, "id = ?", branchID)
					}
					s.service.DB.Unscoped().Delete(&app.ConnectedGithubVCSConfig{}, "id = ?", vcsConfig.ID)
					s.service.DB.Unscoped().Delete(&app.VCSConnection{}, "id = ?", vcsConn.ID)
				})

				return branchIDs
			},
			expectedCount: 3,
			expectedCode:  http.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			branchIDs := tc.setupFunc()

			rr := s.makeGetRequest(http.MethodGet, "/v1/apps/"+s.testApp.ID+"/branches")

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			var response []app.AppBranch
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(s.T(), err)
			require.NotNil(s.T(), response)
			require.Len(s.T(), response, tc.expectedCount)

			if len(branchIDs) > 0 {
				receivedIDs := make(map[string]bool)
				for _, branch := range response {
					receivedIDs[branch.ID] = true
				}
				for _, expectedID := range branchIDs {
					assert.True(s.T(), receivedIDs[expectedID], "Expected branch ID %s not found in response", expectedID)
				}
			}
		})
	}
}
