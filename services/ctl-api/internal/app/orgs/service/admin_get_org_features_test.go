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

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// AdminGetOrgFeaturesTestService holds all fx-injected dependencies.
type AdminGetOrgFeaturesTestService struct {
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

// AdminGetOrgFeaturesTestSuite is the testify suite for AdminGetOrgFeatures endpoint.
type AdminGetOrgFeaturesTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service AdminGetOrgFeaturesTestService
	router  *gin.Engine
	testAcc *app.Account
}

func TestAdminGetOrgFeaturesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AdminGetOrgFeaturesTestSuite))
}

func (s *AdminGetOrgFeaturesTestSuite) SetupSuite() {
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

func (s *AdminGetOrgFeaturesTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminGetOrgFeaturesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminGetOrgFeaturesTestSuite) setupTestData() {
	ctx := context.Background()
	_, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
}

func (s *AdminGetOrgFeaturesTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminGetOrgFeaturesTestSuite) TestAdminGetOrgFeatures() {
	testCases := []struct {
		name                  string
		expectedCode          int
		validateResponseFunc  func([]app.OrgFeature)
		validateResponseCount func(int)
	}{
		{
			name:         "returns all available org features",
			expectedCode: http.StatusOK,
			validateResponseFunc: func(features []app.OrgFeature) {
				// Verify all expected features are present
				expectedFeatures := app.GetFeatures()
				require.Len(s.T(), features, len(expectedFeatures))

				// Create a map for easier lookup
				featureMap := make(map[app.OrgFeature]bool)
				for _, feature := range features {
					featureMap[feature] = true
				}

				// Verify each expected feature is in the response
				for _, expectedFeature := range expectedFeatures {
					assert.True(s.T(), featureMap[expectedFeature],
						"Expected feature %s to be in response", expectedFeature)
				}
			},
		},
		{
			name:         "returns consistent feature list across multiple calls",
			expectedCode: http.StatusOK,
			validateResponseFunc: func(features []app.OrgFeature) {
				// Make a second request
				rr2 := s.makeRequest(http.MethodGet, "/v1/orgs/admin-features")
				require.Equal(s.T(), http.StatusOK, rr2.Code)

				var features2 []app.OrgFeature
				err := json.Unmarshal(rr2.Body.Bytes(), &features2)
				require.NoError(s.T(), err)

				// Verify both responses are identical
				require.Equal(s.T(), features, features2,
					"Feature lists should be consistent across calls")
			},
		},
		{
			name:         "includes all documented features",
			expectedCode: http.StatusOK,
			validateResponseFunc: func(features []app.OrgFeature) {
				// Verify specific critical features are present
				featureMap := make(map[app.OrgFeature]bool)
				for _, feature := range features {
					featureMap[feature] = true
				}

				criticalFeatures := []app.OrgFeature{
					app.OrgFeatureAPIPagination,
					app.OrgFeatureOrgDashboard,
					app.OrgFeatureOrgRunner,
					app.OrgFeatureOrgSettings,
					app.OrgFeatureOrgSupport,
					app.OrgFeatureInstallBreakGlass,
					app.OrgFeatureInstallDeleteComponents,
					app.OrgFeatureInstallDelete,
					app.OrgFeatureTerraformWorkspace,
					app.OrgFeatureDevCommand,
					app.OrgFeatureAppBranches,
					app.OrgFeatureStratusLayout,
					app.OrgFeatureStratusWorkflow,
					app.OrgFeatureTerraformInstaller,
					app.OrgFeatureDashboardSSE,
					app.OrgFeatureUserManagedFeatures,
				}

				for _, criticalFeature := range criticalFeatures {
					assert.True(s.T(), featureMap[criticalFeature],
						"Critical feature %s should be in response", criticalFeature)
				}
			},
		},
		{
			name:         "returns non-empty feature list",
			expectedCode: http.StatusOK,
			validateResponseCount: func(count int) {
				require.Greater(s.T(), count, 0,
					"Feature list should not be empty")
			},
		},
		{
			name:         "returns valid JSON array",
			expectedCode: http.StatusOK,
			validateResponseFunc: func(features []app.OrgFeature) {
				require.NotNil(s.T(), features,
					"Response should be a valid JSON array, not null")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Make request
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/admin-features")

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			// Parse response
			var features []app.OrgFeature
			err := json.Unmarshal(rr.Body.Bytes(), &features)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)

			// Run count validation if provided
			if tc.validateResponseCount != nil {
				tc.validateResponseCount(len(features))
			}

			// Run response validation if provided
			if tc.validateResponseFunc != nil {
				tc.validateResponseFunc(features)
			}
		})
	}
}

