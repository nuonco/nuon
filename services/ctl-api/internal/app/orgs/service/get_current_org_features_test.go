package service

import (
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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/types"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// GetCurrentOrgFeaturesTestSuite is the testify suite for get current org features endpoint.
type GetCurrentOrgFeaturesTestSuite struct {
	tests.BaseDBTestSuite

	app         *fxtest.App
	service     TestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	orgsService *service
}

func TestGetCurrentOrgFeaturesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetCurrentOrgFeaturesTestSuite))
}

func (s *GetCurrentOrgFeaturesTestSuite) SetupSuite() {
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

func (s *GetCurrentOrgFeaturesTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Router setup is deferred to individual test cases
	// since we need different org contexts for different scenarios
}

func (s *GetCurrentOrgFeaturesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetCurrentOrgFeaturesTestSuite) setupTestData() {
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

	// Create basic test org with account context (required by BeforeCreate hook)
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

func (s *GetCurrentOrgFeaturesTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetCurrentOrgFeaturesTestSuite) TestGetCurrentOrgFeatures() {
	testCases := []struct {
		name             string
		setupFunc        func() *app.Org
		expectedStatus   int
		expectedFeatures map[string]bool
		validateFunc     func(map[string]bool)
	}{
		{
			name: "returns empty map when no features configured",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "org-no-features",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					// Features field intentionally omitted (nil)
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			expectedStatus:   http.StatusOK,
			expectedFeatures: map[string]bool{},
			validateFunc: func(features map[string]bool) {
				assert.Empty(s.T(), features, "Features map should be empty")
			},
		},
		{
			name: "returns empty map when features is empty map",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "org-empty-features",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: types.StringBoolMap{},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			expectedStatus:   http.StatusOK,
			expectedFeatures: map[string]bool{},
			validateFunc: func(features map[string]bool) {
				assert.Empty(s.T(), features, "Features map should be empty")
			},
		},
		{
			name: "returns all enabled features",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "org-enabled-features",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: types.StringBoolMap{
						"advanced_analytics": true,
						"api_access":         true,
						"custom_domains":     true,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			expectedStatus: http.StatusOK,
			expectedFeatures: map[string]bool{
				"advanced_analytics": true,
				"api_access":         true,
				"custom_domains":     true,
			},
			validateFunc: func(features map[string]bool) {
				assert.Len(s.T(), features, 3, "Should have 3 features")
				assert.True(s.T(), features["advanced_analytics"])
				assert.True(s.T(), features["api_access"])
				assert.True(s.T(), features["custom_domains"])
			},
		},
		{
			name: "returns all disabled features",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "org-disabled-features",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: types.StringBoolMap{
						"beta_features":     false,
						"experimental_apis": false,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			expectedStatus: http.StatusOK,
			expectedFeatures: map[string]bool{
				"beta_features":     false,
				"experimental_apis": false,
			},
			validateFunc: func(features map[string]bool) {
				assert.Len(s.T(), features, 2, "Should have 2 features")
				assert.False(s.T(), features["beta_features"])
				assert.False(s.T(), features["experimental_apis"])
			},
		},
		{
			name: "returns mixed enabled and disabled features",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "org-mixed-features",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: types.StringBoolMap{
						"advanced_analytics": true,
						"beta_features":      false,
						"api_access":         true,
						"experimental_apis":  false,
						"custom_domains":     true,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			expectedStatus: http.StatusOK,
			expectedFeatures: map[string]bool{
				"advanced_analytics": true,
				"beta_features":      false,
				"api_access":         true,
				"experimental_apis":  false,
				"custom_domains":     true,
			},
			validateFunc: func(features map[string]bool) {
				assert.Len(s.T(), features, 5, "Should have 5 features")

				// Validate enabled features
				assert.True(s.T(), features["advanced_analytics"])
				assert.True(s.T(), features["api_access"])
				assert.True(s.T(), features["custom_domains"])

				// Validate disabled features
				assert.False(s.T(), features["beta_features"])
				assert.False(s.T(), features["experimental_apis"])
			},
		},
		{
			name: "feature flag values preserved correctly across database round-trip",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org with specific feature values
				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "org-feature-persistence",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: types.StringBoolMap{
						"feature_a": true,
						"feature_b": false,
						"feature_c": true,
						"feature_d": false,
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				// Re-fetch from database to verify persistence
				var reloadedOrg app.Org
				err = s.service.DB.Where("id = ?", org.ID).First(&reloadedOrg).Error
				require.NoError(s.T(), err)

				// Verify database persistence
				assert.Equal(s.T(), org.Features, reloadedOrg.Features,
					"Features should persist correctly in database")

				return org
			},
			expectedStatus: http.StatusOK,
			expectedFeatures: map[string]bool{
				"feature_a": true,
				"feature_b": false,
				"feature_c": true,
				"feature_d": false,
			},
			validateFunc: func(features map[string]bool) {
				// Verify exact values are preserved
				for key, expectedValue := range map[string]bool{
					"feature_a": true,
					"feature_b": false,
					"feature_c": true,
					"feature_d": false,
				} {
					actualValue, exists := features[key]
					assert.True(s.T(), exists, "Feature %s should exist", key)
					assert.Equal(s.T(), expectedValue, actualValue,
						"Feature %s value should be preserved", key)
				}
			},
		},
		{
			name: "handles org with many features",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org with many features
				features := types.StringBoolMap{
					"analytics":          true,
					"api_v1":             true,
					"api_v2":             false,
					"beta_testing":       true,
					"custom_domains":     false,
					"data_export":        true,
					"debug_mode":         false,
					"enterprise_support": true,
					"experimental_ui":    false,
					"multi_region":       true,
					"premium_features":   true,
					"rbac":               true,
					"sso":                false,
					"webhooks":           true,
					"whitelabel":         false,
				}

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "org-many-features",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Features: features,
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(features map[string]bool) {
				assert.Len(s.T(), features, 15, "Should have 15 features")

				// Verify all expected features are present with correct values
				expectedFeatures := map[string]bool{
					"analytics":          true,
					"api_v1":             true,
					"api_v2":             false,
					"beta_testing":       true,
					"custom_domains":     false,
					"data_export":        true,
					"debug_mode":         false,
					"enterprise_support": true,
					"experimental_ui":    false,
					"multi_region":       true,
					"premium_features":   true,
					"rbac":               true,
					"sso":                false,
					"webhooks":           true,
					"whitelabel":         false,
				}

				for key, expectedValue := range expectedFeatures {
					actualValue, exists := features[key]
					assert.True(s.T(), exists, "Feature %s should exist", key)
					assert.Equal(s.T(), expectedValue, actualValue,
						"Feature %s should have value %v", key, expectedValue)
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			org := tc.setupFunc()

			// Create router with new org context
			// CRITICAL: Must recreate router for each test to capture correct org context
			s.router = tests.NewTestRouter(tests.RouterOptions{
				L:       s.service.L,
				DB:      s.service.DB,
				TestOrg: org,
				TestAcc: s.testAcc,
			})
			err := s.orgsService.RegisterPublicRoutes(s.router)
			require.NoError(s.T(), err)

			// Make request
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/current/features")

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Parse response
			var response map[string]bool
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)
			require.NotNil(s.T(), response)

			// Validate expected features if provided
			if tc.expectedFeatures != nil {
				assert.Equal(s.T(), tc.expectedFeatures, response,
					"Features should match expected values")
			}

			// Run additional validations if provided
			if tc.validateFunc != nil {
				tc.validateFunc(response)
			}
		})
	}
}

func (s *GetCurrentOrgFeaturesTestSuite) TestGetCurrentOrgFeaturesWithoutOrgContext() {
	// Create router without org context
	router := tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
		// TestOrg intentionally omitted
	})

	err := s.orgsService.RegisterPublicRoutes(router)
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, "/v1/orgs/current/features", nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should fail without org context
	require.NotEqual(s.T(), http.StatusOK, rr.Code)
	s.T().Logf("Status without org context: %d, Body: %s", rr.Code, rr.Body.String())
}

func (s *GetCurrentOrgFeaturesTestSuite) TestGetCurrentOrgFeaturesResponseFormat() {
	s.Run("response is valid JSON object", func() {
		ctx := context.Background()
		ctx = cctx.SetAccountContext(ctx, s.testAcc)

		org := &app.Org{
			ID:   domains.NewOrgID(),
			Name: "org-json-format",
			NotificationsConfig: app.NotificationsConfig{
				InternalSlackWebhookURL: "https://hooks.slack.com/test",
			},
			Features: types.StringBoolMap{
				"test_feature": true,
			},
		}
		err := s.service.DB.WithContext(ctx).Create(org).Error
		require.NoError(s.T(), err)
		s.T().Cleanup(func() {
			s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
		})

		// Create router with org context
		s.router = tests.NewTestRouter(tests.RouterOptions{
			L:       s.service.L,
			DB:      s.service.DB,
			TestOrg: org,
			TestAcc: s.testAcc,
		})
		err = s.orgsService.RegisterPublicRoutes(s.router)
		require.NoError(s.T(), err)

		rr := s.makeRequest(http.MethodGet, "/v1/orgs/current/features")
		require.Equal(s.T(), http.StatusOK, rr.Code)

		// Verify Content-Type header
		assert.Contains(s.T(), rr.Header().Get("Content-Type"), "application/json")

		// Verify response is valid JSON object (not array)
		var response map[string]bool
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err, "Response should be valid JSON object")

		// Verify features are present
		assert.Equal(s.T(), true, response["test_feature"])
	})
}
