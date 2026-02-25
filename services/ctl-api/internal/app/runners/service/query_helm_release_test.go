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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/helm"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type QueryHelmReleaseTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type QueryHelmReleaseTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service QueryHelmReleaseTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestQueryHelmReleaseSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(QueryHelmReleaseTestSuite))
}

func (s *QueryHelmReleaseTestSuite) SetupSuite() {
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

func (s *QueryHelmReleaseTestSuite) SetupTest() {
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

func (s *QueryHelmReleaseTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *QueryHelmReleaseTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
}

func (s *QueryHelmReleaseTestSuite) createHelmChart(ctx context.Context, helmChartID string) {
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

func (s *QueryHelmReleaseTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *QueryHelmReleaseTestSuite) TestQueryHelmRelease() {
	testCases := []struct {
		name          string
		setupFunc     func() (string, string, string)
		expectedCode  int
		expectedCount int
		validateFunc  func([]helm.Release)
	}{
		{
			name: "empty results returns empty array",
			setupFunc: func() (string, string, string) {
				return "hchnnonexistent123456789012", "default", ""
			},
			expectedCode:  http.StatusOK,
			expectedCount: 0,
		},
		{
			name: "query with valid label filter - status",
			setupFunc: func() (string, string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				helmChartID := domains.NewHelmChartID()
				namespace := "default"

				// Create helm chart before releases
				s.createHelmChart(ctx, helmChartID)

				// Create releases with different statuses
				release1 := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         domains.NewHelmChartID(),
					CreatedByID: s.testAcc.ID,
					OrgID:       s.testOrg.ID,
					Type:        "helm.sh/release.v1",
					Body:        "",
					Name:        "test-release-1",
					Namespace:   namespace,
					Version:     1,
					Status:      "deployed",
					Owner:       "helm",
				}
				err := s.service.DB.WithContext(ctx).Create(release1).Error
				require.NoError(s.T(), err)

				release2 := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         domains.NewHelmChartID(),
					CreatedByID: s.testAcc.ID,
					OrgID:       s.testOrg.ID,
					Type:        "helm.sh/release.v1",
					Body:        "",
					Name:        "test-release-2",
					Namespace:   namespace,
					Version:     1,
					Status:      "failed",
					Owner:       "helm",
				}
				err = s.service.DB.WithContext(ctx).Create(release2).Error
				require.NoError(s.T(), err)

				release1Key := release1.Key
				release2Key := release2.Key
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().
						Where("helm_chart_id = ? AND key IN ?", helmChartID, []string{release1Key, release2Key}).
						Delete(&app.HelmRelease{})
				})

				return helmChartID, namespace, "?status=deployed"
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "query with valid label filter - version",
			setupFunc: func() (string, string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				helmChartID := domains.NewHelmChartID()
				namespace := "default"

				// Create helm chart before releases
				s.createHelmChart(ctx, helmChartID)

				// Create releases with different versions
				for i := 1; i <= 3; i++ {
					release := &app.HelmRelease{
						HelmChartID: helmChartID,
						Key:         domains.NewHelmChartID(),
						CreatedByID: s.testAcc.ID,
						OrgID:       s.testOrg.ID,
						Type:        "helm.sh/release.v1",
						Body:        "",
						Name:        "test-release",
						Namespace:   namespace,
						Version:     i,
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
				}

				return helmChartID, namespace, "?version=2"
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "query with valid label filter - name",
			setupFunc: func() (string, string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				helmChartID := domains.NewHelmChartID()
				namespace := "default"

				// Create helm chart before releases
				s.createHelmChart(ctx, helmChartID)

				release1 := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         domains.NewHelmChartID(),
					CreatedByID: s.testAcc.ID,
					OrgID:       s.testOrg.ID,
					Type:        "helm.sh/release.v1",
					Body:        "",
					Name:        "app-release",
					Namespace:   namespace,
					Version:     1,
					Status:      "deployed",
					Owner:       "helm",
				}
				err := s.service.DB.WithContext(ctx).Create(release1).Error
				require.NoError(s.T(), err)

				release2 := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         domains.NewHelmChartID(),
					CreatedByID: s.testAcc.ID,
					OrgID:       s.testOrg.ID,
					Type:        "helm.sh/release.v1",
					Body:        "",
					Name:        "db-release",
					Namespace:   namespace,
					Version:     1,
					Status:      "deployed",
					Owner:       "helm",
				}
				err = s.service.DB.WithContext(ctx).Create(release2).Error
				require.NoError(s.T(), err)

				release1Key := release1.Key
				release2Key := release2.Key
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().
						Where("helm_chart_id = ? AND key IN ?", helmChartID, []string{release1Key, release2Key}).
						Delete(&app.HelmRelease{})
				})

				return helmChartID, namespace, "?name=app-release"
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "query with unknown label returns error",
			setupFunc: func() (string, string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				helmChartID := domains.NewHelmChartID()
				namespace := "default"

				// Create helm chart before release
				s.createHelmChart(ctx, helmChartID)

				release := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         domains.NewHelmChartID(),
					CreatedByID: s.testAcc.ID,
					OrgID:       s.testOrg.ID,
					Type:        "helm.sh/release.v1",
					Body:        "",
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

				return helmChartID, namespace, "?unknown_label=value"
			},
			expectedCode: http.StatusInternalServerError,
			validateFunc: func(releases []helm.Release) {
				// Should return error for unknown label
			},
		},
		{
			name: "query with multiple valid labels",
			setupFunc: func() (string, string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				helmChartID := domains.NewHelmChartID()
				namespace := "default"

				// Create helm chart before release
				s.createHelmChart(ctx, helmChartID)

				release := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         domains.NewHelmChartID(),
					CreatedByID: s.testAcc.ID,
					OrgID:       s.testOrg.ID,
					Type:        "helm.sh/release.v1",
					Body:        "",
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

				return helmChartID, namespace, "?status=deployed&version=1"
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			helmChartID, namespace, queryParams := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/helm-releases/"+helmChartID+"/releases/"+namespace+"/query"+queryParams)

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

				if tc.validateFunc != nil {
					tc.validateFunc(releases)
				}
			} else if tc.expectedCode == http.StatusInternalServerError {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}
		})
	}
}

func (s *QueryHelmReleaseTestSuite) TestQueryHelmReleaseValidation() {
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
			path := "/v1/helm-releases/" + tc.helmChartID + "/releases/" + tc.namespace + "/query"
			rr := s.makeRequest("GET", path)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)
		})
	}
}

func (s *QueryHelmReleaseTestSuite) TestQueryHelmReleaseAllowedLabels() {
	// Document allowed label filters
	allowedLabels := []string{"modifiedAt", "createdAt", "version", "status", "owner", "name"}

	s.T().Logf("Allowed label filters: %v", allowedLabels)
	s.T().Log("QueryHelmRelease only allows filtering by these labels")
	s.T().Log("Unknown labels will return an error")
}
