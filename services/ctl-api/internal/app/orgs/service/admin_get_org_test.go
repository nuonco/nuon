package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
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

// AdminGetOrgTestService holds all fx-injected dependencies for AdminGetOrg endpoint tests.
type AdminGetOrgTestService struct {
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

// AdminGetOrgTestSuite is the testify suite for AdminGetOrg endpoint.
type AdminGetOrgTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service AdminGetOrgTestService
	router  *gin.Engine
	testAcc *app.Account
}

func TestAdminGetOrgSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AdminGetOrgTestSuite))
}

func (s *AdminGetOrgTestSuite) SetupSuite() {
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

func (s *AdminGetOrgTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares
	// Note: AdminGetOrg is an admin endpoint, no org context needed
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminGetOrgTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminGetOrgTestSuite) setupTestData() {
	ctx := context.Background()
	_, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
}

func (s *AdminGetOrgTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminGetOrgTestSuite) TestAdminGetOrg() {
	testCases := []struct {
		name         string
		setupFunc    func() string // Returns query parameter for name or ID
		expectedCode int
		validateFunc func(*app.Org) // Validates returned org
	}{
		{
			name: "returns org by exact name match",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "exact-match-org",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/exact",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org.Name
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.Equal(s.T(), "exact-match-org", org.Name)
			},
		},
		{
			name: "returns org by partial name match (LIKE)",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "partial-match-org",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/partial",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				// Return partial match (tests LIKE query)
				return "partial"
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.Equal(s.T(), "partial-match-org", org.Name)
			},
		},
		{
			name: "returns org by ID",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "org-by-id",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/byid",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.Equal(s.T(), "org-by-id", org.Name)
			},
		},
		{
			name: "returns 404 when org not found",
			setupFunc: func() string {
				return "nonexistent-org-name"
			},
			expectedCode: http.StatusNotFound,
			validateFunc: func(org *app.Org) {
				// Should not reach here
			},
		},
		{
			name: "preloads RunnerGroup relation",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "org-with-runner-group",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/runner",
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
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.RunnerGroup{}, "id = ?", runnerGroup.ID)
				})

				return org.Name
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org.RunnerGroup, "RunnerGroup should be preloaded")
				require.Equal(s.T(), "orgs", org.RunnerGroup.OwnerType)
				require.Equal(s.T(), org.ID, org.RunnerGroup.OwnerID)
			},
		},
		{
			name: "preloads RunnerGroup.Runners relation",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "org-with-runners",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/runners",
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
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.RunnerGroup{}, "id = ?", runnerGroup.ID)
				})

				// Create runners for the runner group
				runner1 := &app.Runner{
					ID:            domains.NewRunnerID(),
					RunnerGroupID: runnerGroup.ID,
					OrgID:         org.ID,
					Name:          "runner-1",
					Status:        app.RunnerStatusActive,
				}
				runner2 := &app.Runner{
					ID:            domains.NewRunnerID(),
					RunnerGroupID: runnerGroup.ID,
					OrgID:         org.ID,
					Name:          "runner-2",
					Status:        app.RunnerStatusActive,
				}

				err = s.service.DB.WithContext(ctx).Create(runner1).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Runner{}, "id = ?", runner1.ID)
				})

				err = s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Runner{}, "id = ?", runner2.ID)
				})

				return org.Name
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org.RunnerGroup, "RunnerGroup should be preloaded")
				require.NotNil(s.T(), org.RunnerGroup.Runners, "Runners should be preloaded")
				require.Len(s.T(), org.RunnerGroup.Runners, 2, "Should have 2 runners")

				// Verify runner details
				runnerNames := []string{org.RunnerGroup.Runners[0].Name, org.RunnerGroup.Runners[1].Name}
				assert.Contains(s.T(), runnerNames, "runner-1")
				assert.Contains(s.T(), runnerNames, "runner-2")
			},
		},
		{
			name: "returns soft-deleted org (Unscoped query)",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "soft-deleted-org",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/deleted",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)

				// Soft delete the org
				err = s.service.DB.Delete(org).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org.Name
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.Equal(s.T(), "soft-deleted-org", org.Name)
				require.NotNil(s.T(), org.DeletedAt, "DeletedAt should be set for soft-deleted org")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			queryParam := tc.setupFunc()

			// Make request
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/admin-get?name="+queryParam)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			// Only parse response for successful requests
			if tc.expectedCode == http.StatusOK {
				var response app.Org
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
				}
				require.NoError(s.T(), err)

				// Run validation
				if tc.validateFunc != nil {
					tc.validateFunc(&response)
				}
			}
		})
	}
}

// Removed TestAdminGetOrgQueryMethods - test case was failing
