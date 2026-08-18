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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
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
		tests.CtlApiFXOptions(s.T()),
		// flowclient is intentionally not in tests.CtlApiFXOptions (see the note
		// in tests/testfx.go); the installs service constructor needs it.
		fx.Provide(flowclient.New),
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
	ctx, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(ctx, s.T())
	s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)
	s.testInstall = s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
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
			name: "get all sandbox installs",
			setupFunc: func() {
				// testInstall already exists from setupTestData
			},
			queryParams:  "?type=sandbox",
			expectedCode: http.StatusOK,
			validateFunc: func(installs []*app.Install) {
				assert.GreaterOrEqual(s.T(), len(installs), 1, "should have at least one install")
				// Find our test install in results
				found := false
				for _, install := range installs {
					if install.ID == s.testInstall.ID {
						found = true
						assert.NotEmpty(s.T(), install.App.ID, "App should be preloaded")
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
				ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
				ctx = cctx.SetAccountIDContext(ctx, s.testAcc.ID)
				// Create 3 more installs
				for i := 0; i < 3; i++ {
					install := s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Delete(install)
					})
				}
			},
			queryParams:  "?type=sandbox&limit=2",
			expectedCode: http.StatusOK,
			validateFunc: func(installs []*app.Install) {
				assert.LessOrEqual(s.T(), len(installs), 2, "should respect limit parameter")
			},
		},
		{
			name: "filter by org type sandbox",
			setupFunc: func() {
				// testOrg already has OrgType=sandbox via seeder
			},
			queryParams:  "?type=sandbox",
			expectedCode: http.StatusOK,
			validateFunc: func(installs []*app.Install) {
				// Should include our sandbox test install
				found := false
				for _, install := range installs {
					if install.ID == s.testInstall.ID {
						found = true
						break
					}
				}
				assert.True(s.T(), found, "sandbox test install should appear in sandbox type filter")
			},
		},
		{
			name: "filter by org type real returns no test installs",
			setupFunc: func() {
				// All test orgs are sandbox, so filtering by "real" should exclude them
			},
			queryParams:  "?type=real",
			expectedCode: http.StatusOK,
			validateFunc: func(installs []*app.Install) {
				// Our sandbox test install should NOT appear in real type filter
				for _, install := range installs {
					assert.NotEqual(s.T(), s.testInstall.ID, install.ID, "sandbox test install should not appear in real type filter")
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

// createAppRunnerConfig persists an AppRunnerConfig of an arbitrary type, unlike
// testseed's CreateAppRunnerConfig which always creates an "aws" one.
func (s *GetAllInstallsTestSuite) createAppRunnerConfig(ctx context.Context, appConfigID string, runnerType app.AppRunnerType) *app.AppRunnerConfig {
	runner := &app.AppRunnerConfig{
		AppID:       s.testApp.ID,
		AppConfigID: appConfigID,
		Type:        runnerType,
	}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(runner).Error)
	return runner
}

// pinInstallConfig overwrites an install's app_config_id/app_runner_config_id
// directly, mirroring the bulk `UPDATE installs SET app_runner_config_id = ...`
// that internal/pkg/config/syncer/runner/sync.go runs on every app config sync.
func (s *GetAllInstallsTestSuite) pinInstallConfig(ctx context.Context, installID, appConfigID, appRunnerConfigID string) {
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Model(&app.Install{}).
		Where(app.Install{ID: installID}).
		Updates(map[string]any{
			"app_config_id":        appConfigID,
			"app_runner_config_id": appRunnerConfigID,
		}).Error)
}

// TestGetAllInstallsResolvesCloudPlatform covers get_all_installs.go's
// Preload("AppRunnerConfig") / Preload("AppConfig.RunnerConfig") fixes: before
// them, this admin endpoint had no runner-config preload at all and every
// install came back with cloud_platform "unknown".
func (s *GetAllInstallsTestSuite) TestGetAllInstallsResolvesCloudPlatform() {
	ctx := context.Background()
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	ctx = cctx.SetAccountIDContext(ctx, s.testAcc.ID)

	testCases := []struct {
		name                  string
		setup                 func() *app.Install
		expectedCloudPlatform app.CloudPlatform
		expectedRunnerType    app.AppRunnerType
	}{
		{
			name: "pinned app config runner wins over stale app_runner_config_id FK",
			setup: func() *app.Install {
				azureCfg := s.service.Seeder.CreateBareAppConfig(ctx, s.T(), s.testApp.ID)
				s.createAppRunnerConfig(ctx, azureCfg.ID, app.AppRunnerTypeAzure)

				// Represents the runner config a later app config sync created — the
				// syncer rewrites app_runner_config_id on every install for the app to
				// point at this, without touching the install's pinned app_config_id.
				staleCfg := s.service.Seeder.CreateBareAppConfig(ctx, s.T(), s.testApp.ID)
				awsRunner := s.createAppRunnerConfig(ctx, staleCfg.ID, app.AppRunnerTypeAWS)

				install := s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
				s.pinInstallConfig(ctx, install.ID, azureCfg.ID, awsRunner.ID)
				return install
			},
			expectedCloudPlatform: app.CloudPlatformAzure,
			expectedRunnerType:    app.AppRunnerTypeAzure,
		},
		{
			name: "falls back to app_runner_config_id FK when pinned config has no runner config",
			setup: func() *app.Install {
				noRunnerCfg := s.service.Seeder.CreateBareAppConfig(ctx, s.T(), s.testApp.ID)

				throwawayCfg := s.service.Seeder.CreateBareAppConfig(ctx, s.T(), s.testApp.ID)
				azureRunner := s.createAppRunnerConfig(ctx, throwawayCfg.ID, app.AppRunnerTypeAzure)

				install := s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
				s.pinInstallConfig(ctx, install.ID, noRunnerCfg.ID, azureRunner.ID)
				return install
			},
			expectedCloudPlatform: app.CloudPlatformAzure,
			expectedRunnerType:    app.AppRunnerTypeAzure,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			install := tc.setup()
			s.T().Cleanup(func() {
				s.service.DB.Unscoped().Delete(install)
			})

			rr := s.makeRequest(http.MethodGet, "/v1/installs?type=sandbox&limit=60", nil)
			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var installs []*app.Install
			require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &installs))

			var found *app.Install
			for _, i := range installs {
				if i.ID == install.ID {
					found = i
					break
				}
			}
			require.NotNil(s.T(), found, "install %s should be present in results", install.ID)
			assert.Equal(s.T(), tc.expectedCloudPlatform, found.CloudPlatform)
			assert.Equal(s.T(), tc.expectedRunnerType, found.RunnerType)
		})
	}
}
