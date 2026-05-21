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
		validateResponseFunc  func([]app.OrgFeatureInfo)
		validateResponseCount func(int)
	}{
		{
			name:         "returns all available org features with descriptions",
			expectedCode: http.StatusOK,
			validateResponseFunc: func(features []app.OrgFeatureInfo) {
				expectedFeatures := app.GetFeatures()
				require.Len(s.T(), features, len(expectedFeatures))

				featureMap := make(map[string]bool)
				for _, feature := range features {
					featureMap[feature.Name] = true
				}

				for _, expectedFeature := range expectedFeatures {
					assert.True(s.T(), featureMap[string(expectedFeature)],
						"Expected feature %s to be in response", expectedFeature)
				}
			},
		},
		{
			name:         "returns consistent feature list across multiple calls",
			expectedCode: http.StatusOK,
			validateResponseFunc: func(features []app.OrgFeatureInfo) {
				rr2 := s.makeRequest(http.MethodGet, "/v1/orgs/admin-features")
				require.Equal(s.T(), http.StatusOK, rr2.Code)

				var features2 []app.OrgFeatureInfo
				err := json.Unmarshal(rr2.Body.Bytes(), &features2)
				require.NoError(s.T(), err)

				require.Equal(s.T(), features, features2,
					"Feature lists should be consistent across calls")
			},
		},
		{
			name:         "includes all documented features",
			expectedCode: http.StatusOK,
			validateResponseFunc: func(features []app.OrgFeatureInfo) {
				featureMap := make(map[string]bool)
				for _, feature := range features {
					featureMap[feature.Name] = true
				}

				criticalFeatures := []app.OrgFeature{
					app.OrgFeatureOrgDashboard,
					app.OrgFeatureOrgRunner,
					app.OrgFeatureOrgSettings,
					app.OrgFeatureAppBranches,
					app.OrgFeatureUserManagedFeatures,
				}

				for _, criticalFeature := range criticalFeatures {
					assert.True(s.T(), featureMap[string(criticalFeature)],
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
			validateResponseFunc: func(features []app.OrgFeatureInfo) {
				require.NotNil(s.T(), features,
					"Response should be a valid JSON array, not null")
			},
		},
		{
			name:         "all features have descriptions",
			expectedCode: http.StatusOK,
			validateResponseFunc: func(features []app.OrgFeatureInfo) {
				for _, feature := range features {
					assert.NotEmpty(s.T(), feature.Description,
						"Feature %s should have a description", feature.Name)
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/admin-features")

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			var features []app.OrgFeatureInfo
			err := json.Unmarshal(rr.Body.Bytes(), &features)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)

			if tc.validateResponseCount != nil {
				tc.validateResponseCount(len(features))
			}

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
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequest(tc.method, "/v1/orgs/admin-features")

			if rr.Code != tc.expectedCode {
				s.T().Logf("Method: %s, Status: %d, Body: %s",
					tc.method, rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var features []app.OrgFeatureInfo
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
			name: "returns array of objects with name and description",
			validateFunc: func(rr *httptest.ResponseRecorder) {
				var features []app.OrgFeatureInfo
				err := json.Unmarshal(rr.Body.Bytes(), &features)
				require.NoError(s.T(), err)

				for _, feature := range features {
					assert.NotEmpty(s.T(), feature.Name,
						"Each feature should have a non-empty name")
					assert.NotEmpty(s.T(), feature.Description,
						"Each feature should have a non-empty description")
				}
			},
		},
		{
			name: "feature names have expected string format",
			validateFunc: func(rr *httptest.ResponseRecorder) {
				var features []app.OrgFeatureInfo
				err := json.Unmarshal(rr.Body.Bytes(), &features)
				require.NoError(s.T(), err)

				for _, feature := range features {
					assert.Regexp(s.T(), `^[a-z]+(-[a-z0-9]+)*$`, feature.Name,
						"Feature %s should be in kebab-case format", feature.Name)
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

			tc.validateFunc(rr)
		})
	}
}

func (s *AdminGetOrgFeaturesTestSuite) TestAdminGetOrgFeaturesMatchesGetFeatures() {
	s.Run("API response matches app.GetFeaturesWithDescriptions()", func() {
		rr := s.makeRequest(http.MethodGet, "/v1/orgs/admin-features")
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var apiFeatures []app.OrgFeatureInfo
		err := json.Unmarshal(rr.Body.Bytes(), &apiFeatures)
		require.NoError(s.T(), err)

		functionFeatures := app.GetFeaturesWithDescriptions()

		require.Equal(s.T(), len(functionFeatures), len(apiFeatures),
			"API response count should match GetFeaturesWithDescriptions() count")

		apiFeatureMap := make(map[string]string)
		for _, feature := range apiFeatures {
			apiFeatureMap[feature.Name] = feature.Description
		}

		for _, funcFeature := range functionFeatures {
			desc, ok := apiFeatureMap[funcFeature.Name]
			assert.True(s.T(), ok,
				"Feature %s from GetFeaturesWithDescriptions() should be in API response", funcFeature.Name)
			assert.Equal(s.T(), funcFeature.Description, desc,
				"Description for feature %s should match", funcFeature.Name)
		}
	})
}
