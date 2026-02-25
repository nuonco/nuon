package service

import (
	"bytes"
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
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type RunnerLifecycleSignalsTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type RunnerLifecycleSignalsTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       RunnerLifecycleSignalsTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	mockEvClient  *tests.MockEventLoopClient
}

func TestRunnerLifecycleSignalsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(RunnerLifecycleSignalsTestSuite))
}

func (s *RunnerLifecycleSignalsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	// Create and inject mock EventLoop client
	s.mockEvClient = tests.NewMockEventLoopClient()

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			Mocks: &tests.TestMocks{MockEv: s.mockEvClient},

			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *RunnerLifecycleSignalsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.mockEvClient.Reset()
	s.setupTestData()

	// Create router with public routes (lifecycle signal endpoints are public)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *RunnerLifecycleSignalsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *RunnerLifecycleSignalsTestSuite) setupTestData() {
	ctx := context.Background()

	// Use Seeder for account and org creation
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create runner group (domain-specific entity, built manually)
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

	// Create runner (domain-specific entity, built manually)
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

func (s *RunnerLifecycleSignalsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(s.T(), err)
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *RunnerLifecycleSignalsTestSuite) TestLifecycleSignals() {
	testCases := []struct {
		name           string
		path           string
		expectedSignal eventloop.SignalType
	}{
		{
			name:           "graceful shutdown",
			path:           "/v1/runners/" + s.testRunner.ID + "/graceful-shutdown",
			expectedSignal: signals.OperationGracefulShutdown,
		},
		{
			name:           "force shutdown",
			path:           "/v1/runners/" + s.testRunner.ID + "/force-shutdown",
			expectedSignal: signals.OperationForceShutdown,
		},
		{
			name:           "mng shutdown",
			path:           "/v1/runners/" + s.testRunner.ID + "/mng/shutdown",
			expectedSignal: signals.OperationMngShutDown,
		},
		{
			name:           "mng shutdown vm",
			path:           "/v1/runners/" + s.testRunner.ID + "/mng/shutdown-vm",
			expectedSignal: signals.OperationMngVMShutDown,
		},
		{
			name:           "mng update",
			path:           "/v1/runners/" + s.testRunner.ID + "/mng/update",
			expectedSignal: signals.OperationMngUpdate,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.mockEvClient.Reset()

			// Send empty JSON body
			rr := s.makeRequest("POST", tc.path, map[string]interface{}{})

			if rr.Code != http.StatusCreated {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusCreated, rr.Code)

			// Verify response is true
			var response bool
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(s.T(), err)
			assert.True(s.T(), response)

			// Verify signal was sent with correct type and runner ID
			sigs := s.mockEvClient.GetSignals()
			require.Len(s.T(), sigs, 1)
			assert.Equal(s.T(), s.testRunner.ID, sigs[0].ID)

			sig, ok := sigs[0].Signal.(*signals.Signal)
			require.True(s.T(), ok, "signal should be *signals.Signal type")
			assert.Equal(s.T(), tc.expectedSignal, sig.Type)
		})
	}
}

func (s *RunnerLifecycleSignalsTestSuite) TestLifecycleSignals_RunnerNotFound() {
	testCases := []struct {
		name string
		path string
	}{
		{
			name: "graceful shutdown - runner not found",
			path: "/v1/runners/rnrnonexistent123456789012/graceful-shutdown",
		},
		{
			name: "force shutdown - runner not found",
			path: "/v1/runners/rnrnonexistent123456789012/force-shutdown",
		},
		{
			name: "mng shutdown - runner not found",
			path: "/v1/runners/rnrnonexistent123456789012/mng/shutdown",
		},
		{
			name: "mng shutdown vm - runner not found",
			path: "/v1/runners/rnrnonexistent123456789012/mng/shutdown-vm",
		},
		{
			name: "mng update - runner not found",
			path: "/v1/runners/rnrnonexistent123456789012/mng/update",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.mockEvClient.Reset()

			rr := s.makeRequest("POST", tc.path, map[string]interface{}{})

			// Should return error (not 201)
			require.Equal(s.T(), http.StatusNotFound, rr.Code)
			assert.Contains(s.T(), rr.Body.String(), "unable to get runner")

			// Verify no signal was sent
			sigs := s.mockEvClient.GetSignals()
			assert.Len(s.T(), sigs, 0, "should not send signal when runner not found")
		})
	}
}

func (s *RunnerLifecycleSignalsTestSuite) TestLifecycleSignals_CrossOrgIsolation() {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	// Create second org
	org2 := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "other-org",
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/other",
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
		Name:          "runner-org2",
		DisplayName:   "Runner in Org 2",
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

	testCases := []struct {
		name string
		path string
	}{
		{
			name: "graceful shutdown - cross org",
			path: "/v1/runners/" + runner2.ID + "/graceful-shutdown",
		},
		{
			name: "force shutdown - cross org",
			path: "/v1/runners/" + runner2.ID + "/force-shutdown",
		},
		{
			name: "mng shutdown - cross org",
			path: "/v1/runners/" + runner2.ID + "/mng/shutdown",
		},
		{
			name: "mng shutdown vm - cross org",
			path: "/v1/runners/" + runner2.ID + "/mng/shutdown-vm",
		},
		{
			name: "mng update - cross org",
			path: "/v1/runners/" + runner2.ID + "/mng/update",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.mockEvClient.Reset()

			// Try to access runner from different org (router has s.testOrg context)
			rr := s.makeRequest("POST", tc.path, map[string]interface{}{})

			// Should return error due to cross-org isolation
			require.Equal(s.T(), http.StatusNotFound, rr.Code)
			assert.Contains(s.T(), rr.Body.String(), "unable to get runner")

			// Verify no signal was sent
			sigs := s.mockEvClient.GetSignals()
			assert.Len(s.T(), sigs, 0, "should not send signal for cross-org access")
		})
	}
}

func (s *RunnerLifecycleSignalsTestSuite) TestLifecycleSignals_CorrectRunnerID() {
	// Verify signals are sent to runner ID, not org ID or runner group ID
	s.mockEvClient.Reset()

	rr := s.makeRequest("POST", "/v1/runners/"+s.testRunner.ID+"/graceful-shutdown", map[string]interface{}{})
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	sigs := s.mockEvClient.GetSignals()
	require.Len(s.T(), sigs, 1)

	// Signal ID should be runner ID
	assert.Equal(s.T(), s.testRunner.ID, sigs[0].ID, "signal should be sent to runner ID")
	assert.NotEqual(s.T(), s.testOrg.ID, sigs[0].ID, "signal should not be sent to org ID")
	assert.NotEqual(s.T(), s.testRunnerGrp.ID, sigs[0].ID, "signal should not be sent to runner group ID")
}
