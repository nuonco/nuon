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

type MngFetchTokenTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type MngFetchTokenTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       MngFetchTokenTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	mockEvClient  *tests.MockEventLoopClient
}

func TestMngFetchTokenSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(MngFetchTokenTestSuite))
}

func (s *MngFetchTokenTestSuite) SetupSuite() {
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

func (s *MngFetchTokenTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.mockEvClient.Reset()
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

func (s *MngFetchTokenTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *MngFetchTokenTestSuite) setupTestData() {
	ctx := context.Background()

	// Use Seeder for account and org creation
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

func (s *MngFetchTokenTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *MngFetchTokenTestSuite) TestMngFetchToken() {
	testCases := []struct {
		name           string
		setupFunc      func() string
		expectedCode   int
		shouldSendSig  bool
		expectedNotFnd bool
	}{
		{
			name: "successfully send fetch token signal",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			expectedCode:  http.StatusCreated,
			shouldSendSig: true,
		},
		{
			name: "runner not found",
			setupFunc: func() string {
				return "rnrnonexistent123456789012"
			},
			expectedCode:   http.StatusNotFound,
			shouldSendSig:  false,
			expectedNotFnd: true,
		},
		{
			name: "runner in different org not accessible",
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

				return runner2.ID
			},
			expectedCode:   http.StatusNotFound,
			shouldSendSig:  false,
			expectedNotFnd: true,
		},
		{
			name: "empty JSON body accepted",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			expectedCode:  http.StatusCreated,
			shouldSendSig: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.mockEvClient.Reset()
			runnerID := tc.setupFunc()

			rr := s.makeRequest("POST", "/v1/runners/"+runnerID+"/mng/fetch-token", map[string]interface{}{})

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFnd {
				assert.Contains(s.T(), rr.Body.String(), "unable to get runner")
			} else {
				// Verify response is true
				var response bool
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)
				assert.True(s.T(), response)
			}

			// Verify signal was sent or not
			sigs := s.mockEvClient.GetSignals()
			if tc.shouldSendSig {
				require.Len(s.T(), sigs, 1, "should send signal when runner found")
				assert.Equal(s.T(), runnerID, sigs[0].ID)

				sig, ok := sigs[0].Signal.(*signals.Signal)
				require.True(s.T(), ok, "signal should be *signals.Signal type")
				assert.Equal(s.T(), signals.OperationMngFetchToken, sig.Type)
			} else {
				assert.Len(s.T(), sigs, 0, "should not send signal when runner not found")
			}
		})
	}
}

func (s *MngFetchTokenTestSuite) TestMngFetchTokenCorrectRunnerID() {
	// Verify signal is sent to runner ID, not org ID or runner group ID
	s.mockEvClient.Reset()

	rr := s.makeRequest("POST", "/v1/runners/"+s.testRunner.ID+"/mng/fetch-token", map[string]interface{}{})
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	sigs := s.mockEvClient.GetSignals()
	require.Len(s.T(), sigs, 1)

	// Signal ID should be runner ID
	assert.Equal(s.T(), s.testRunner.ID, sigs[0].ID, "signal should be sent to runner ID")
	assert.NotEqual(s.T(), s.testOrg.ID, sigs[0].ID, "signal should not be sent to org ID")
	assert.NotEqual(s.T(), s.testRunnerGrp.ID, sigs[0].ID, "signal should not be sent to runner group ID")
}
