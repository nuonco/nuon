package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// AdminUpdateOrgFeaturesTestService holds all fx-injected dependencies for admin update org features tests.
type AdminUpdateOrgFeaturesTestService struct {
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

// AdminUpdateOrgFeaturesTestSuite is the testify suite for the AdminUpdateOrgFeatures endpoint.
type AdminUpdateOrgFeaturesTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service AdminUpdateOrgFeaturesTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestAdminUpdateOrgFeaturesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AdminUpdateOrgFeaturesTestSuite))
}

func (s *AdminUpdateOrgFeaturesTestSuite) SetupSuite() {
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

func (s *AdminUpdateOrgFeaturesTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares (no org context for admin endpoints)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminUpdateOrgFeaturesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminUpdateOrgFeaturesTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())

	// Create test org with account context (required by BeforeCreate hook)
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	orgID := domains.NewOrgID()
	testOrg := &app.Org{
		ID:          orgID,
		Name:        fmt.Sprintf("admin-upd-feat-%s", orgID),
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/test",
		},
		Features: map[string]bool{
			string(app.OrgFeatureUserManagedFeatures): false, // Disabled by default
			string(app.OrgFeatureOrgDashboard):        false,
		},
	}
	err := s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
}

func (s *AdminUpdateOrgFeaturesTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminUpdateOrgFeaturesTestSuite) TestAdminUpdateOrgFeatures() {
	testCases := []struct {
		name          string
		setupFunc     func() *app.Org
		requestBody   interface{}
		expectedCode  int
		validateFunc  func(*app.Org)
		checkDBFunc   func(*app.Org)
		errorContains string
	}{
		{
			name: "successfully updates single feature",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-update-single-admin",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureOrgDashboard): false,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminUpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureOrgDashboard): true,
				},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.Equal(s.T(), "test-update-single-admin", org.Name)
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgDashboard)])
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.True(s.T(), dbOrg.Features[string(app.OrgFeatureOrgDashboard)])
			},
		},
		{
			name: "successfully updates multiple features",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-update-multiple-admin",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureOrgDashboard):  false,
						string(app.OrgFeatureAPIPagination): true,
						string(app.OrgFeatureOrgRunner):     false,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminUpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureOrgDashboard):  true,
					string(app.OrgFeatureAPIPagination): false,
					string(app.OrgFeatureOrgRunner):     true,
				},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgDashboard)])
				assert.False(s.T(), org.Features[string(app.OrgFeatureAPIPagination)])
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgRunner)])
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.True(s.T(), dbOrg.Features[string(app.OrgFeatureOrgDashboard)])
				assert.False(s.T(), dbOrg.Features[string(app.OrgFeatureAPIPagination)])
				assert.True(s.T(), dbOrg.Features[string(app.OrgFeatureOrgRunner)])
			},
		},
		{
			name: "successfully updates user-managed-features flag (admin privilege)",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-admin-managed-flag",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureUserManagedFeatures): false,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminUpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureUserManagedFeatures): true,
				},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.True(s.T(), org.Features[string(app.OrgFeatureUserManagedFeatures)])
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.True(s.T(), dbOrg.Features[string(app.OrgFeatureUserManagedFeatures)])
			},
		},
		{
			name: "successfully disables user-managed-features flag",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-admin-disable-flag",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureUserManagedFeatures): true,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminUpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureUserManagedFeatures): false,
				},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.False(s.T(), org.Features[string(app.OrgFeatureUserManagedFeatures)])
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.False(s.T(), dbOrg.Features[string(app.OrgFeatureUserManagedFeatures)])
			},
		},
		{
			name: "successfully updates all features at once",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-update-all-features",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureOrgDashboard):        false,
						string(app.OrgFeatureAPIPagination):       false,
						string(app.OrgFeatureOrgRunner):           false,
						string(app.OrgFeatureUserManagedFeatures): false,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminUpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureOrgDashboard):        true,
					string(app.OrgFeatureAPIPagination):       true,
					string(app.OrgFeatureOrgRunner):           true,
					string(app.OrgFeatureUserManagedFeatures): true,
					string(app.OrgFeatureOrgSettings):         true,
					string(app.OrgFeatureInstallBreakGlass):   true,
				},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgDashboard)])
				assert.True(s.T(), org.Features[string(app.OrgFeatureAPIPagination)])
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgRunner)])
				assert.True(s.T(), org.Features[string(app.OrgFeatureUserManagedFeatures)])
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgSettings)])
				assert.True(s.T(), org.Features[string(app.OrgFeatureInstallBreakGlass)])
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.True(s.T(), dbOrg.Features[string(app.OrgFeatureOrgDashboard)])
				assert.True(s.T(), dbOrg.Features[string(app.OrgFeatureUserManagedFeatures)])
			},
		},
		{
			name: "preserves unmodified features during update",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-preserve-features-admin",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureOrgDashboard):  true,
						string(app.OrgFeatureAPIPagination): false,
						string(app.OrgFeatureOrgRunner):     true,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminUpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureOrgDashboard): false, // Toggle this one
				},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				// Modified feature
				assert.False(s.T(), org.Features[string(app.OrgFeatureOrgDashboard)])
				// Unmodified features should be preserved
				assert.False(s.T(), org.Features[string(app.OrgFeatureAPIPagination)])
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgRunner)])
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				// Verify database state matches
				assert.False(s.T(), dbOrg.Features[string(app.OrgFeatureOrgDashboard)])
				assert.False(s.T(), dbOrg.Features[string(app.OrgFeatureAPIPagination)])
				assert.True(s.T(), dbOrg.Features[string(app.OrgFeatureOrgRunner)])
			},
		},
		{
			name: "toggles feature from true to false",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-toggle-false-admin",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureOrgDashboard): true,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminUpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureOrgDashboard): false,
				},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.False(s.T(), org.Features[string(app.OrgFeatureOrgDashboard)])
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.False(s.T(), dbOrg.Features[string(app.OrgFeatureOrgDashboard)])
			},
		},
		{
			name: "toggles feature from false to true",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-toggle-true-admin",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureOrgDashboard): false,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminUpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureOrgDashboard): true,
				},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgDashboard)])
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.True(s.T(), dbOrg.Features[string(app.OrgFeatureOrgDashboard)])
			},
		},
		{
			name: "handles empty features map",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-empty-features-admin",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureOrgDashboard): true,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminUpdateOrgFeaturesRequest{
				Features: map[string]bool{},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				// Empty map should not modify existing features
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgDashboard)])
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.True(s.T(), dbOrg.Features[string(app.OrgFeatureOrgDashboard)])
			},
		},
		{
			name: "fails with invalid JSON",
			setupFunc: func() *app.Org {
				return s.testOrg
			},
			requestBody:   "invalid-json-string",
			expectedCode:  http.StatusBadRequest,
			errorContains: "json:", // BindJSON returns JSON parsing errors with this prefix
		},
		{
			name: "fails when features field is missing",
			setupFunc: func() *app.Org {
				return s.testOrg
			},
			requestBody: map[string]interface{}{
				"wrong_field": "value",
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "invalid request", // BindJSON validation error
		},
		{
			name: "fails when org not found",
			setupFunc: func() *app.Org {
				// Return org with ID that doesn't exist
				return &app.Org{
					ID:          domains.NewOrgID(), // Non-existent org ID
					Name:        "nonexistent",
					SandboxMode: true,
				}
			},
			requestBody: AdminUpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureOrgDashboard): true,
				},
			},
			expectedCode:  http.StatusNotFound,
			errorContains: "org not found",
		},
		{
			name: "successfully updates features without user-managed-features flag enabled",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-no-user-managed-flag",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureUserManagedFeatures): false, // User flag disabled
						string(app.OrgFeatureOrgDashboard):        false,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminUpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureOrgDashboard): true,
				},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				// Admin can update even when user-managed-features is disabled
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgDashboard)])
				assert.False(s.T(), org.Features[string(app.OrgFeatureUserManagedFeatures)])
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.True(s.T(), dbOrg.Features[string(app.OrgFeatureOrgDashboard)])
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test org
			org := tc.setupFunc()
			require.NotNil(s.T(), org)

			// Make request with org_id path parameter
			path := "/v1/orgs/" + org.ID + "/admin-features"
			rr := s.makeRequest(http.MethodPatch, path, tc.requestBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			// For success cases
			if tc.expectedCode == http.StatusOK {
				// Parse response
				var response app.Org
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
				}
				require.NoError(s.T(), err)

				// Run validations
				if tc.validateFunc != nil {
					tc.validateFunc(&response)
				}

				// Check database state
				if tc.checkDBFunc != nil {
					tc.checkDBFunc(&response)
				}
			}

			// For error cases
			if tc.errorContains != "" {
				body := rr.Body.String()
				assert.Contains(s.T(), body, tc.errorContains,
					"Expected error message to contain: %s", tc.errorContains)
			}
		})
	}
}
