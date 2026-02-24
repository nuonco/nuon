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

type GetAllInstallsTestService struct {
	fx.In
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	Seeder          *testseed.Seeder
	InstallsService *service
}

type GetAllInstallsTestSuite struct {
	tests.BaseDBTestSuite
	app         *fxtest.App
	service     GetAllInstallsTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	testApp     *app.App
	testInstall *app.Install
}

func TestGetAllInstallsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetAllInstallsTestSuite))
}

func (s *GetAllInstallsTestSuite) SetupSuite() {
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

func (s *GetAllInstallsTestSuite) SetupTest() {
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

func (s *GetAllInstallsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetAllInstallsTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(ctx, s.T())
	s.testInstall = s.service.Seeder.CreateInstall(ctx, s.T())
}

func (s *GetAllInstallsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *GetAllInstallsTestSuite) TestGetAllInstalls() {
	testCases := []struct {
		name         string
		setupFunc    func()
		queryParams  string
		expectedCode int
		validateFunc func([]*app.Install)
	}{
		{
			name: "get all installs with default params",
			setupFunc: func() {
				// testInstall already exists from setupTestData
			},
			queryParams:  "",
			expectedCode: http.StatusOK,
			validateFunc: func(installs []*app.Install) {
				assert.GreaterOrEqual(s.T(), len(installs), 1, "should have at least one install")
				// Find our test install in results
				found := false
				for _, install := range installs {
					if install.ID == s.testInstall.ID {
						found = true
						assert.NotNil(s.T(), install.App, "App should be preloaded")
						assert.NotNil(s.T(), install.App.Org, "App.Org should be preloaded")
						break
					}
				}
				assert.True(s.T(), found, "test install should be in results")
			},
		},
		{
			name: "get installs with custom limit",
			setupFunc: func() {
				ctx := context.Background()
				// Create 3 more installs
				for i := 0; i < 3; i++ {
					install := s.service.Seeder.CreateInstall(ctx, s.T())
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Delete(install)
					})
				}
			},
			queryParams:  "?limit=2",
			expectedCode: http.StatusOK,
			validateFunc: func(installs []*app.Install) {
				assert.LessOrEqual(s.T(), len(installs), 2, "should respect limit parameter")
			},
		},
		{
			name: "filter by org type real",
			setupFunc: func() {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create a real org (non-sandbox)
				realOrg := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "real-org",
					SandboxMode: false,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/real",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(realOrg).Error
				require.NoError(s.T(), err)

				realApp := &app.App{
					ID:    domains.NewAppID(),
					OrgID: realOrg.ID,
					Name:  "real-app",
				}
				err = s.service.DB.WithContext(ctx).Create(realApp).Error
				require.NoError(s.T(), err)

				realInstall := &app.Install{
					ID:    domains.NewInstallID(),
					OrgID: realOrg.ID,
					AppID: realApp.ID,
					Name:  "real-install",
				}
				err = s.service.DB.WithContext(ctx).Create(realInstall).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(realInstall)
					s.service.DB.Unscoped().Delete(realApp)
					s.service.DB.Unscoped().Delete(realOrg)
				})
			},
			queryParams:  "?type=real",
			expectedCode: http.StatusOK,
			validateFunc: func(installs []*app.Install) {
				// All returned installs should be from non-sandbox orgs
				for _, install := range installs {
					if install.App != nil && install.App.Org != nil {
						assert.False(s.T(), install.App.Org.SandboxMode, "should only return real (non-sandbox) orgs")
					}
				}
			},
		},
		{
			name: "filter by org type sandbox",
			setupFunc: func() {
				// testOrg is already sandbox mode
			},
			queryParams:  "?type=sandbox",
			expectedCode: http.StatusOK,
			validateFunc: func(installs []*app.Install) {
				// All returned installs should be from sandbox orgs
				for _, install := range installs {
					if install.App != nil && install.App.Org != nil {
						assert.True(s.T(), install.App.Org.SandboxMode, "should only return sandbox orgs")
					}
				}
			},
		},
		{
			name: "invalid limit returns error",
			setupFunc: func() {
				// No setup needed
			},
			queryParams:  "?limit=invalid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/installs"+tc.queryParams, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil && rr.Code == http.StatusOK {
				var installs []*app.Install
				err := json.Unmarshal(rr.Body.Bytes(), &installs)
				require.NoError(s.T(), err)
				tc.validateFunc(installs)
			}
		})
	}
}