func (s *AdminGetOrgFeaturesTestSuite) TestAdminGetOrgFeaturesHTTPMethods() {
	testCases := []struct {
		name         string
		method       string
		expectedCode int
	}{
		{
			name:         "GET method is supported",
			method:       http.MethodGet,
			expectedCode: http.StatusOK,
		},
		{
			name:         "POST method is not supported",
			method:       http.MethodPost,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "PUT method is not supported",
			method:       http.MethodPut,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "DELETE method is not supported",
			method:       http.MethodDelete,
			expectedCode: http.StatusNotFound,
		},
		// Note: PATCH is actually supported on /admin-features (AdminUpdateOrgsFeatures endpoint)
		// so it's not included in the "not supported" test cases
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequest(tc.method, "/v1/orgs/admin-features")

			if rr.Code != tc.expectedCode {
				s.T().Logf("Method: %s, Status: %d, Body: %s",
					tc.method, rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			// For successful GET, verify response is valid JSON array
			if tc.expectedCode == http.StatusOK {
				var features []app.OrgFeature
				err := json.Unmarshal(rr.Body.Bytes(), &features)
				require.NoError(s.T(), err)
				require.NotNil(s.T(), features)
			}
		})
	}
}

func (s *AdminGetOrgFeaturesTestSuite) TestAdminGetOrgFeaturesResponseFormat() {
	testCases := []struct {
		name         string
		validateFunc func(*httptest.ResponseRecorder)
	}{
		{
			name: "returns valid content type",
			validateFunc: func(rr *httptest.ResponseRecorder) {
				contentType := rr.Header().Get("Content-Type")
				assert.Contains(s.T(), contentType, "application/json",
					"Content-Type should be application/json")
			},
		},
		{
			name: "returns array of strings",
			validateFunc: func(rr *httptest.ResponseRecorder) {
				var features []string
				err := json.Unmarshal(rr.Body.Bytes(), &features)
				require.NoError(s.T(), err)

				// Verify each feature is a non-empty string
				for _, feature := range features {
					assert.NotEmpty(s.T(), feature,
						"Each feature should be a non-empty string")
				}
			},
		},
		{
			name: "features have expected string format",
			validateFunc: func(rr *httptest.ResponseRecorder) {
				var features []string
				err := json.Unmarshal(rr.Body.Bytes(), &features)
				require.NoError(s.T(), err)

				// Verify features follow expected naming pattern (kebab-case)
				for _, feature := range features {
					// All features should contain hyphens (kebab-case)
					assert.Regexp(s.T(), `^[a-z]+(-[a-z]+)*$`, feature,
						"Feature %s should be in kebab-case format", feature)
				}
			},
		},
		{
			name: "response body is not empty",
			validateFunc: func(rr *httptest.ResponseRecorder) {
				assert.NotEmpty(s.T(), rr.Body.String(),
					"Response body should not be empty")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/admin-features")

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			// Run validation
			tc.validateFunc(rr)
		})
	}
}

func (s *AdminGetOrgFeaturesTestSuite) TestAdminGetOrgFeaturesMatchesGetFeatures() {
	s.Run("API response matches app.GetFeatures()", func() {
		// Make API request
		rr := s.makeRequest(http.MethodGet, "/v1/orgs/admin-features")
		require.Equal(s.T(), http.StatusOK, rr.Code)

		// Parse API response
		var apiFeatures []app.OrgFeature
		err := json.Unmarshal(rr.Body.Bytes(), &apiFeatures)
		require.NoError(s.T(), err)

		// Get features from function
		functionFeatures := app.GetFeatures()

		// Verify counts match
		require.Equal(s.T(), len(functionFeatures), len(apiFeatures),
			"API response count should match GetFeatures() count")

		// Create maps for comparison
		apiFeatureMap := make(map[app.OrgFeature]bool)
		for _, feature := range apiFeatures {
			apiFeatureMap[feature] = true
		}

		// Verify all function features are in API response
		for _, funcFeature := range functionFeatures {
			assert.True(s.T(), apiFeatureMap[funcFeature],
				"Feature %s from GetFeatures() should be in API response", funcFeature)
		}

		// Verify no extra features in API response
		functionFeatureMap := make(map[app.OrgFeature]bool)
		for _, feature := range functionFeatures {
			functionFeatureMap[feature] = true
		}

		for _, apiFeature := range apiFeatures {
			assert.True(s.T(), functionFeatureMap[apiFeature],
				"API response feature %s should exist in GetFeatures()", apiFeature)
		}
	})
}
