package service

import (
	"bytes"
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

type CreateHeartbeatTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type CreateHeartbeatTestSuite struct {
	tests.BaseDBTestSuite
	app             *fxtest.App
	service         CreateHeartbeatTestService
	router          *gin.Engine
	testOrg         *app.Org
	testAcc         *app.Account
	testRunner      *app.Runner
	testRunnerGroup *app.RunnerGroup
}

func TestCreateHeartbeatSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(CreateHeartbeatTestSuite))
}

func (s *CreateHeartbeatTestSuite) SetupSuite() {
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
	s.SetCHDB(s.service.CHDB)
}

func (s *CreateHeartbeatTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with runner routes
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterRunnerRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *CreateHeartbeatTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CreateHeartbeatTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create runner group
	s.testRunnerGroup = &app.RunnerGroup{
		ID:        domains.NewRunnerGroupID(),
		OrgID:     s.testOrg.ID,
		OwnerID:   s.testOrg.ID,
		OwnerType: "org",
		Type:      app.RunnerGroupTypeOrg,
		Platform:  app.AppRunnerTypeAWSEKS,
	}
	err := s.service.DB.WithContext(ctx).Create(s.testRunnerGroup).Error
	require.NoError(s.T(), err)

	// Create runner
	s.testRunner = &app.Runner{
		ID:            domains.NewRunnerID(),
		OrgID:         s.testOrg.ID,
		RunnerGroupID: s.testRunnerGroup.ID,
		Name:          "test-runner",
		DisplayName:   "Test Runner",
		Status:        app.RunnerStatusActive,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunner).Error
	require.NoError(s.T(), err)
}

func (s *CreateHeartbeatTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *CreateHeartbeatTestSuite) getHeartbeatsFromCH(runnerID string) []app.RunnerHeartBeat {
	ctx := context.Background()
	ctx = cctx.SetOrgContext(ctx, s.testOrg)

	var heartbeats []app.RunnerHeartBeat
	err := s.service.CHDB.WithContext(ctx).
		Where("runner_id = ?", runnerID).
		Order("created_at DESC").
		Find(&heartbeats).Error
	require.NoError(s.T(), err)

	return heartbeats
}

