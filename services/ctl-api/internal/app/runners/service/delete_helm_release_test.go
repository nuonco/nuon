package service

import (
	"context"
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
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type DeleteHelmReleaseTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type DeleteHelmReleaseTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service DeleteHelmReleaseTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestDeleteHelmReleaseSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(DeleteHelmReleaseTestSuite))
}

func (s *DeleteHelmReleaseTestSuite) SetupSuite() {
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

func (s *DeleteHelmReleaseTestSuite) SetupTest() {
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

func (s *DeleteHelmReleaseTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *DeleteHelmReleaseTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
}

func (s *DeleteHelmReleaseTestSuite) createHelmChart(ctx context.Context, helmChartID string) {
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

func (s *DeleteHelmReleaseTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *DeleteHelmReleaseTestSuite) TestDeleteHelmRelease() {
	testCases := []struct {
		name         string
		setupFunc    func() (string, string, string)
		expectedCode int
		validateFunc func(string, string, string)
	}{
		{
			name: "successfully delete existing release",
			setupFunc: func() (string, string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				helmChartID := domains.NewHelmChartID()
				s.createHelmChart(ctx, helmChartID)

				namespace := "default"
				key := "sh.helm.release.v1.test-release.v1"

				release := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         key,
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

				return helmChartID, namespace, key
			},
			expectedCode: http.StatusOK,
			validateFunc: func(helmChartID, namespace, key string) {
				// Verify release is deleted (soft delete)
				var release app.HelmRelease
				err := s.service.DB.Unscoped().
					Where("helm_chart_id = ? AND namespace = ? AND key = ?", helmChartID, namespace, key).
					First(&release).Error
				assert.NoError(s.T(), err)
				assert.NotZero(s.T(), release.DeletedAt)
			},
		},
		{
			name: "delete nonexistent release returns 200 (no rows affected)",
			setupFunc: func() (string, string, string) {
				return "hchnonexistent123456789012", "default", "nonexistent-key"
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "delete with wrong namespace returns 200 (no rows affected)",
			setupFunc: func() (string, string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				helmChartID := domains.NewHelmChartID()
				s.createHelmChart(ctx, helmChartID)

				key := "sh.helm.release.v1.test-release.v1"

				release := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         key,
					CreatedByID: s.testAcc.ID,
					OrgID:       s.testOrg.ID,
					Type:        "helm.sh/release.v1",
					Body:        "",
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

				// Try to delete with wrong namespace
				return helmChartID, "staging", key
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "delete with wrong key returns 200 (no rows affected)",
			setupFunc: func() (string, string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				helmChartID := domains.NewHelmChartID()
				s.createHelmChart(ctx, helmChartID)

				namespace := "default"

				release := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         "sh.helm.release.v1.test-release.v1",
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

				// Try to delete with wrong key
				return helmChartID, namespace, "sh.helm.release.v1.different-release.v1"
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "delete already deleted release returns 200 (no rows affected)",
			setupFunc: func() (string, string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				helmChartID := domains.NewHelmChartID()
				s.createHelmChart(ctx, helmChartID)

				namespace := "default"
				key := "sh.helm.release.v1.test-release.v1"

				release := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         key,
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

				// Soft delete the release
				err = s.service.DB.WithContext(ctx).
					Where("helm_chart_id = ? AND namespace = ? AND key = ?", helmChartID, namespace, key).
					Delete(&app.HelmRelease{}).Error
				require.NoError(s.T(), err)

				releaseKey := release.Key
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().
						Where("helm_chart_id = ? AND key = ?", helmChartID, releaseKey).
						Delete(&app.HelmRelease{})
				})

				return helmChartID, namespace, key
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "delete multiple releases with same helm_chart_id but different keys",
			setupFunc: func() (string, string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				helmChartID := domains.NewHelmChartID()
				s.createHelmChart(ctx, helmChartID)

				namespace := "default"

				// Create first release
				release1 := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         "sh.helm.release.v1.release-1.v1",
					CreatedByID: s.testAcc.ID,
					OrgID:       s.testOrg.ID,
					Type:        "helm.sh/release.v1",
					Body:        "",
					Name:        "release-1",
					Namespace:   namespace,
					Version:     1,
					Status:      "deployed",
					Owner:       "helm",
				}
				err := s.service.DB.WithContext(ctx).Create(release1).Error
				require.NoError(s.T(), err)

				// Create second release
				release2 := &app.HelmRelease{
					HelmChartID: helmChartID,
					Key:         "sh.helm.release.v1.release-2.v1",
					CreatedByID: s.testAcc.ID,
					OrgID:       s.testOrg.ID,
					Type:        "helm.sh/release.v1",
					Body:        "",
					Name:        "release-2",
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

				// Return first release to delete
				return helmChartID, namespace, release1.Key
			},
			expectedCode: http.StatusOK,
			validateFunc: func(helmChartID, namespace, key string) {
				// Verify first release is deleted
				var release1 app.HelmRelease
				err := s.service.DB.Unscoped().
					Where("helm_chart_id = ? AND namespace = ? AND key = ?", helmChartID, namespace, key).
					First(&release1).Error
				assert.NoError(s.T(), err)
				assert.NotZero(s.T(), release1.DeletedAt)

				// Verify second release still exists
				var release2 app.HelmRelease
				err = s.service.DB.
					Where("helm_chart_id = ? AND namespace = ? AND key = ?", helmChartID, namespace, "sh.helm.release.v1.release-2.v1").
					First(&release2).Error
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), "release-2", release2.Name)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			helmChartID, namespace, key := tc.setupFunc()
			rr := s.makeRequest("DELETE", "/v1/helm-releases/"+helmChartID+"/releases/"+namespace+"/"+key)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc(helmChartID, namespace, key)
			}
		})
	}
}

func (s *DeleteHelmReleaseTestSuite) TestDeleteHelmReleaseSoftDelete() {
	// Verify that delete is a soft delete (sets DeletedAt, doesn't remove from DB)

	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	helmChartID := domains.NewHelmChartID()
	s.createHelmChart(ctx, helmChartID)

	namespace := "default"
	key := "sh.helm.release.v1.test-release.v1"

	release := &app.HelmRelease{
		HelmChartID: helmChartID,
		Key:         key,
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

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().
			Where("helm_chart_id = ? AND key = ?", helmChartID, key).
			Delete(&app.HelmRelease{})
	})

	// Delete the release
	rr := s.makeRequest("DELETE", "/v1/helm-releases/"+helmChartID+"/releases/"+namespace+"/"+key)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Verify soft delete - record still exists but DeletedAt is set
	var deletedRelease app.HelmRelease
	err = s.service.DB.Unscoped().
		Where("helm_chart_id = ? AND namespace = ? AND key = ?", helmChartID, namespace, key).
		First(&deletedRelease).Error
	require.NoError(s.T(), err)

	assert.NotZero(s.T(), deletedRelease.DeletedAt, "DeletedAt should be set for soft delete")
	assert.Equal(s.T(), "test-release", deletedRelease.Name, "Record should still exist in database")

	// Verify release cannot be found with normal query
	var normalQuery app.HelmRelease
	err = s.service.DB.
		Where("helm_chart_id = ? AND namespace = ? AND key = ?", helmChartID, namespace, key).
		First(&normalQuery).Error
	assert.Error(s.T(), err, "Deleted release should not be found in normal query")
	assert.Equal(s.T(), gorm.ErrRecordNotFound, err)
}
