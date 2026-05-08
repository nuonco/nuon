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
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AdminGenerateStateTestService struct {
	fx.In
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	Seeder          *testseed.Seeder
	InstallsService *service
}

type AdminGenerateStateTestSuite struct {
	tests.BaseDBTestSuite
	app          *fxtest.App
	service      AdminGenerateStateTestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	testApp      *app.App
	testInstall  *app.Install
	mockEvClient *tests.MockEventLoopClient
}

func TestAdminGenerateStateSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminGenerateStateTestSuite))
}

func (s *AdminGenerateStateTestSuite) SetupSuite() {
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

func (s *AdminGenerateStateTestSuite) SetupTest() {
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

func (s *AdminGenerateStateTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminGenerateStateTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	ctx, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(ctx, s.T())
	s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)
	s.testInstall = s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
}

func (s *AdminGenerateStateTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminGenerateStateTestSuite) TestAdminInstallGenerateInstallState() {
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
			name: "successfully generate state for install",
			setupFunc: func() string {
				return s.testInstall.ID
			},
			expectedCode:   http.StatusOK,
			expectedSignal: true,
			validateFunc: func(installID string) {
				capturedSignals := s.mockEvClient.GetSignals()
				require.Len(s.T(), capturedSignals, 1)
				assert.Equal(s.T(), installID, capturedSignals[0].ID)

				sig, ok := capturedSignals[0].Signal.(*signals.Signal)
				require.True(s.T(), ok)
				assert.Equal(s.T(), signals.OperationGenerateState, sig.Type)
			},
		},
		{
			name: "generate state with nil body",
			setupFunc: func() string {
				return s.testInstall.ID
			},
			requestBody:    nil,
			expectedCode:   http.StatusOK,
			expectedSignal: true,
			validateFunc: func(installID string) {
				sigs := s.mockEvClient.GetSignals()
				require.Len(s.T(), sigs, 1)
				assert.Equal(s.T(), installID, sigs[0].ID)
			},
		},
		{
			name: "generate state by install name",
			setupFunc: func() string {
				return s.testInstall.Name
			},
			expectedCode:   http.StatusOK,
			expectedSignal: true,
			validateFunc: func(installName string) {
				sigs := s.mockEvClient.GetSignals()
				require.Len(s.T(), sigs, 1)
				assert.Equal(s.T(), s.testInstall.ID, sigs[0].ID)
			},
		},
		{
			name: "nonexistent install returns error",
			setupFunc: func() string {
				return "ins000000000000000000000000"
			},
			expectedCode:     http.StatusNotFound,
			expectedSignal:   false,
			expectedNotFound: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.mockEvClient.Reset()
			installID := tc.setupFunc()
			rr := s.makeRequest("POST", "/v1/installs/"+installID+"/admin-generate-state", tc.requestBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}

			if tc.validateFunc != nil && rr.Code == http.StatusOK {
				tc.validateFunc(installID)
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