func (s *CreateHeartbeatTestSuite) TestCreateHeartbeat() {
	testCases := []struct {
		name         string
		setupFunc    func() (string, CreateRunnerHeartBeatRequest)
		expectedCode int
		validateFunc func(*app.RunnerHeartBeat)
	}{
		{
			name: "successfully create heartbeat with all fields",
			setupFunc: func() (string, CreateRunnerHeartBeatRequest) {
				return s.testRunner.ID, CreateRunnerHeartBeatRequest{
					AliveTime: 5 * time.Minute,
					Version:   "v1.2.3",
					Process:   app.RunnerProcessInstall,
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(hb *app.RunnerHeartBeat) {
				assert.Equal(s.T(), s.testRunner.ID, hb.RunnerID)
				assert.Equal(s.T(), 5*time.Minute, hb.AliveTime)
				assert.Equal(s.T(), "v1.2.3", hb.Version)
				assert.Equal(s.T(), app.RunnerProcessInstall, hb.Process)
				assert.NotEmpty(s.T(), hb.ID)
				assert.False(s.T(), hb.CreatedAt.IsZero())

				// Verify stored in ClickHouse
				chHeartbeats := s.getHeartbeatsFromCH(s.testRunner.ID)
				require.NotEmpty(s.T(), chHeartbeats)
				assert.Equal(s.T(), s.testRunner.ID, chHeartbeats[0].RunnerID)
				assert.Equal(s.T(), 5*time.Minute, chHeartbeats[0].AliveTime)
			},
		},
		{
			name: "successfully create heartbeat with mng process",
			setupFunc: func() (string, CreateRunnerHeartBeatRequest) {
				return s.testRunner.ID, CreateRunnerHeartBeatRequest{
					AliveTime: 10 * time.Minute,
					Version:   "v2.0.0",
					Process:   app.RunnerProcessMng,
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(hb *app.RunnerHeartBeat) {
				assert.Equal(s.T(), app.RunnerProcessMng, hb.Process)
				assert.Equal(s.T(), 10*time.Minute, hb.AliveTime)
			},
		},
		{
			name: "successfully create heartbeat with org process",
			setupFunc: func() (string, CreateRunnerHeartBeatRequest) {
				return s.testRunner.ID, CreateRunnerHeartBeatRequest{
					AliveTime: 3 * time.Minute,
					Version:   "v1.0.0",
					Process:   app.RunnerProcessOrg,
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(hb *app.RunnerHeartBeat) {
				assert.Equal(s.T(), app.RunnerProcessOrg, hb.Process)
			},
		},
		{
			name: "defaults to unknown process when process not provided",
			setupFunc: func() (string, CreateRunnerHeartBeatRequest) {
				return s.testRunner.ID, CreateRunnerHeartBeatRequest{
					AliveTime: 2 * time.Minute,
					Version:   "v1.5.0",
					// Process field omitted
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(hb *app.RunnerHeartBeat) {
				assert.Equal(s.T(), app.RunnerProcessUknown, hb.Process)
				assert.Equal(s.T(), 2*time.Minute, hb.AliveTime)
			},
		},
		{
			name: "version is optional",
			setupFunc: func() (string, CreateRunnerHeartBeatRequest) {
				return s.testRunner.ID, CreateRunnerHeartBeatRequest{
					AliveTime: 1 * time.Minute,
					Process:   app.RunnerProcessInstall,
					// Version field omitted
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(hb *app.RunnerHeartBeat) {
				assert.Equal(s.T(), "", hb.Version)
				assert.Equal(s.T(), app.RunnerProcessInstall, hb.Process)
			},
		},
		{
			name: "stores multiple heartbeats for same runner",
			setupFunc: func() (string, CreateRunnerHeartBeatRequest) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create first heartbeat via database
				hb1 := &app.RunnerHeartBeat{
					RunnerID:  s.testRunner.ID,
					AliveTime: 1 * time.Minute,
					Version:   "v1.0.0",
					Process:   app.RunnerProcessInstall,
				}
				err := s.service.CHDB.WithContext(ctx).Create(hb1).Error
				require.NoError(s.T(), err)

				// Allow ClickHouse eventual consistency
				time.Sleep(200 * time.Millisecond)

				// Now create second heartbeat via API
				return s.testRunner.ID, CreateRunnerHeartBeatRequest{
					AliveTime: 2 * time.Minute,
					Version:   "v1.0.0",
					Process:   app.RunnerProcessInstall,
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(hb *app.RunnerHeartBeat) {
				// Allow ClickHouse eventual consistency for the API-created heartbeat
				time.Sleep(200 * time.Millisecond)

				// Verify both heartbeats exist in ClickHouse
				chHeartbeats := s.getHeartbeatsFromCH(s.testRunner.ID)
				assert.GreaterOrEqual(s.T(), len(chHeartbeats), 2)
			},
		},
		{
			name: "runner in different org creates heartbeat successfully",
			setupFunc: func() (string, CreateRunnerHeartBeatRequest) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create second org
				org2 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "other-org",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/bar",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)

				// Create runner group in org2
				rg2 := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OrgID:     org2.ID,
					OwnerID:   org2.ID,
					OwnerType: "org",
					Type:      app.RunnerGroupTypeOrg,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err = s.service.DB.WithContext(ctx).Create(rg2).Error
				require.NoError(s.T(), err)

				// Create runner in org2
				runner2 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         org2.ID,
					RunnerGroupID: rg2.ID,
					Name:          "runner-org2",
					DisplayName:   "Runner Org2",
					Status:        app.RunnerStatusActive,
				}
				err = s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner2)
					s.service.DB.Unscoped().Delete(rg2)
					s.service.DB.Unscoped().Delete(org2)
				})

				return runner2.ID, CreateRunnerHeartBeatRequest{
					AliveTime: 5 * time.Minute,
					Version:   "v1.2.3",
					Process:   app.RunnerProcessInstall,
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(hb *app.RunnerHeartBeat) {
				assert.NotEmpty(s.T(), hb.RunnerID)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID, req := tc.setupFunc()
			path := fmt.Sprintf("/v1/runners/%s/heart-beats", runnerID)
			rr := s.makeRequest("POST", path, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusCreated {
				var heartbeat app.RunnerHeartBeat
				err := json.Unmarshal(rr.Body.Bytes(), &heartbeat)
				require.NoError(s.T(), err)

				if tc.validateFunc != nil {
					tc.validateFunc(&heartbeat)
				}
			}
		})
	}
}

func (s *CreateHeartbeatTestSuite) TestCreateHeartbeatValidation() {
	testCases := []struct {
		name         string
		request      interface{}
		expectedCode int
		errorSubstr  string
	}{
		{
			name: "missing alive_time still succeeds",
			request: map[string]interface{}{
				"version": "v1.0.0",
				"process": string(app.RunnerProcessInstall),
			},
			// The struct uses validate:"required" not binding:"required",
			// so Gin's ShouldBindJSON does not enforce validation.
			// AliveTime defaults to 0 and the heartbeat is created successfully.
			expectedCode: http.StatusCreated,
		},
		{
			name: "alive_time of zero still succeeds",
			request: CreateRunnerHeartBeatRequest{
				AliveTime: 0,
				Version:   "v1.0.0",
				Process:   app.RunnerProcessInstall,
			},
			// Same as above: validate:"required" is not enforced by ShouldBindJSON.
			expectedCode: http.StatusCreated,
		},
		{
			name: "invalid JSON returns error",
			request: map[string]interface{}{
				"alive_time": "not-a-duration",
			},
			expectedCode: http.StatusBadRequest,
			errorSubstr:  "error",
		},
		{
			name: "empty request body still succeeds",
			// All fields default to zero values; ShouldBindJSON succeeds.
			request:      map[string]interface{}{},
			expectedCode: http.StatusCreated,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := fmt.Sprintf("/v1/runners/%s/heart-beats", s.testRunner.ID)
			rr := s.makeRequest("POST", path, tc.request)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.errorSubstr != "" {
				bodyStr := rr.Body.String()
				assert.Contains(s.T(), bodyStr, tc.errorSubstr)
			}
		})
	}
}

func (s *CreateHeartbeatTestSuite) TestCreateHeartbeatRunnerNotFound() {
	s.Run("nonexistent runner ID returns error", func() {
		// The heartbeat is created in ClickHouse first, but then the handler
		// tries to look up the runner in PostgreSQL for metrics tagging.
		// When the runner doesn't exist, the lookup fails with gorm.ErrRecordNotFound
		// which the stderr middleware maps to 404.
		nonexistentRunnerID := "runnon" + "existent123456"
		req := CreateRunnerHeartBeatRequest{
			AliveTime: 5 * time.Minute,
			Version:   "v1.0.0",
			Process:   app.RunnerProcessInstall,
		}

		path := fmt.Sprintf("/v1/runners/%s/heart-beats", nonexistentRunnerID)
		rr := s.makeRequest("POST", path, req)

		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())

		// The heartbeat is written to ClickHouse, but heartbeatGetRunner fails
		// because the runner doesn't exist in PostgreSQL. The wrapped error contains
		// gorm.ErrRecordNotFound, which the stderr middleware detects via errors.Is
		// and returns 404.
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}

func (s *CreateHeartbeatTestSuite) TestCreateHeartbeatDifferentProcesses() {
	processes := []struct {
		process     app.RunnerProcess
		description string
	}{
		{app.RunnerProcessMng, "mng process"},
		{app.RunnerProcessInstall, "install process"},
		{app.RunnerProcessOrg, "org process"},
	}

	for _, tc := range processes {
		s.Run(tc.description, func() {
			req := CreateRunnerHeartBeatRequest{
				AliveTime: 5 * time.Minute,
				Version:   "v1.0.0",
				Process:   tc.process,
			}

			path := fmt.Sprintf("/v1/runners/%s/heart-beats", s.testRunner.ID)
			rr := s.makeRequest("POST", path, req)

			if rr.Code != http.StatusCreated {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusCreated, rr.Code)

			var heartbeat app.RunnerHeartBeat
			err := json.Unmarshal(rr.Body.Bytes(), &heartbeat)
			require.NoError(s.T(), err)
			assert.Equal(s.T(), tc.process, heartbeat.Process)

			// Verify in ClickHouse
			chHeartbeats := s.getHeartbeatsFromCH(s.testRunner.ID)
			found := false
			for _, chHB := range chHeartbeats {
				if chHB.Process == tc.process {
					found = true
					break
				}
			}
			assert.True(s.T(), found, "Heartbeat with process %s should exist in ClickHouse", tc.process)
		})
	}
}
