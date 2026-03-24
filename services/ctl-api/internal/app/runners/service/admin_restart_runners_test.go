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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AdminRestartRunnersTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminRestartRunnersTestSuite struct {
	tests.BaseDBTestSuite
	app          *fxtest.App
	service      AdminRestartRunnersTestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	mockEvClient *tests.MockEventLoopClient
}

func TestAdminRestartRunnersSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminRestartRunnersTestSuite))
}

func (s *AdminRestartRunnersTestSuite) SetupSuite() {
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

func (s *AdminRestartRunnersTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.mockEvClient.Reset()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminRestartRunnersTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminRestartRunnersTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
	// Handler filters by OrgTypeDefault, so update from default OrgTypeSandbox
	err := s.service.DB.WithContext(ctx).Model(s.testOrg).Updates(map[string]interface{}{
		"org_type":     app.OrgTypeDefault,
		"sandbox_mode": false,
	}).Error
	require.NoError(s.T(), err)
}

func (s *AdminRestartRunnersTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminRestartRunnersTestSuite) TestAdminRestartRunners() {
	testCases := []struct {
		name         string
		setupFunc    func() []string
		requestBody  interface{}
		expectedCode int
		validateFunc func([]string)
	}{
		{
			name: "empty runners returns empty array",
			setupFunc: func() []string {
				return []string{}
			},
			requestBody:  AdminRestartRunnersRequest{},
			expectedCode: http.StatusOK,
			validateFunc: func(runnerIDs []string) {
				signals := s.mockEvClient.GetSignals()
				assert.Len(s.T(), signals, 0)
			},
		},
		{
			name: "with test runners returns responses",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runnerGrp := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   s.testOrg.ID,
					OwnerType: "org",
					Type:      app.RunnerGroupTypeOrg,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err := s.service.DB.WithContext(ctx).Create(runnerGrp).Error
				require.NoError(s.T(), err)

				runner1 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "test-runner-1",
					DisplayName:   "Test Runner 1",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: runnerGrp.ID,
				}
				err = s.service.DB.WithContext(ctx).Create(runner1).Error
				require.NoError(s.T(), err)

				runner2 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "test-runner-2",
					DisplayName:   "Test Runner 2",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: runnerGrp.ID,
				}
				err = s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner1)
					s.service.DB.Unscoped().Delete(runner2)
					s.service.DB.Unscoped().Delete(runnerGrp)
				})

				return []string{runner1.ID, runner2.ID}
			},
			requestBody:  AdminRestartRunnersRequest{},
			expectedCode: http.StatusOK,
			validateFunc: func(runnerIDs []string) {
				// Handler correctly finds runners for non-sandbox orgs and sends restart signals
				signals := s.mockEvClient.GetSignals()
				assert.Len(s.T(), signals, 2, "should send restart signal for each runner")
			},
		},
		{
			name: "empty body accepted",
			setupFunc: func() []string {
				return []string{}
			},
			requestBody:  nil,
			expectedCode: http.StatusOK,
			validateFunc: func(runnerIDs []string) {
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerIDs := tc.setupFunc()
			rr := s.makeRequest("POST", "/v1/runners/restart", tc.requestBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc(runnerIDs)
			}
		})
	}
}
