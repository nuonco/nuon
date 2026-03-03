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
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AdminRestartTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminRestartTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       AdminRestartTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	mockEvClient  *tests.MockEventLoopClient
}

func TestAdminRestartSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminRestartTestSuite))
}

func (s *AdminRestartTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

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

func (s *AdminRestartTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.mockEvClient.Reset()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminRestartTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminRestartTestSuite) setupTestData() {
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

func (s *AdminRestartTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminRestartTestSuite) TestRestartRunner() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		requestBody      interface{}
		expectedCode     int
		expectedSignal   bool
		validateFunc     func(string)
		expectedNotFound bool
	}{
		{
			name: "successfully restart runner",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			requestBody:    RestartRunnerRequest{},
			expectedCode:   http.StatusOK,
			expectedSignal: true,
			validateFunc: func(runnerID string) {
				capturedSignals := s.mockEvClient.GetSignals()
				require.Len(s.T(), capturedSignals, 1)
				assert.Equal(s.T(), runnerID, capturedSignals[0].ID)

				sig, ok := capturedSignals[0].Signal.(*signals.Signal)
				require.True(s.T(), ok)
				assert.Equal(s.T(), signals.OperationRestart, sig.Type)
			},
		},
		{
			name: "restart runner with empty body",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			requestBody:    nil,
			expectedCode:   http.StatusOK,
			expectedSignal: true,
			validateFunc: func(runnerID string) {
				signals := s.mockEvClient.GetSignals()
				require.Len(s.T(), signals, 1)
				assert.Equal(s.T(), runnerID, signals[0].ID)
			},
		},
		{
			name: "restart runner in different org",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org2ID := domains.NewOrgID()
				org2 := &app.Org{
					ID:          org2ID,
					Name:        fmt.Sprintf("other-org-%s", org2ID),
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/other",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)

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

				return runner2.ID
			},
			requestBody:    RestartRunnerRequest{},
			expectedCode:   http.StatusOK,
			expectedSignal: true,
		},
		{
			name: "nonexistent runner returns error",
			setupFunc: func() string {
				return "rnrnonexistent123456789012"
			},
			requestBody:      RestartRunnerRequest{},
			expectedCode:     http.StatusNotFound,
			expectedSignal:   false,
			expectedNotFound: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.mockEvClient.Reset()
			runnerID := tc.setupFunc()
			rr := s.makeRequest("POST", "/v1/runners/"+runnerID+"/restart", tc.requestBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}

			if tc.validateFunc != nil && rr.Code == http.StatusOK {
				tc.validateFunc(runnerID)
			}

			// Verify signal presence matches expectation
			capturedSignals := s.mockEvClient.GetSignals()
			if tc.expectedSignal {
				assert.Len(s.T(), capturedSignals, 1, "expected signal to be sent")
			} else {
				assert.Len(s.T(), capturedSignals, 0, "expected no signal to be sent")
			}
		})
	}
}
