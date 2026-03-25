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
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// TestService holds all fx-injected dependencies for orgs endpoint tests.
type TestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	OrgsService     *service
	Seeder          *testseed.Seeder
}

// OrgsTestSuite is the testify suite for orgs endpoints.
type OrgsTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service TestService
	router  *gin.Engine
	testAcc *app.Account
}

func TestOrgsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(OrgsTestSuite))
}

func (s *OrgsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)

	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *OrgsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *OrgsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *OrgsTestSuite) setupTestData() {
	ctx := context.Background()
	_, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
}

func (s *OrgsTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *OrgsTestSuite) TestGetOrgs() {
	// Generate unique names for each test case to avoid cross-run collisions
	org1Name := fmt.Sprintf("test-org-1-%s", domains.NewOrgID()[:8])
	org2Name := fmt.Sprintf("test-org-2-%s", domains.NewOrgID()[:8])
	searchSuffix := domains.NewOrgID()[:8]
	frontendName := fmt.Sprintf("frontend-team-%s", searchSuffix)
	backendName := fmt.Sprintf("backend-team-%s", searchSuffix)

	testCases := []struct {
		name          string
		setupFunc     func() []string // Returns org IDs that should be accessible
		queryParams   string
		expectedCount int
		validateFunc  func([]app.Org) // Additional validations
	}{
		{
			name: "returns empty array when no orgs",
			setupFunc: func() []string {
				return []string{}
			},
			queryParams:   "",
			expectedCount: 0,
		},
		{
			name: "returns created orgs",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org1 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        org1Name,
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				org2 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        org2Name,
					SandboxMode: true,
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
			queryParams:   "",
			expectedCount: 2,
			validateFunc: func(orgs []app.Org) {
				orgNames := []string{orgs[0].Name, orgs[1].Name}
				require.Contains(s.T(), orgNames, org1Name)
				require.Contains(s.T(), orgNames, org2Name)
			},
		},
		{
			name: "filters with search query",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org1 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        frontendName,
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				org2 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        backendName,
					SandboxMode: true,
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
			queryParams:   fmt.Sprintf("?q=frontend-team-%s", searchSuffix),
			expectedCount: 1,
			validateFunc: func(orgs []app.Org) {
				require.Equal(s.T(), frontendName, orgs[0].Name)
			},
		},
		{
			name: "respects pagination",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				orgIDs := make([]string, 0, 15)
				for i := 0; i < 15; i++ {
					testOrg := &app.Org{
						ID:          domains.NewOrgID(),
						Name:        fmt.Sprintf("test-org-%02d", i),
						SandboxMode: true,
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
			queryParams:   "?limit=5",
			expectedCount: 5,
			validateFunc: func(orgs []app.Org) {
				require.LessOrEqual(s.T(), len(orgs), 5)
			},
		},
		{
			name: "only returns user accessible orgs",
			setupFunc: func() []string {
				// Create second account
				acc2ID := domains.NewAccountID()
				acc2 := &app.Account{
					ID:          acc2ID,
					Email:       fmt.Sprintf("%s@test.nuon.co", acc2ID),
					Subject:     acc2ID,
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(acc2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc2.ID)
				})

				// Create two orgs with different account contexts
				ctx1 := context.Background()
				ctx1 = cctx.SetAccountContext(ctx1, s.testAcc)

				ctx2 := context.Background()
				ctx2 = cctx.SetAccountContext(ctx2, acc2)

				myOrgID := domains.NewOrgID()
				myOrg := &app.Org{
					ID:          myOrgID,
					Name:        fmt.Sprintf("my-org-%s", myOrgID),
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err = s.service.DB.WithContext(ctx1).Create(myOrg).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", myOrg.ID)
				})

				otherOrgID := domains.NewOrgID()
				otherOrg := &app.Org{
					ID:          otherOrgID,
					Name:        fmt.Sprintf("other-org-%s", otherOrgID),
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err = s.service.DB.WithContext(ctx2).Create(otherOrg).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", otherOrg.ID)
				})

				// Return only myOrg ID (simulates RBAC permission resolution)
				return []string{myOrg.ID}
			},
			queryParams:   "",
			expectedCount: 1,
			validateFunc: func(orgs []app.Org) {
				require.Contains(s.T(), orgs[0].Name, "my-org-")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			orgIDs := tc.setupFunc()

			// Update account's OrgIDs (simulates permission resolution)
			s.testAcc.OrgIDs = orgIDs
			err := s.service.DB.Save(s.testAcc).Error
			require.NoError(s.T(), err)

			// Make request
			rr := s.makeRequest(http.MethodGet, "/v1/orgs"+tc.queryParams)

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			// Parse response
			var response []app.Org
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)
			require.NotNil(s.T(), response)

			// Validate expected count
			if tc.expectedCount > 0 {
				require.Len(s.T(), response, tc.expectedCount)
			}

			// Run additional validations if provided
			if tc.validateFunc != nil && len(response) > 0 {
				tc.validateFunc(response)
			}
		})
	}
}
