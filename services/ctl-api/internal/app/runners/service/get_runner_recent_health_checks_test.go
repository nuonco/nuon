package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

type GetRunnerRecentHealthChecksTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetRunnerRecentHealthChecksTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       GetRunnerRecentHealthChecksTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
}

func TestGetRunnerRecentHealthChecksSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetRunnerRecentHealthChecksTestSuite))
}

func (s *GetRunnerRecentHealthChecksTestSuite) SetupSuite() {
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
	s.SetCHDB(s.service.CHDB)
}

func (s *GetRunnerRecentHealthChecksTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with public routes
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetRunnerRecentHealthChecksTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetRunnerRecentHealthChecksTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create runner group
	s.testRunnerGrp = &app.RunnerGroup{
		ID:        domains.NewRunnerGroupID(),
		OrgID:     s.testOrg.ID,
		OwnerID:   s.testOrg.ID,
		OwnerType: "org",
		Type:      app.RunnerGroupTypeOrg,
		Platform:  app.AppRunnerTypeAWSEKS,
	}
	err := s.service.DB.WithContext(ctx).Create(s.testRunnerGrp).Error
	require.NoError(s.T(), err)

	// Create runner
	s.testRunner = &app.Runner{
		ID:            domains.NewRunnerID(),
		OrgID:         s.testOrg.ID,
		Name:          "test-runner",
		DisplayName:   "Test Runner",
		Status:        app.RunnerStatusActive,
		RunnerGroupID: s.testRunnerGrp.ID,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunner).Error
	require.NoError(s.T(), err)
}

func (s *GetRunnerRecentHealthChecksTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetRunnerRecentHealthChecksTestSuite) createHealthChecks(runnerID string, count int, baseTimestamp time.Time, status app.RunnerStatus) []string {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	ctx = cctx.SetOrgContext(ctx, s.testOrg)

	healthCheckIDs := make([]string, count)
	for i := 0; i < count; i++ {
		healthCheck := &app.RunnerHealthCheck{
			ID:           domains.NewRunnerHealthCheckID(),
			RunnerID:     runnerID,
			RunnerStatus: status,
			// Space health checks by 1 minute so each falls in a different minute bucket.
			// The runner_health_checks_view_v1 deduplicates to 1 row per (runner_id, minute_bucket),
			// so checks within the same minute would be collapsed into a single row.
			CreatedAt: baseTimestamp.Add(time.Duration(i) * time.Minute),
		}
		err := s.service.CHDB.WithContext(ctx).Create(healthCheck).Error
		require.NoError(s.T(), err)
		healthCheckIDs[i] = healthCheck.ID
	}

	// Allow ClickHouse time to flush and make the inserted rows visible to subsequent queries.
	time.Sleep(500 * time.Millisecond)

	return healthCheckIDs
}

