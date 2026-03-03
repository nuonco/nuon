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

type GetRunnerLatestHeartBeatTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetRunnerLatestHeartBeatTestSuite struct {
	tests.BaseDBTestSuite
	app            *fxtest.App
	service        GetRunnerLatestHeartBeatTestService
	router         *gin.Engine
	testOrg        *app.Org
	testAcc        *app.Account
	testRunner     *app.Runner
	testRunnerGrp  *app.RunnerGroup
	testRunnerGrpS *app.RunnerGroupSettings
}

func TestGetRunnerLatestHeartBeatSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetRunnerLatestHeartBeatTestSuite))
}

func (s *GetRunnerLatestHeartBeatTestSuite) SetupSuite() {
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

func (s *GetRunnerLatestHeartBeatTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetRunnerLatestHeartBeatTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetRunnerLatestHeartBeatTestSuite) setupTestData() {
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

	// Create runner group settings
	s.testRunnerGrpS = &app.RunnerGroupSettings{
		ID:            domains.NewRunnerGroupSettingsID(),
		OrgID:         s.testOrg.ID,
		RunnerGroupID: s.testRunnerGrp.ID,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunnerGrpS).Error
	require.NoError(s.T(), err)

	// Create runner in Postgres
	s.testRunner = &app.Runner{
		ID:            domains.NewRunnerID(),
		OrgID:         s.testOrg.ID,
		RunnerGroupID: s.testRunnerGrp.ID,
		Name:          "test-runner",
		DisplayName:   "Test Runner",
		Status:        app.RunnerStatusActive,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunner).Error
	require.NoError(s.T(), err)
}

func (s *GetRunnerLatestHeartBeatTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetRunnerLatestHeartBeatTestSuite) createRunnerHeartBeat(runnerID string, process app.RunnerProcess, timestamp time.Time, aliveTime time.Duration, version string) string {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	ctx = cctx.SetOrgContext(ctx, s.testOrg)

	heartBeat := &app.RunnerHeartBeat{
		ID:        domains.NewRunnerHeartBeatID(),
		RunnerID:  runnerID,
		Process:   process,
		AliveTime: aliveTime,
		Version:   version,
		CreatedAt: timestamp,
	}
	err := s.service.CHDB.WithContext(ctx).Create(heartBeat).Error
	require.NoError(s.T(), err)
	return heartBeat.ID
}

func (s *GetRunnerLatestHeartBeatTestSuite) TestGetRunnerLatestHeartBeat() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		expectedCode int
		validateFunc func(*app.RunnerHeartBeat)
	}{
		{
			name: "returns latest heartbeat for runner",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					RunnerGroupID: s.testRunnerGrp.ID,
					Name:          "runner-with-heartbeats",
					DisplayName:   "Runner With Heartbeats",
					Status:        app.RunnerStatusActive,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				// Create multiple heartbeats with different timestamps
				baseTime := time.Now().Add(-10 * time.Minute)
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessMng, baseTime, time.Minute*5, "1.0.0")
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessMng, baseTime.Add(2*time.Minute), time.Minute*7, "1.0.1")
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessMng, baseTime.Add(5*time.Minute), time.Minute*10, "1.0.2")

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(heartBeat *app.RunnerHeartBeat) {
				assert.NotNil(s.T(), heartBeat)
				assert.Equal(s.T(), "1.0.2", heartBeat.Version, "Should return the latest heartbeat")
				assert.Equal(s.T(), time.Minute*10, heartBeat.AliveTime)
				assert.Equal(s.T(), app.RunnerProcessMng, heartBeat.Process)
			},
		},
		{
			name: "returns latest heartbeat across multiple processes",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					RunnerGroupID: s.testRunnerGrp.ID,
					Name:          "runner-multi-process",
					DisplayName:   "Runner Multi Process",
					Status:        app.RunnerStatusActive,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				// Create heartbeats from different processes
				baseTime := time.Now().Add(-10 * time.Minute)
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessMng, baseTime, time.Minute*5, "1.0.0")
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessInstall, baseTime.Add(2*time.Minute), time.Minute*7, "1.1.0")
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessOrg, baseTime.Add(5*time.Minute), time.Minute*10, "1.2.0")

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(heartBeat *app.RunnerHeartBeat) {
				assert.NotNil(s.T(), heartBeat)
				assert.Equal(s.T(), "1.2.0", heartBeat.Version, "Should return the latest heartbeat across all processes")
				assert.Equal(s.T(), app.RunnerProcessOrg, heartBeat.Process)
			},
		},
		{
			name: "runner not found returns 404",
			setupFunc: func() string {
				return "runnonexistent123456789012"
			},
			expectedCode: http.StatusNotFound,
			validateFunc: func(heartBeat *app.RunnerHeartBeat) {
				// Error response
			},
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

				// Create runner group for org2
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
					RunnerGroupID: runnerGrp2.ID,
					Name:          "runner-in-org2",
					DisplayName:   "Runner In Org2",
					Status:        app.RunnerStatusActive,
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
			validateFunc: func(heartBeat *app.RunnerHeartBeat) {
				// Error response
			},
		},
		{
			name: "runner with no heartbeats returns error",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					RunnerGroupID: s.testRunnerGrp.ID,
					Name:          "runner-no-heartbeats",
					DisplayName:   "Runner No Heartbeats",
					Status:        app.RunnerStatusActive,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			expectedCode: http.StatusNotFound,
			validateFunc: func(heartBeat *app.RunnerHeartBeat) {
				// Error response
			},
		},
		{
			name: "heartbeat StartedAt is calculated correctly",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					RunnerGroupID: s.testRunnerGrp.ID,
					Name:          "runner-started-at",
					DisplayName:   "Runner Started At",
					Status:        app.RunnerStatusActive,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				// Create heartbeat with known AliveTime
				createdAt := time.Now()
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessMng, createdAt, time.Minute*15, "1.0.0")

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(heartBeat *app.RunnerHeartBeat) {
				assert.NotNil(s.T(), heartBeat)
				// StartedAt should be CreatedAt - AliveTime
				expectedStartedAt := heartBeat.CreatedAt.Add(-1 * heartBeat.AliveTime)
				assert.WithinDuration(s.T(), expectedStartedAt, heartBeat.StartedAt, time.Second,
					"StartedAt should be calculated as CreatedAt - AliveTime")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID := tc.setupFunc()
			path := fmt.Sprintf("/v1/runners/%s/latest-heart-beat", runnerID)
			rr := s.makeRequest("GET", path)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var heartBeat app.RunnerHeartBeat
				err := json.Unmarshal(rr.Body.Bytes(), &heartBeat)
				require.NoError(s.T(), err)

				if tc.validateFunc != nil {
					tc.validateFunc(&heartBeat)
				}
			} else {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}
		})
	}
}
