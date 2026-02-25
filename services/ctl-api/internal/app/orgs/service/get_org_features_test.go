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

// GetOrgFeaturesTestService holds all fx-injected dependencies for GetOrgFeatures endpoint tests.
type GetOrgFeaturesTestService struct {
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

// GetOrgFeaturesTestSuite is the testify suite for GetOrgFeatures endpoint.
type GetOrgFeaturesTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service GetOrgFeaturesTestService
	router  *gin.Engine
	testAcc *app.Account
}

func TestGetOrgFeaturesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetOrgFeaturesTestSuite))
}

func (s *GetOrgFeaturesTestSuite) SetupSuite() {
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

func (s *GetOrgFeaturesTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares
	// Note: No TestOrg needed - GetOrgFeatures is a global endpoint
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetOrgFeaturesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetOrgFeaturesTestSuite) setupTestData() {
	ctx := context.Background()
	_, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
}

func (s *GetOrgFeaturesTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetOrgFeaturesTestSuite) TestGetOrgFeatures() {
	testCases := []struct {
		name             string
		validateFunc     func([]app.OrgFeatureInfo)
		expectedMinCount int // Minimum number of features we expect
	}{
		{
			name:             "returns non-empty feature list",
			expectedMinCount: 1,
			validateFunc: func(features []app.OrgFeatureInfo) {
				assert.NotEmpty(s.T(), features, "feature list should not be empty")
			},
		},
		{
			name:             "each feature has name and description",
			expectedMinCount: 1,
			validateFunc: func(features []app.OrgFeatureInfo) {
				for _, feature := range features {
					assert.NotEmpty(s.T(), feature.Name, "feature name should not be empty")
					assert.NotEmpty(s.T(), feature.Description, "feature description should not be empty")
					s.T().Logf("Feature: %s - %s", feature.Name, feature.Description)
				}
			},
		},
		{
			name:             "returns all defined features",
			expectedMinCount: 16, // Based on GetFeatures() returning 16 features
			validateFunc: func(features []app.OrgFeatureInfo) {
				expectedFeatures := app.GetFeatures()
				assert.Len(s.T(), features, len(expectedFeatures), "should return all defined features")

				// Verify all expected features are present
				featureNames := make(map[string]bool)
				for _, feature := range features {
					featureNames[feature.Name] = true
				}

				for _, expectedFeature := range expectedFeatures {
					assert.True(s.T(), featureNames[string(expectedFeature)],
						"expected feature %s should be present", expectedFeature)
				}
			},
		},
		{
			name:             "feature descriptions match expected mappings",
			expectedMinCount: 1,
			validateFunc: func(features []app.OrgFeatureInfo) {
				expectedDescriptions := app.GetFeatureDescriptions()

				for _, feature := range features {
					expectedDesc, exists := expectedDescriptions[app.OrgFeature(feature.Name)]
					assert.True(s.T(), exists, "feature %s should have description mapping", feature.Name)
					assert.Equal(s.T(), expectedDesc, feature.Description,
						"feature %s description should match expected value", feature.Name)
				}
			},
		},
		{
			name:             "returns consistent results across multiple calls",
			expectedMinCount: 1,
			validateFunc: func(features []app.OrgFeatureInfo) {
				// Make a second request
				rr2 := s.makeRequest(http.MethodGet, "/v1/orgs/features")
				require.Equal(s.T(), http.StatusOK, rr2.Code)

				var features2 []app.OrgFeatureInfo
				err := json.Unmarshal(rr2.Body.Bytes(), &features2)
				require.NoError(s.T(), err)

				// Results should be identical
				assert.Equal(s.T(), len(features), len(features2), "feature count should be consistent")
				assert.Equal(s.T(), features, features2, "feature lists should be identical")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Make request
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/features")

			// Log response for debugging
			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			// Parse response
			var features []app.OrgFeatureInfo
			err := json.Unmarshal(rr.Body.Bytes(), &features)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)
			require.NotNil(s.T(), features)

			// Validate minimum count
			assert.GreaterOrEqual(s.T(), len(features), tc.expectedMinCount,
				"should have at least %d features", tc.expectedMinCount)

			// Run additional validations if provided
			if tc.validateFunc != nil {
				tc.validateFunc(features)
			}
		})
	}
}

func (s *GetOrgFeaturesTestSuite) TestGetOrgFeaturesNoOrgContextRequired() {
	s.Run("succeeds without org context", func() {
		// Create router without TestOrg to verify endpoint works without org context
		routerNoOrg := tests.NewTestRouter(tests.RouterOptions{
			L:       s.service.L,
			DB:      s.service.DB,
			TestAcc: s.testAcc,
			// Explicitly no TestOrg
		})

		err := s.service.OrgsService.RegisterPublicRoutes(routerNoOrg)
		require.NoError(s.T(), err)

		req, err := http.NewRequest(http.MethodGet, "/v1/orgs/features", nil)
		require.NoError(s.T(), err)

		rr := httptest.NewRecorder()
		routerNoOrg.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var features []app.OrgFeatureInfo
		err = json.Unmarshal(rr.Body.Bytes(), &features)
		require.NoError(s.T(), err)
		assert.NotEmpty(s.T(), features, "should return features without org context")
	})
}

func (s *GetOrgFeaturesTestSuite) TestGetOrgFeaturesResponseStructure() {
	s.Run("response has correct JSON structure", func() {
		rr := s.makeRequest(http.MethodGet, "/v1/orgs/features")
		require.Equal(s.T(), http.StatusOK, rr.Code)

		// Verify response is valid JSON array
		var rawArray []json.RawMessage
		err := json.Unmarshal(rr.Body.Bytes(), &rawArray)
		require.NoError(s.T(), err)
		assert.NotEmpty(s.T(), rawArray, "response should be non-empty array")

		// Verify each element has required fields
		for i, raw := range rawArray {
			var feature map[string]interface{}
			err := json.Unmarshal(raw, &feature)
			require.NoError(s.T(), err, "feature %d should be valid JSON object", i)

			// Verify required fields exist
			name, hasName := feature["name"]
			assert.True(s.T(), hasName, "feature %d should have 'name' field", i)
			assert.IsType(s.T(), "", name, "name should be string")

			description, hasDesc := feature["description"]
			assert.True(s.T(), hasDesc, "feature %d should have 'description' field", i)
			assert.IsType(s.T(), "", description, "description should be string")
		}
	})
}

func (s *GetOrgFeaturesTestSuite) TestGetOrgFeaturesKnownFeatureFlags() {
	s.Run("includes known feature flags", func() {
		rr := s.makeRequest(http.MethodGet, "/v1/orgs/features")
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var features []app.OrgFeatureInfo
		err := json.Unmarshal(rr.Body.Bytes(), &features)
		require.NoError(s.T(), err)

		// Build map for easy lookup
		featureMap := make(map[string]string)
		for _, feature := range features {
			featureMap[feature.Name] = feature.Description
		}

		// Verify some known feature flags are present
		knownFeatures := []string{
			"api-pagination",
			"org-dashboard",
			"org-runner",
			"org-settings",
			"install-break-glass",
			"terraform-workspace",
			"stratus-layout",
			"stratus-workflow",
		}

		for _, knownFeature := range knownFeatures {
			description, exists := featureMap[knownFeature]
			assert.True(s.T(), exists, "known feature %s should be present", knownFeature)
			assert.NotEmpty(s.T(), description, "feature %s should have description", knownFeature)
		}
	})
}
