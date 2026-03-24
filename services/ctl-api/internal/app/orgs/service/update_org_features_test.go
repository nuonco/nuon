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

// UpdateOrgFeaturesTestService holds all fx-injected dependencies for update org features tests.
type UpdateOrgFeaturesTestService struct {
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

// UpdateOrgFeaturesTestSuite is the testify suite for the UpdateOrgFeatures endpoint.
type UpdateOrgFeaturesTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service UpdateOrgFeaturesTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestUpdateOrgFeaturesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(UpdateOrgFeaturesTestSuite))
}

func (s *UpdateOrgFeaturesTestSuite) SetupSuite() {
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

func (s *UpdateOrgFeaturesTestSuite) SetupTest() {
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

func (s *UpdateOrgFeaturesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *UpdateOrgFeaturesTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())

	// Create test org with account context (required by BeforeCreate hook)
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	orgID := domains.NewOrgID()
	testOrg := &app.Org{
		ID:          orgID,
		Name:        fmt.Sprintf("upd-feat-%s", orgID),
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

func (s *UpdateOrgFeaturesTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *UpdateOrgFeaturesTestSuite) TestUpdateOrgFeatures() {
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
			name: "successfully updates single user-manageable feature",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-update-single",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureUserManagedFeatures): true,
						string(app.OrgFeatureOrgDashboard):        false,
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
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.Equal(s.T(), "test-update-single", org.Name)
				assert.True(s.T(), org.Features[string(app.OrgFeatureOrgDashboard)])
				assert.True(s.T(), org.Features[string(app.OrgFeatureUserManagedFeatures)])
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.True(s.T(), dbOrg.Features[string(app.OrgFeatureOrgDashboard)])
			},
		},
		{
			name: "successfully updates multiple user-manageable features",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-update-multiple",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureUserManagedFeatures): true,
						string(app.OrgFeatureOrgDashboard):        false,
						string(app.OrgFeatureAPIPagination):       true,
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
				assert.True(s.T(), org.Features[string(app.OrgFeatureUserManagedFeatures)])
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
			name: "fails when user-managed-features flag not enabled",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-flag-disabled",
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
			name: "fails when user-managed-features flag missing",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-flag-missing",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{},
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
			name: "rejects attempt to modify user-managed-features flag itself",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-modify-self",
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
					string(app.OrgFeatureUserManagedFeatures): false,
				},
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "cannot be modified through this endpoint",
		},
		{
			name: "rejects attempt to enable user-managed-features flag",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-enable-self",
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
					string(app.OrgFeatureUserManagedFeatures): true,
				},
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "cannot be modified through this endpoint",
		},
		{
			name: "rejects non-user-manageable features (admin-only)",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-admin-only",
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
					// user-managed-features is excluded from GetUserManageableFeatures()
					string(app.OrgFeatureUserManagedFeatures): true,
				},
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "cannot be modified through this endpoint",
		},
		{
			name: "validation error when features map is missing",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-missing-map",
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
			requestBody: map[string]interface{}{
				"wrong_field": "value",
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "invalid request",
		},
		{
			name: "validation error when features field is null",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-null-features",
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
			requestBody: map[string]interface{}{
				"features": nil,
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "invalid request",
		},
		{
			name: "invalid JSON handling",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-invalid-json",
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
			requestBody:   "invalid-json-string",
			expectedCode:  http.StatusBadRequest,
			errorContains: "json:", // JSON parsing error message contains "json:" prefix
		},
		{
			name: "rejects invalid feature names",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-invalid-feature",
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
					"invalid-feature-name": true,
				},
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "is not user-manageable",
		},
		{
			name: "preserves unmodified features during update",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-preserve-features",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureUserManagedFeatures): true,
						string(app.OrgFeatureOrgDashboard):        true,
						string(app.OrgFeatureAPIPagination):       false,
						string(app.OrgFeatureOrgRunner):           true,
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
				assert.True(s.T(), org.Features[string(app.OrgFeatureUserManagedFeatures)])
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
					Name:        "test-toggle-false",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureUserManagedFeatures): true,
						string(app.OrgFeatureOrgDashboard):        true,
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
					Name:        "test-toggle-true",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: map[string]bool{
						string(app.OrgFeatureUserManagedFeatures): true,
						string(app.OrgFeatureOrgDashboard):        false,
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