func (s *GetRunnerRecentHealthChecksTestSuite) TestGetRunnerRecentHealthChecks() {
	testCases := []struct {
		name          string
		setupFunc     func() string
		queryParams   string
		expectedCode  int
		expectedCount int
		validateFunc  func([]*app.RunnerHealthCheck)
	}{
		{
			name: "empty health checks returns empty array",
			setupFunc: func() string {
				// Use existing test runner with no health checks
				return s.testRunner.ID
			},
			expectedCode:  http.StatusOK,
			expectedCount: 0,
		},
		{
			name: "returns health checks for correct runner",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create new runner
				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "runner-with-checks",
					DisplayName:   "Runner With Checks",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				// Create 5 health check records within 1 hour window
				s.createHealthChecks(runner.ID, 5, time.Now().Add(-30*time.Minute), app.RunnerStatusActive)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			expectedCode:  http.StatusOK,
			expectedCount: 5,
			validateFunc: func(healthChecks []*app.RunnerHealthCheck) {
				assert.Len(s.T(), healthChecks, 5)
				// Verify ascending order (default)
				for i := 0; i < len(healthChecks)-1; i++ {
					assert.True(s.T(), healthChecks[i].CreatedAt.Before(healthChecks[i+1].CreatedAt) || healthChecks[i].CreatedAt.Equal(healthChecks[i+1].CreatedAt))
				}
				// Verify all have Active status
				for _, hc := range healthChecks {
					assert.Equal(s.T(), app.RunnerStatusActive, hc.RunnerStatus)
					assert.Equal(s.T(), 0, hc.RunnerStatusCode)
				}
			},
		},
		{
			name: "does not return health checks from different runner",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create two runners
				runner1 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "runner-one",
					DisplayName:   "Runner One",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err := s.service.DB.WithContext(ctx).Create(runner1).Error
				require.NoError(s.T(), err)

				runner2 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "runner-two",
					DisplayName:   "Runner Two",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err = s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)

				// Create health checks in both runners
				s.createHealthChecks(runner1.ID, 3, time.Now().Add(-30*time.Minute), app.RunnerStatusActive)
				s.createHealthChecks(runner2.ID, 2, time.Now().Add(-20*time.Minute), app.RunnerStatusActive)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner1)
					s.service.DB.Unscoped().Delete(runner2)
				})

				// Request health checks for runner1, should only get 3
				return runner1.ID
			},
			expectedCode:  http.StatusOK,
			expectedCount: 3,
			validateFunc: func(healthChecks []*app.RunnerHealthCheck) {
				assert.Len(s.T(), healthChecks, 3)
			},
		},
		{
			name: "runner not found returns 404",
			setupFunc: func() string {
				return "runnonexistent123456789012"
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "runner in different org returns 404",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create second org
				org2ID := domains.NewOrgID()
				org2 := &app.Org{
					ID:          org2ID,
					Name:        fmt.Sprintf("other-org-%s", org2ID),
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/bar",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)

				// Create runner group in org2
				runnerGrp2 := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OrgID:     org2.ID,
					OwnerID:   org2.ID,
					OwnerType: "org",
					Type:      app.RunnerGroupTypeOrg,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err = s.service.DB.WithContext(ctx).Create(runnerGrp2).Error
				require.NoError(s.T(), err)

				// Create runner in org2
				runner2 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         org2.ID,
					Name:          "runner-in-org2",
					DisplayName:   "Runner in Org2",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: runnerGrp2.ID,
				}
				err = s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner2)
					s.service.DB.Unscoped().Delete(runnerGrp2)
					s.service.DB.Unscoped().Delete(org2)
				})

				return runner2.ID
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "window parameter filters health checks by time",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "runner-time-filter",
					DisplayName:   "Runner Time Filter",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				// Create health checks at different times
				// 3 checks within 10 minutes
				s.createHealthChecks(runner.ID, 3, time.Now().Add(-5*time.Minute), app.RunnerStatusActive)
				// 2 checks beyond 30 minutes (should be filtered out with window=10m)
				s.createHealthChecks(runner.ID, 2, time.Now().Add(-2*time.Hour), app.RunnerStatusActive)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			queryParams:   "?window=10m",
			expectedCode:  http.StatusOK,
			expectedCount: 3,
			validateFunc: func(healthChecks []*app.RunnerHealthCheck) {
				assert.Len(s.T(), healthChecks, 3)
				// All health checks should be within 10 minutes
				for _, hc := range healthChecks {
					assert.True(s.T(), time.Since(hc.CreatedAt) <= 10*time.Minute)
				}
			},
		},
		{
			name: "default window is 1 hour",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "runner-default-window",
					DisplayName:   "Runner Default Window",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				// Create health checks at different times
				// 3 checks within 30 minutes
				s.createHealthChecks(runner.ID, 3, time.Now().Add(-30*time.Minute), app.RunnerStatusActive)
				// 2 checks beyond 2 hours (should be filtered out with default 1h window)
				s.createHealthChecks(runner.ID, 2, time.Now().Add(-3*time.Hour), app.RunnerStatusActive)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			expectedCode:  http.StatusOK,
			expectedCount: 3,
			validateFunc: func(healthChecks []*app.RunnerHealthCheck) {
				assert.Len(s.T(), healthChecks, 3)
			},
		},
		{
			name: "invalid window parameter returns error",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			queryParams:  "?window=invalid",
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "returns health checks with different statuses",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "runner-multi-status",
					DisplayName:   "Runner Multi Status",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				// Create health checks with different statuses
				s.createHealthChecks(runner.ID, 2, time.Now().Add(-30*time.Minute), app.RunnerStatusActive)
				s.createHealthChecks(runner.ID, 2, time.Now().Add(-20*time.Minute), app.RunnerStatusError)
				s.createHealthChecks(runner.ID, 1, time.Now().Add(-10*time.Minute), app.RunnerStatusOffline)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			expectedCode:  http.StatusOK,
			expectedCount: 5,
			validateFunc: func(healthChecks []*app.RunnerHealthCheck) {
				assert.Len(s.T(), healthChecks, 5)
				// Verify we got different statuses
				statusCounts := make(map[app.RunnerStatus]int)
				for _, hc := range healthChecks {
					statusCounts[hc.RunnerStatus]++
				}
				assert.Equal(s.T(), 2, statusCounts[app.RunnerStatusActive])
				assert.Equal(s.T(), 2, statusCounts[app.RunnerStatusError])
				assert.Equal(s.T(), 1, statusCounts[app.RunnerStatusOffline])
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID := tc.setupFunc()
			path := fmt.Sprintf("/v1/runners/%s/recent-health-checks%s", runnerID, tc.queryParams)
			rr := s.makeRequest("GET", path)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var healthChecks []*app.RunnerHealthCheck
				err := json.Unmarshal(rr.Body.Bytes(), &healthChecks)
				require.NoError(s.T(), err)

				if tc.expectedCount > 0 {
					assert.Len(s.T(), healthChecks, tc.expectedCount)
				}

				if tc.validateFunc != nil {
					tc.validateFunc(healthChecks)
				}
			} else {
				// Error cases should have error in response
				assert.Contains(s.T(), rr.Body.String(), "error")
			}
		})
	}
}
