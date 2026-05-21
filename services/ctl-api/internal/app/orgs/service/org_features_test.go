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

// OrgFeaturesTestService holds all fx-injected dependencies for org features endpoint tests.
type OrgFeaturesTestService struct {
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

// OrgFeaturesTestSuite is the testify suite for org features endpoints.
type OrgFeaturesTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service OrgFeaturesTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestOrgFeaturesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(OrgFeaturesTestSuite))
}

func (s *OrgFeaturesTestSuite) SetupSuite() {
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

func (s *OrgFeaturesTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *OrgFeaturesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *OrgFeaturesTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())

	// Create test org with account context (required by BeforeCreate hook)
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	orgID := domains.NewOrgID()
	testOrg := &app.Org{
		ID:          orgID,
		Name:        fmt.Sprintf("feat-%s", orgID),
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/test",
		},
		Features: map[string]bool{
			string(app.OrgFeatureUserManagedFeatures): false, // Disabled by default
		},
	}
	err := s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
}

func (s *OrgFeaturesTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

// TestGetOrgFeatures tests GET /v1/orgs/features (static feature list)
func (s *OrgFeaturesTestSuite) TestGetOrgFeatures() {
	testCases := []struct {
		name         string
		validateFunc func([]app.OrgFeatureInfo)
	}{
		{
			name: "returns all available features",
			validateFunc: func(features []app.OrgFeatureInfo) {
				// Verify we get all active features
				require.Len(s.T(), features, len(app.GetFeatures()))

				// Verify structure - each feature has name and description
				for _, feature := range features {
					assert.NotEmpty(s.T(), feature.Name)
					assert.NotEmpty(s.T(), feature.Description)
				}

				// Verify specific well-known features are present
				featureNames := make(map[string]bool)
				for _, feature := range features {
					featureNames[feature.Name] = true
				}

				assert.True(s.T(), featureNames[string(app.OrgFeatureUserManagedFeatures)])
				assert.True(s.T(), featureNames[string(app.OrgFeatureOrgDashboard)])
				assert.True(s.T(), featureNames[string(app.OrgFeatureAppBranches)])
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Make request (no org context required - static endpoint)
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/features", nil)

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			// Parse response
			var response []app.OrgFeatureInfo
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)
			require.NotNil(s.T(), response)

			// Run validations
			if tc.validateFunc != nil {
				tc.validateFunc(response)
			}
		})
	}
}

// TestGetCurrentOrgFeatures tests GET /v1/orgs/current/features
func (s *OrgFeaturesTestSuite) TestGetCurrentOrgFeatures() {
	testCases := []struct {
		name         string
		setupFunc    func() *app.Org
		expectedCode int
		validateFunc func(map[string]bool)
	}{
		{
			name: "returns org features map",
			setupFunc: func() *app.Org {
				// Return default test org
				return s.testOrg
			},
			expectedCode: http.StatusOK,
			validateFunc: func(features map[string]bool) {
				// Verify structure - should be map[string]bool
				require.NotNil(s.T(), features)

				// Verify user-managed-features flag exists and is false
				managed, ok := features[string(app.OrgFeatureUserManagedFeatures)]
				require.True(s.T(), ok)
				assert.False(s.T(), managed)

				// Verify all active features are present with default values
				for _, feature := range app.GetFeatures() {
					_, ok := features[string(feature)]
					assert.True(s.T(), ok, "feature %s should be present", feature)
				}
			},
		},
		{
			name: "returns empty map for org with no features",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org with nil features
				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "features-test-org-empty",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: nil,
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				// Update router with new org context
				s.router = tests.NewTestRouter(tests.RouterOptions{
					L:       s.service.L,
					DB:      s.service.DB,
					TestOrg: org,
					TestAcc: s.testAcc,
				})
				err = s.service.OrgsService.RegisterPublicRoutes(s.router)
				require.NoError(s.T(), err)

				return org
			},
			expectedCode: http.StatusOK,
			validateFunc: func(features map[string]bool) {
				// AfterQuery hook initializes empty features map
				require.NotNil(s.T(), features)
			},
		},
		{
			name: "returns org with custom feature flags",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org with custom features
				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "features-test-org-custom",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureUserManagedFeatures): true,
						string(app.OrgFeatureOrgDashboard):        true,
						string(app.OrgFeatureAppBranches):         false,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				// Update router with new org context
				s.router = tests.NewTestRouter(tests.RouterOptions{
					L:       s.service.L,
					DB:      s.service.DB,
					TestOrg: org,
					TestAcc: s.testAcc,
				})
				err = s.service.OrgsService.RegisterPublicRoutes(s.router)
				require.NoError(s.T(), err)

				return org
			},
			expectedCode: http.StatusOK,
			validateFunc: func(features map[string]bool) {
				// Verify custom values are preserved
				assert.True(s.T(), features[string(app.OrgFeatureUserManagedFeatures)])
				assert.True(s.T(), features[string(app.OrgFeatureOrgDashboard)])
				assert.False(s.T(), features[string(app.OrgFeatureAppBranches)])
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test org
			org := tc.setupFunc()
			require.NotNil(s.T(), org)

			// Make request
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/current/features", nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			// Parse response
			var response map[string]bool
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)

			// Run validations
			if tc.validateFunc != nil {
				tc.validateFunc(response)
			}
		})
	}
}

