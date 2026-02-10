package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// GetAllOrgsTestService holds all fx-injected dependencies for GetAllOrgs endpoint tests.
type GetAllOrgsTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	OrgsService     *service
}

// GetAllOrgsTestSuite is the testify suite for GetAllOrgs endpoint.
type GetAllOrgsTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service GetAllOrgsTestService
	router  *gin.Engine
	testAcc *app.Account
}

func TestGetAllOrgsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetAllOrgsTestSuite))
}

func (s *GetAllOrgsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)

	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *GetAllOrgsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares
	// Note: GetAllOrgs is an admin endpoint, no org context needed
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetAllOrgsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetAllOrgsTestSuite) setupTestData() {
	// Create test account
	testAcc := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "admin@example.com",
		Subject:     "admin-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc
}

func (s *GetAllOrgsTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetAllOrgsTestSuite) TestGetAllOrgs() {
	testCases := []struct {
		name          string
		setupFunc     func() []string // Returns org IDs that should be returned
		queryParams   string
		expectedCount int
		expectedCode  int
		validateFunc  func([]app.Org) // Additional validations
	}{
		{
			name: "returns empty array when no orgs",
			setupFunc: func() []string {
				return []string{}
			},
			queryParams:   "",
			expectedCount: 0,
			expectedCode:  http.StatusOK,
		},
		{
			name: "returns all types when type is empty string",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org1 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "sandbox-org",
					SandboxMode: true,
					OrgType:     app.OrgTypeSandbox,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				org2 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "default-org",
					SandboxMode: true,
					OrgType:     app.OrgTypeDefault,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}

				err := s.service.DB.WithContext(ctx).Create(org1).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org1.ID)
				})

				err = s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org2.ID)
				})

				return []string{org1.ID, org2.ID}
			},
			queryParams:   "?type=",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
		},
		{
			name: "respects pagination limit",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				orgIDs := make([]string, 0, 15)
				for i := 0; i < 15; i++ {
					testOrg := &app.Org{
						ID:          domains.NewOrgID(),
						Name:        fmt.Sprintf("test-org-%02d", i),
						SandboxMode: true,
						OrgType:     app.OrgTypeSandbox, // Use sandbox so type=sandbox filter works
						NotificationsConfig: app.NotificationsConfig{
							InternalSlackWebhookURL: "https://hooks.slack.com/foo",
						},
					}
					err := s.service.DB.WithContext(ctx).Create(testOrg).Error
					require.NoError(s.T(), err)
					orgID := testOrg.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", orgID)
					})
					orgIDs = append(orgIDs, testOrg.ID)
				}
				return orgIDs
			},
			queryParams:   "?type=sandbox&limit=5",
			expectedCount: 5,
			expectedCode:  http.StatusOK,
		},
		{
			name: "respects pagination offset",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				orgIDs := make([]string, 0, 10)
				for i := 0; i < 10; i++ {
					testOrg := &app.Org{
						ID:          domains.NewOrgID(),
						Name:        fmt.Sprintf("test-org-%02d", i),
						SandboxMode: true,
						OrgType:     app.OrgTypeDefault,
						NotificationsConfig: app.NotificationsConfig{
							InternalSlackWebhookURL: "https://hooks.slack.com/foo",
						},
					}
					err := s.service.DB.WithContext(ctx).Create(testOrg).Error
					require.NoError(s.T(), err)
					orgID := testOrg.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", orgID)
					})
					orgIDs = append(orgIDs, testOrg.ID)
				}
				return orgIDs
			},
			queryParams:   "?type=default&offset=5&limit=3",
			expectedCount: 3,
			expectedCode:  http.StatusOK,
		},
		{
			name: "orders by created_at DESC",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create orgs sequentially
				org1 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "oldest-org",
					SandboxMode: true,
					OrgType:     app.OrgTypeIntegration,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org1).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org1.ID)
				})

				org2 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "middle-org",
					SandboxMode: true,
					OrgType:     app.OrgTypeIntegration,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err = s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org2.ID)
				})

				org3 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "newest-org",
					SandboxMode: true,
					OrgType:     app.OrgTypeIntegration,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err = s.service.DB.WithContext(ctx).Create(org3).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org3.ID)
				})

				return []string{org3.ID, org2.ID, org1.ID} // Expect DESC order
			},
			queryParams:   "?type=integration",
			expectedCount: 3,
			expectedCode:  http.StatusOK,
			validateFunc: func(orgs []app.Org) {
				require.Equal(s.T(), "newest-org", orgs[0].Name)
				require.Equal(s.T(), "middle-org", orgs[1].Name)
				require.Equal(s.T(), "oldest-org", orgs[2].Name)
			},
		},
		{
			name: "pagination with page parameter",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				orgIDs := make([]string, 0, 25)
				for i := 0; i < 25; i++ {
					testOrg := &app.Org{
						ID:          domains.NewOrgID(),
						Name:        fmt.Sprintf("page-test-org-%02d", i),
						SandboxMode: true,
						OrgType:     app.OrgTypeSandbox,
						NotificationsConfig: app.NotificationsConfig{
							InternalSlackWebhookURL: "https://hooks.slack.com/foo",
						},
					}
					err := s.service.DB.WithContext(ctx).Create(testOrg).Error
					require.NoError(s.T(), err)
					orgID := testOrg.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", orgID)
					})
					orgIDs = append(orgIDs, testOrg.ID)
				}
				return orgIDs
			},
			queryParams:   "?type=sandbox&limit=10&page=2",
			expectedCount: 10,
			expectedCode:  http.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			_ = tc.setupFunc()

			// Make request
			rr := s.makeRequest(http.MethodGet, "/v1/orgs"+tc.queryParams)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			// Parse response
			var response []app.Org
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)
			require.NotNil(s.T(), response)

			// Validate expected count
			if tc.expectedCount >= 0 {
				require.Len(s.T(), response, tc.expectedCount)
			}

			// Run additional validations if provided
			if tc.validateFunc != nil && len(response) > 0 {
				tc.validateFunc(response)
			}
		})
	}
}
