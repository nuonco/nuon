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

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AdminForgetOrgInstallsTestService struct {
	fx.In
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	Seeder          *testseed.Seeder
	InstallsService *service
}

type AdminForgetOrgInstallsTestSuite struct {
	tests.BaseDBTestSuite
	app          *fxtest.App
	service      AdminForgetOrgInstallsTestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	testApp      *app.App
	testInstall  *app.Install
	mockEvClient *tests.MockEventLoopClient
}

func TestAdminForgetOrgInstallsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminForgetOrgInstallsTestSuite))
}

func (s *AdminForgetOrgInstallsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	s.mockEvClient = tests.NewMockEventLoopClient()
	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T:               s.T(),
			Mocks:           &tests.TestMocks{MockEv: s.mockEvClient},
			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AdminForgetOrgInstallsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.mockEvClient.Reset()
	s.setupTestData()

	// Admin routes do NOT use TestOrg/TestAcc context
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.InstallsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminForgetOrgInstallsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminForgetOrgInstallsTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	ctx, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(ctx, s.T())
	s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)
	s.testInstall = s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
}

func (s *AdminForgetOrgInstallsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminForgetOrgInstallsTestSuite) TestForgetOrgInstalls() {
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
			name: "successfully forget all installs for org",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
				ctx = cctx.SetAccountIDContext(ctx, s.testAcc.ID)

				// Create 3 installs for the test org
				var installIDs []string
				for i := 0; i < 3; i++ {
					install := s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
					installIDs = append(installIDs, install.ID)

					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Delete(&app.Install{ID: install.ID})
					})
				}

				return s.testOrg.ID
			},
			requestBody:    AdminForgetOrgInstallsRequest{},
			expectedCode:   http.StatusOK,
			expectedSignal: true,
			validateFunc: func(orgID string) {
				// Verify signals were sent
				sigs := s.mockEvClient.GetSignals()
				assert.GreaterOrEqual(s.T(), len(sigs), 3, "expected at least 3 signals for 3 installs")

				for _, sig := range sigs {
					evSig, ok := sig.Signal.(*signals.Signal)
					require.True(s.T(), ok)
					assert.Equal(s.T(), signals.OperationForget, evSig.Type)
				}
			},
		},
		{
			name: "forget with empty body",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
				ctx = cctx.SetAccountIDContext(ctx, s.testAcc.ID)
				install := s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(install)
				})

				return s.testOrg.ID
			},
			requestBody:    nil,
			expectedCode:   http.StatusOK,
			expectedSignal: true,
		},
		{
			name: "org with no installs returns success",
			setupFunc: func() string {
				// Previous test cases already forgot all installs for testOrg
				return s.testOrg.ID
			},
			requestBody:    AdminForgetOrgInstallsRequest{},
			expectedCode:   http.StatusOK,
			expectedSignal: false,
		},
		{
			name: "nonexistent org returns empty response",
			setupFunc: func() string {
				return "org000000000000000000000000"
			},
			requestBody:  AdminForgetOrgInstallsRequest{},
			expectedCode: http.StatusOK,
			// No installs means no signals, but not an error
			expectedSignal: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.mockEvClient.Reset()
			orgID := tc.setupFunc()
			rr := s.makeRequest("POST", "/v1/orgs/"+orgID+"/admin-forget-installs", tc.requestBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}

			if tc.validateFunc != nil && rr.Code == http.StatusOK {
				tc.validateFunc(orgID)
			}

			// Verify signal presence matches expectation
			capturedSignals := s.mockEvClient.GetSignals()
			if tc.expectedSignal {
				assert.GreaterOrEqual(s.T(), len(capturedSignals), 1, "expected at least one signal to be sent")
			} else {
				assert.Len(s.T(), capturedSignals, 0, "expected no signals to be sent")
			}
		})
	}
}
