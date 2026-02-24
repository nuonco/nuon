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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AdminUpdateInstallRunnerTestService struct {
	fx.In
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	Seeder          *testseed.Seeder
	InstallsService *service
}

type AdminUpdateInstallRunnerTestSuite struct {
	tests.BaseDBTestSuite
	app                   *fxtest.App
	service               AdminUpdateInstallRunnerTestService
	router                *gin.Engine
	testOrg               *app.Org
	testAcc               *app.Account
	testApp               *app.App
	testInstall           *app.Install
	testRunnerGrp         *app.RunnerGroup
	testRunnerGrpSettings *app.RunnerGroupSettings
	mockEvClient          *tests.FakeEventLoopClient
}

func TestAdminUpdateInstallRunnerSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminUpdateInstallRunnerTestSuite))
}

func (s *AdminUpdateInstallRunnerTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	s.mockEvClient = tests.NewFakeEventLoopClient()
	options := append(
		tests.CtlApiFXOptions(),
		fx.Decorate(func() eventloop.Client {
			return s.mockEvClient
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AdminUpdateInstallRunnerTestSuite) SetupTest() {
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

func (s *AdminUpdateInstallRunnerTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminUpdateInstallRunnerTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	ctx, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(ctx, s.T())
	s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)
	s.testInstall = s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)

	// Create runner group linked to install via polymorphic association
	s.testRunnerGrp = &app.RunnerGroup{
		OrgID:     s.testOrg.ID,
		OwnerID:   s.testInstall.ID,
		OwnerType: "installs",
		Type:      app.RunnerGroupTypeInstall,
		Platform:  app.AppRunnerTypeAWSEKS,
	}
	err := s.service.DB.WithContext(ctx).Create(s.testRunnerGrp).Error
	require.NoError(s.T(), err)

	// Create runner group settings
	s.testRunnerGrpSettings = &app.RunnerGroupSettings{
		RunnerGroupID:     s.testRunnerGrp.ID,
		ContainerImageTag: "v1.0.0",
		Metadata:          map[string]*string{},
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunnerGrpSettings).Error
	require.NoError(s.T(), err)
}

func (s *AdminUpdateInstallRunnerTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminUpdateInstallRunnerTestSuite) TestAdminUpdateInstallRunner() {
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
			name: "successfully update runner image tag",
			setupFunc: func() string {
				return s.testInstall.ID
			},
			requestBody: AdminUpdateInstallRunnerRequest{
				ContainerImageTag: "v2.0.0",
			},
			expectedCode:   http.StatusOK,
			expectedSignal: true,
			validateFunc: func(installID string) {
				// Verify settings were updated
				var settings app.RunnerGroupSettings
				err := s.service.DB.Where("runner_group_id = ?", s.testRunnerGrp.ID).First(&settings).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "v2.0.0", settings.ContainerImageTag)

				// Verify signal was sent
				capturedSignals := s.mockEvClient.GetSignals()
				require.Len(s.T(), capturedSignals, 1)
				assert.Equal(s.T(), installID, capturedSignals[0].ID)

				sig, ok := capturedSignals[0].Signal.(*signals.Signal)
				require.True(s.T(), ok)
				assert.Equal(s.T(), signals.OperationReprovisionRunner, sig.Type)
			},
		},
		{
			name: "update with different tag version",
			setupFunc: func() string {
				return s.testInstall.ID
			},
			requestBody: AdminUpdateInstallRunnerRequest{
				ContainerImageTag: "latest",
			},
			expectedCode:   http.StatusOK,
			expectedSignal: true,
			validateFunc: func(installID string) {
				var settings app.RunnerGroupSettings
				err := s.service.DB.Where("runner_group_id = ?", s.testRunnerGrp.ID).First(&settings).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "latest", settings.ContainerImageTag)
			},
		},
		{
			name: "empty body returns bad request",
			setupFunc: func() string {
				return s.testInstall.ID
			},
			requestBody:    nil,
			expectedCode:   http.StatusBadRequest,
			expectedSignal: false,
		},
		{
			name: "nonexistent install returns error",
			setupFunc: func() string {
				return "ins000000000000000000000000"
			},
			requestBody: AdminUpdateInstallRunnerRequest{
				ContainerImageTag: "v2.0.0",
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
			rr := s.makeRequest("PATCH", "/v1/installs/"+installID+"/admin-update-runner", tc.requestBody)

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
