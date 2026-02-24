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
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type UpdateInstallSandboxTestService struct {
	fx.In
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	Seeder          *testseed.Seeder
	InstallsService *service
}

type UpdateInstallSandboxTestSuite struct {
	tests.BaseDBTestSuite
	app               *fxtest.App
	service           UpdateInstallSandboxTestService
	router            *gin.Engine
	testOrg           *app.Org
	testAcc           *app.Account
	testApp           *app.App
	testInstall       *app.Install
	testSandboxConfig *app.AppSandboxConfig
}

func TestUpdateInstallSandboxSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(UpdateInstallSandboxTestSuite))
}

func (s *UpdateInstallSandboxTestSuite) SetupSuite() {
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
}

func (s *UpdateInstallSandboxTestSuite) SetupTest() {
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

func (s *UpdateInstallSandboxTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *UpdateInstallSandboxTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(ctx, s.T())
	s.testInstall = s.service.Seeder.CreateInstall(ctx, s.T())

	// Create sandbox config
	s.testSandboxConfig = &app.AppSandboxConfig{
		ID:    domains.NewAppSandboxConfigID(),
		AppID: s.testApp.ID,
		OrgID: s.testOrg.ID,
	}
	err := s.service.DB.WithContext(ctx).Create(s.testSandboxConfig).Error
	require.NoError(s.T(), err)
}

func (s *UpdateInstallSandboxTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *UpdateInstallSandboxTestSuite) TestAdminUpdateSandbox() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		requestBody      interface{}
		expectedCode     int
		validateFunc     func(string)
		expectedNotFound bool
	}{
		{
			name: "successfully update install to latest sandbox",
			setupFunc: func() string {
				return s.testInstall.ID
			},
			requestBody:  AdminUpdateSandboxRequest{},
			expectedCode: http.StatusOK,
			validateFunc: func(installID string) {
				var install app.Install
				err := s.service.DB.Where("id = ?", installID).First(&install).Error
				require.NoError(s.T(), err)
			},
		},
		{
			name: "update with empty body",
			setupFunc: func() string {
				return s.testInstall.ID
			},
			requestBody:  nil,
			expectedCode: http.StatusOK,
		},
		{
			name: "install without sandbox config returns error",
			setupFunc: func() string {
				ctx := context.Background()

				// Create app without sandbox config
				app2 := &app.App{
					ID:    domains.NewAppID(),
					OrgID: s.testOrg.ID,
					Name:  "app-no-sandbox",
				}
				err := s.service.DB.WithContext(ctx).Create(app2).Error
				require.NoError(s.T(), err)

				install2 := &app.Install{
					ID:    domains.NewInstallID(),
					OrgID: s.testOrg.ID,
					AppID: app2.ID,
					Name:  "install-no-sandbox",
				}
				err = s.service.DB.WithContext(ctx).Create(install2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(install2)
					s.service.DB.Unscoped().Delete(app2)
				})

				return install2.ID
			},
			requestBody:      AdminUpdateSandboxRequest{},
			expectedCode:     http.StatusInternalServerError,
			expectedNotFound: true,
		},
		{
			name: "nonexistent install returns error",
			setupFunc: func() string {
				return "ins000000000000000000000000"
			},
			requestBody:      AdminUpdateSandboxRequest{},
			expectedCode:     http.StatusInternalServerError,
			expectedNotFound: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			installID := tc.setupFunc()
			rr := s.makeRequest("POST", "/v1/installs/"+installID+"/admin-update-sandbox", tc.requestBody)

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
		})
	}
}
