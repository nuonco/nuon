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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/helm"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type GetHelmReleasesTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetHelmReleasesTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service GetHelmReleasesTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestGetHelmReleasesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetHelmReleasesTestSuite))
}

func (s *GetHelmReleasesTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *GetHelmReleasesTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with runner routes
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.RunnersService.RegisterRunnerRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetHelmReleasesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetHelmReleasesTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
}

func (s *GetHelmReleasesTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetHelmReleasesTestSuite) TestGetHelmReleases() {
	testCases := []struct {
		name          string
		setupFunc     func() (string, string)
		expectedCode  int
		expectedCount int
	}{
		{
			name: "empty results returns empty array",
			setupFunc: func() (string, string) {
				return "hchnnonexistent123456789012", "default"
			},
			expectedCode:  http.StatusOK,
			expectedCount: 0,
		},
		{
			name: "nonexistent helm_chart_id returns empty array",
			setupFunc: func() (string, string) {
				return domains.NewHelmChartID(), "default"
			},
			expectedCode:  http.StatusOK,
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			helmChartID, namespace := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/helm-releases/"+helmChartID+"/releases/"+namespace)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var releases []helm.Release
				err := json.Unmarshal(rr.Body.Bytes(), &releases)
				require.NoError(s.T(), err)

				if tc.expectedCount >= 0 {
					assert.Len(s.T(), releases, tc.expectedCount)
				}
			}
		})
	}
}

func (s *GetHelmReleasesTestSuite) TestGetHelmReleasesValidation() {
	testCases := []struct {
		name         string
		helmChartID  string
		namespace    string
		expectedCode int
	}{
		{
			name:         "missing helm_chart_id returns 404",
			helmChartID:  "",
			namespace:    "default",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "missing namespace returns 404",
			helmChartID:  "hchvalid123456789012345678",
			namespace:    "",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := "/v1/helm-releases/" + tc.helmChartID + "/releases/" + tc.namespace
			rr := s.makeRequest("GET", path)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)
		})
	}
}

func (s *GetHelmReleasesTestSuite) TestGetHelmReleasesWithReleases() {
	// Document that testing with actual helm releases requires valid encoded data
	s.T().Log("Testing with actual helm releases requires:")
	s.T().Log("1. Valid helm chart records (foreign key)")
	s.T().Log("2. Valid encoded helm.Release data (Body field)")
	s.T().Log("3. helm.EncodeRelease does protobuf + gzip + base64 encoding")
	s.T().Log("4. DecodeRelease will fail if Body is empty or invalid")
	s.T().Log("")
	s.T().Log("In production, helm releases are created by actual Helm operations")
	s.T().Log("These tests focus on endpoint behavior with empty/nonexistent data")
}
