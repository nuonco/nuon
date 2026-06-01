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
	forgotten "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/forgotten"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AdminForgetInstallTestService struct {
	fx.In
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	Seeder          *testseed.Seeder
	InstallsService *service
}

type AdminForgetInstallTestSuite struct {
	tests.BaseDBTestSuite
	app         *fxtest.App
	service     AdminForgetInstallTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	testApp     *app.App
	testInstall *app.Install
}

func TestAdminForgetInstallSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminForgetInstallTestSuite))
}

func (s *AdminForgetInstallTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T:               s.T(),
			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AdminForgetInstallTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Admin routes do NOT use TestOrg/TestAcc context
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.InstallsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminForgetInstallTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminForgetInstallTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	ctx, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(ctx, s.T())
	s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)
	s.testInstall = s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
}

func (s *AdminForgetInstallTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminForgetInstallTestSuite) TestAdminForgetInstall() {
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
			name: "successfully forget install",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
				ctx = cctx.SetAccountIDContext(ctx, s.testAcc.ID)
				install := s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
				return install.ID
			},
			requestBody:    AdminForgetInstallRequest{},
			expectedCode:   http.StatusOK,
			expectedSignal: true,
			validateFunc: func(installID string) {
				// Verify install was deleted from DB
				var install app.Install
				err := s.service.DB.Where("id = ?", installID).First(&install).Error
				assert.Error(s.T(), err, "install should be deleted")
				assert.Equal(s.T(), gorm.ErrRecordNotFound, err)

				// Verify signal was sent
				sigs := tests.GetQueueSignals(s.T(), s.service.DB)
				require.Len(s.T(), sigs, 1)
				assert.Equal(s.T(), forgotten.SignalType, sigs[0].Type)
			},
		},
		{
			name: "forget with empty body",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
				ctx = cctx.SetAccountIDContext(ctx, s.testAcc.ID)
				install := s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
				return install.ID
			},
			requestBody:    nil,
			expectedCode:   http.StatusOK,
			expectedSignal: true,
			validateFunc: func(installID string) {
				var install app.Install
				err := s.service.DB.Where("id = ?", installID).First(&install).Error
				assert.Error(s.T(), err, "install should be deleted")
			},
		},
		{
			name: "nonexistent install returns error",
			setupFunc: func() string {
				return "ins000000000000000000000000"
			},
			requestBody:      AdminForgetInstallRequest{},
			expectedCode:     http.StatusNotFound,
			expectedSignal:   false,
			expectedNotFound: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			installID := tc.setupFunc()
			rr := s.makeRequest("POST", "/v1/installs/"+installID+"/admin-forget", tc.requestBody)

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
			capturedSignals := tests.GetQueueSignals(s.T(), s.service.DB)
			if tc.expectedSignal {
				assert.GreaterOrEqual(s.T(), len(capturedSignals), 1, "expected signal to be sent")
			} else {
				assert.Len(s.T(), capturedSignals, 0, "expected no signal to be sent")
			}
		})
	}
}