// TestUpdateOrgFeatures tests PATCH /v1/orgs/current/features
func (s *OrgFeaturesTestSuite) TestUpdateOrgFeatures() {
	testCases := []struct {
		name          string
		setupFunc     func() *app.Org
		requestBody   interface{}
		expectedCode  int
		validateFunc  func(*app.Org)
		checkDBFunc   func(*app.Org) // Check database state after update
		errorContains string         // Expected error message substring
	}{
		{
			name: "successfully updates user-manageable features",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org with user-managed-features enabled
				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "features-test-org-update-1",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureUserManagedFeatures): true, // ENABLED
						string(app.OrgFeatureOrgDashboard):        false,
						string(app.OrgFeatureAppBranches):         true,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				// Update router with new org context
				s.router = tests.NewTestRouter(tests.RouterOptions{
					L:       s.service.L,
					DB:      s.service.DB,
					TestOrg: org,
					TestAcc: s.testAcc,
				})
				err = s.service.OrgsService.RegisterPublicRoutes(s.router)
				require.NoError(s.T(), err)

				return org
			},
			requestBody: UpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureOrgDashboard): true,  // Toggle to true
					string(app.OrgFeatureAppBranches):  false, // Toggle to false
				},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				// Verify response is updated org
				assert.NotNil(s.T(), org)
				assert.Equal(s.T(), "features-test-org-update-1", org.Name)

				// Verify features were updated
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgDashboard)])
				assert.False(s.T(), org.Features[string(app.OrgFeatureAppBranches)])

				// Verify user-managed-features flag preserved
				assert.True(s.T(), org.Features[string(app.OrgFeatureUserManagedFeatures)])
			},
			checkDBFunc: func(org *app.Org) {
				// Verify database was actually updated
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)

				assert.True(s.T(), dbOrg.Features[string(app.OrgFeatureOrgDashboard)])
				assert.False(s.T(), dbOrg.Features[string(app.OrgFeatureAppBranches)])
			},
		},
		{
			name: "fails when user-managed-features flag NOT enabled",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org WITHOUT user-managed-features enabled
				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "features-test-org-update-2",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureUserManagedFeatures): false, // DISABLED
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				// Update router with new org context
				s.router = tests.NewTestRouter(tests.RouterOptions{
					L:       s.service.L,
					DB:      s.service.DB,
					TestOrg: org,
					TestAcc: s.testAcc,
				})
				err = s.service.OrgsService.RegisterPublicRoutes(s.router)
				require.NoError(s.T(), err)

				return org
			},
			requestBody: UpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureOrgDashboard): true,
				},
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "user-managed-features flag is not enabled",
		},
		{
			name: "fails when attempting to modify user-managed-features flag",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org with flag enabled
				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "features-test-org-update-3",
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

				// Update router with new org context
				s.router = tests.NewTestRouter(tests.RouterOptions{
					L:       s.service.L,
					DB:      s.service.DB,
					TestOrg: org,
					TestAcc: s.testAcc,
				})
				err = s.service.OrgsService.RegisterPublicRoutes(s.router)
				require.NoError(s.T(), err)

				return org
			},
			requestBody: UpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureUserManagedFeatures): false, // Try to disable itself
				},
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "cannot be modified through this endpoint",
		},
		{
			name: "fails when attempting to modify admin-only features",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org with flag enabled
				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "features-test-org-update-4",
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

				// Update router with new org context
				s.router = tests.NewTestRouter(tests.RouterOptions{
					L:       s.service.L,
					DB:      s.service.DB,
					TestOrg: org,
					TestAcc: s.testAcc,
				})
				err = s.service.OrgsService.RegisterPublicRoutes(s.router)
				require.NoError(s.T(), err)

				return org
			},
			requestBody: UpdateOrgFeaturesRequest{
				Features: map[string]bool{
					// This is the only admin-only feature (excluded by GetUserManageableFeatures)
					string(app.OrgFeatureUserManagedFeatures): true,
				},
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "cannot be modified through this endpoint",
		},
		{
			name: "fails with invalid request body",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org with flag enabled for test to reach parsing logic
				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "features-test-org-invalid-body",
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

				// Update router with new org context
				s.router = tests.NewTestRouter(tests.RouterOptions{
					L:       s.service.L,
					DB:      s.service.DB,
					TestOrg: org,
					TestAcc: s.testAcc,
				})
				err = s.service.OrgsService.RegisterPublicRoutes(s.router)
				require.NoError(s.T(), err)

				return org
			},
			requestBody:   map[string]interface{}{"invalid": "field"},
			expectedCode:  http.StatusBadRequest,
			errorContains: "invalid request",
		},
		{
			name: "updates multiple features at once",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org with flag enabled
				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "features-test-org-update-5",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureUserManagedFeatures): true,
						string(app.OrgFeatureOrgDashboard):        false,
						string(app.OrgFeatureAppBranches):         false,
						string(app.OrgFeatureOrgRunner):           false,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				// Update router with new org context
				s.router = tests.NewTestRouter(tests.RouterOptions{
					L:       s.service.L,
					DB:      s.service.DB,
					TestOrg: org,
					TestAcc: s.testAcc,
				})
				err = s.service.OrgsService.RegisterPublicRoutes(s.router)
				require.NoError(s.T(), err)

				return org
			},
			requestBody: UpdateOrgFeaturesRequest{
				Features: map[string]bool{
					string(app.OrgFeatureOrgDashboard): true,
					string(app.OrgFeatureAppBranches):  true,
					string(app.OrgFeatureOrgRunner):    true,
				},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				// Verify all three features were updated
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgDashboard)])
				assert.True(s.T(), org.Features[string(app.OrgFeatureAppBranches)])
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgRunner)])
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test org
			org := tc.setupFunc()
			require.NotNil(s.T(), org)

			// Make request
			rr := s.makeRequest(http.MethodPatch, "/v1/orgs/current/features", tc.requestBody)

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
