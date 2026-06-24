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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/helm"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type GetHelmReleaseTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetHelmReleaseTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service GetHelmReleaseTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestGetHelmReleaseSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetHelmReleaseTestSuite))
}

func (s *GetHelmReleaseTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *GetHelmReleaseTestSuite) SetupTest() {
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

func (s *GetHelmReleaseTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetHelmReleaseTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
}

func (s *GetHelmReleaseTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetHelmReleaseTestSuite) createHelmChart(ctx context.Context, helmChartID string) {
	chart := &app.HelmChart{
		ID:          helmChartID,
		CreatedByID: s.testAcc.ID,
		OrgID:       s.testOrg.ID,
		OwnerID:     domains.NewInstallID(),
		OwnerType:   "install",
	}
	err := s.service.DB.WithContext(ctx).Create(chart).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(chart)
	})
}

func (s *GetHelmReleaseTestSuite) TestGetHelmRelease() {
	testCases := []struct {
		name         string
		setupFunc    func() (string, string, string)
		expectedCode int
		validateFunc func(*helm.Release)
	}{
		{
			name: "release not found with empty body returns error",
			setupFunc: func() (string, string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				helmChartID := domains.NewHelmChartID()
				namespace := "default"
				key := "sh.helm.release.v1.test-release.v1"

				// Create helm chart first to satisfy FK constraint
				s.createHelmChart(ctx, helmChartID)

				// Create release with empty body - decode will fail
				release := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         key,
					CreatedByID: s.testAcc.ID,
					OrgID:       s.testOrg.ID,
					Type:        "helm.sh/release.v1",
					Body:        &blobstore.Blob{},
					Name:        "test-release",
					Namespace:   namespace,
					Version:     1,
					Status:      "deployed",
					Owner:       "helm",
				}
				err := s.service.DB.WithContext(ctx).Create(release).Error
				require.NoError(s.T(), err)

				releaseKey := release.Key
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().
						Where("helm_chart_id = ? AND key = ?", helmChartID, releaseKey).
						Delete(&app.HelmRelease{})
				})

				return helmChartID, namespace, key
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "nonexistent release returns error",
			setupFunc: func() (string, string, string) {
				return "hchnonexistent123456789012", "default", "nonexistent-key"
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "different namespace returns error",
			setupFunc: func() (string, string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				helmChartID := domains.NewHelmChartID()
				key := "sh.helm.release.v1.test-release.v1"

				// Create helm chart first to satisfy FK constraint
				s.createHelmChart(ctx, helmChartID)

				release := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         key,
					CreatedByID: s.testAcc.ID,
					OrgID:       s.testOrg.ID,
					Type:        "helm.sh/release.v1",
					Body:        &blobstore.Blob{},
					Name:        "test-release",
					Namespace:   "production",
					Version:     1,
					Status:      "deployed",
					Owner:       "helm",
				}
				err := s.service.DB.WithContext(ctx).Create(release).Error
				require.NoError(s.T(), err)

				releaseKey := release.Key
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().
						Where("helm_chart_id = ? AND key = ?", helmChartID, releaseKey).
						Delete(&app.HelmRelease{})
				})

				// Query different namespace
				return helmChartID, "staging", key
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "different key returns error",
			setupFunc: func() (string, string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				helmChartID := domains.NewHelmChartID()
				namespace := "default"

				// Create helm chart first to satisfy FK constraint
				s.createHelmChart(ctx, helmChartID)

				release := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         "sh.helm.release.v1.test-release.v1",
					CreatedByID: s.testAcc.ID,
					OrgID:       s.testOrg.ID,
					Type:        "helm.sh/release.v1",
					Body:        &blobstore.Blob{},
					Name:        "test-release",
					Namespace:   namespace,
					Version:     1,
					Status:      "deployed",
					Owner:       "helm",
				}
				err := s.service.DB.WithContext(ctx).Create(release).Error
				require.NoError(s.T(), err)

				releaseKey := release.Key
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().
						Where("helm_chart_id = ? AND key = ?", helmChartID, releaseKey).
						Delete(&app.HelmRelease{})
				})

				// Query different key
				return helmChartID, namespace, "sh.helm.release.v1.different-release.v1"
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			helmChartID, namespace, key := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/helm-releases/"+helmChartID+"/releases/"+namespace+"/"+key)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var release helm.Release
				err := json.Unmarshal(rr.Body.Bytes(), &release)
				require.NoError(s.T(), err)
				tc.validateFunc(&release)
			} else if tc.expectedCode == http.StatusInternalServerError {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}
		})
	}
}

func (s *GetHelmReleaseTestSuite) TestGetHelmReleaseValidation() {
	testCases := []struct {
		name         string
		helmChartID  string
		namespace    string
		key          string
		expectedCode int
	}{
		{
			name:         "missing helm_chart_id returns 404",
			helmChartID:  "",
			namespace:    "default",
			key:          "key",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "missing namespace returns 404",
			helmChartID:  "hchvalid123456789012345678",
			namespace:    "",
			key:          "key",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "missing key returns 301 redirect",
			helmChartID:  "hchvalid123456789012345678",
			namespace:    "default",
			key:          "",
			expectedCode: http.StatusMovedPermanently,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := "/v1/helm-releases/" + tc.helmChartID + "/releases/" + tc.namespace + "/" + tc.key
			rr := s.makeRequest("GET", path)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)
		})
	}
}
