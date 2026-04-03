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

type GetRunnerLatestHeartBeatFromViewTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetRunnerLatestHeartBeatFromViewTestSuite struct {
	tests.BaseDBTestSuite
	app            *fxtest.App
	service        GetRunnerLatestHeartBeatFromViewTestService
	router         *gin.Engine
	testOrg        *app.Org
	testAcc        *app.Account
	testRunner     *app.Runner
	testRunnerGrp  *app.RunnerGroup
	testRunnerGrpS *app.RunnerGroupSettings
}

func TestGetRunnerLatestHeartBeatFromViewSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetRunnerLatestHeartBeatFromViewTestSuite))
}

func (s *GetRunnerLatestHeartBeatFromViewTestSuite) SetupSuite() {
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

func (s *GetRunnerLatestHeartBeatFromViewTestSuite) SetupTest() {
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

func (s *GetRunnerLatestHeartBeatFromViewTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetRunnerLatestHeartBeatFromViewTestSuite) setupTestData() {
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

func (s *GetRunnerLatestHeartBeatFromViewTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetRunnerLatestHeartBeatFromViewTestSuite) createRunnerHeartBeat(runnerID string, process app.RunnerProcessType, timestamp time.Time, aliveTime time.Duration, version string) string {
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

func (s *GetRunnerLatestHeartBeatFromViewTestSuite) TestGetRunnerLatestHeartBeatFromView() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		expectedCode int
		validateFunc func(LatestRunnerHeartBeats)
	}{
		{
			name: "returns latest heartbeats grouped by process",
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
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessTypeMng, baseTime, time.Minute*5, "1.0.0")
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessTypeMng, baseTime.Add(2*time.Minute), time.Minute*7, "1.0.1")
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessTypeInstall, baseTime.Add(1*time.Minute), time.Minute*6, "2.0.0")
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessTypeOrg, baseTime.Add(3*time.Minute), time.Minute*8, "3.0.0")

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(heartBeats LatestRunnerHeartBeats) {
				// Should have one entry per process
				assert.Len(s.T(), heartBeats, 3, "Should have 3 process types")

				// Verify each process has the latest heartbeat
				mngHB, ok := heartBeats[string(app.RunnerProcessTypeMng)]
				assert.True(s.T(), ok, "Should have mng process")
				if ok {
					assert.Equal(s.T(), "1.0.1", mngHB.Version, "Should have latest mng heartbeat")
					assert.Equal(s.T(), app.RunnerProcessTypeMng, mngHB.Process)
				}

				installHB, ok := heartBeats[string(app.RunnerProcessTypeInstall)]
				assert.True(s.T(), ok, "Should have install process")
				if ok {
					assert.Equal(s.T(), "2.0.0", installHB.Version)
					assert.Equal(s.T(), app.RunnerProcessTypeInstall, installHB.Process)
				}

				orgHB, ok := heartBeats[string(app.RunnerProcessTypeOrg)]
				assert.True(s.T(), ok, "Should have org process")
				if ok {
					assert.Equal(s.T(), "3.0.0", orgHB.Version)
					assert.Equal(s.T(), app.RunnerProcessTypeOrg, orgHB.Process)
				}
			},
		},
		{
			name: "deduplicates heartbeats by process",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					RunnerGroupID: s.testRunnerGrp.ID,
					Name:          "runner-dedupe",
					DisplayName:   "Runner Dedupe",
					Status:        app.RunnerStatusActive,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				// Create multiple heartbeats for same process
				baseTime := time.Now().Add(-10 * time.Minute)
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessTypeMng, baseTime, time.Minute*5, "1.0.0")
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessTypeMng, baseTime.Add(1*time.Minute), time.Minute*6, "1.0.1")
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessTypeMng, baseTime.Add(2*time.Minute), time.Minute*7, "1.0.2")
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessTypeMng, baseTime.Add(3*time.Minute), time.Minute*8, "1.0.3")

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(heartBeats LatestRunnerHeartBeats) {
				// Should dedupe to one entry per process
				assert.Len(s.T(), heartBeats, 1, "Should dedupe to single entry")

				mngHB, ok := heartBeats[string(app.RunnerProcessTypeMng)]
				assert.True(s.T(), ok, "Should have mng process")
				if ok {
					assert.Equal(s.T(), "1.0.3", mngHB.Version, "Should have latest heartbeat after deduplication")
				}
			},
		},
		{
			name: "returns empty map for runner with no heartbeats",
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
			expectedCode: http.StatusOK,
			validateFunc: func(heartBeats LatestRunnerHeartBeats) {
				assert.Empty(s.T(), heartBeats, "Should return empty map for runner with no heartbeats")
			},
		},
		{
			name: "runner not found returns 404",
			setupFunc: func() string {
				return "runnonexistent123456789012"
			},
			expectedCode: http.StatusNotFound,
			validateFunc: func(heartBeats LatestRunnerHeartBeats) {
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
			validateFunc: func(heartBeats LatestRunnerHeartBeats) {
				// Error response
			},
		},
		{
			name: "heartbeat StartedAt is calculated correctly in view results",
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
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessTypeMng, createdAt, time.Minute*15, "1.0.0")

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(heartBeats LatestRunnerHeartBeats) {
				assert.Len(s.T(), heartBeats, 1)
				mngHB, ok := heartBeats[string(app.RunnerProcessTypeMng)]
				assert.True(s.T(), ok)
				if ok {
					// StartedAt should be CreatedAt - AliveTime
					expectedStartedAt := mngHB.CreatedAt.Add(-1 * mngHB.AliveTime)
					assert.WithinDuration(s.T(), expectedStartedAt, mngHB.StartedAt, time.Second,
						"StartedAt should be calculated as CreatedAt - AliveTime")
				}
			},
		},
		{
			name: "handles all three process types",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					RunnerGroupID: s.testRunnerGrp.ID,
					Name:          "runner-all-processes",
					DisplayName:   "Runner All Processes",
					Status:        app.RunnerStatusActive,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				// Create heartbeats for all process types
				baseTime := time.Now()
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessTypeMng, baseTime, time.Minute*5, "mng-1.0.0")
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessTypeInstall, baseTime.Add(time.Minute), time.Minute*6, "install-2.0.0")
				s.createRunnerHeartBeat(runner.ID, app.RunnerProcessTypeOrg, baseTime.Add(2*time.Minute), time.Minute*7, "org-3.0.0")

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(heartBeats LatestRunnerHeartBeats) {
				assert.Len(s.T(), heartBeats, 3, "Should have all three process types")

				// Verify each process type
				processTypes := []app.RunnerProcessType{
					app.RunnerProcessTypeMng,
					app.RunnerProcessTypeInstall,
					app.RunnerProcessTypeOrg,
				}

				for _, process := range processTypes {
					hb, ok := heartBeats[string(process)]
					assert.True(s.T(), ok, "Should have %s process", process)
					if ok {
						assert.Equal(s.T(), process, hb.Process)
						assert.NotEmpty(s.T(), hb.Version)
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID := tc.setupFunc()
			path := fmt.Sprintf("/v1/runners/%s/heart-beats/latest", runnerID)
			rr := s.makeRequest("GET", path)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var heartBeats LatestRunnerHeartBeats
				err := json.Unmarshal(rr.Body.Bytes(), &heartBeats)
				require.NoError(s.T(), err)

				if tc.validateFunc != nil {
					tc.validateFunc(heartBeats)
				}
			} else {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}
		})
	}
}
