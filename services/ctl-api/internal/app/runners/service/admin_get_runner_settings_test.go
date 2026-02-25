package service

import (
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

type AdminGetRunnerSettingsTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminGetRunnerSettingsTestSuite struct {
	tests.BaseDBTestSuite
	app                   *fxtest.App
	service               AdminGetRunnerSettingsTestService
	router                *gin.Engine
	testOrg               *app.Org
	testAcc               *app.Account
	testRunner            *app.Runner
	testRunnerGrp         *app.RunnerGroup
	testRunnerGrpSettings *app.RunnerGroupSettings
}

func TestAdminGetRunnerSettingsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminGetRunnerSettingsTestSuite))
}

func (s *AdminGetRunnerSettingsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AdminGetRunnerSettingsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with internal routes (no org context for admin routes)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.RunnersService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminGetRunnerSettingsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminGetRunnerSettingsTestSuite) setupTestData() {
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

	// Create runner group settings
	s.testRunnerGrpSettings = &app.RunnerGroupSettings{
		ID:                domains.NewRunnerGroupSettingsID(),
		OrgID:             s.testOrg.ID,
		RunnerGroupID:     s.testRunnerGrp.ID,
		ContainerImageURL: "test.ecr.aws/runner",
		ContainerImageTag: "v1.0.0",
		RunnerAPIURL:      "https://runner-api.test.com",
		SandboxMode:       true,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunnerGrpSettings).Error
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

func (s *AdminGetRunnerSettingsTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminGetRunnerSettingsTestSuite) TestAdminGetRunnerSettings() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(*app.RunnerGroupSettings)
		expectedNotFound bool
	}{
		{
			name: "successfully get runner settings",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(settings *app.RunnerGroupSettings) {
				assert.Equal(s.T(), s.testRunnerGrpSettings.ID, settings.ID)
				assert.Equal(s.T(), s.testRunnerGrp.ID, settings.RunnerGroupID)
				assert.Equal(s.T(), "test.ecr.aws/runner", settings.ContainerImageURL)
				assert.Equal(s.T(), "v1.0.0", settings.ContainerImageTag)
				assert.Equal(s.T(), "https://runner-api.test.com", settings.RunnerAPIURL)
				assert.True(s.T(), settings.SandboxMode)
			},
		},
		{
			name: "runner not found returns error",
			setupFunc: func() string {
				return "rnrnonexistent123456789012"
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "runner with different settings values",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create runner group with different settings
				runnerGrp2 := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   s.testOrg.ID,
					OwnerType: "org",
					Type:      app.RunnerGroupTypeOrg,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err := s.service.DB.WithContext(ctx).Create(runnerGrp2).Error
				require.NoError(s.T(), err)

				settings2 := &app.RunnerGroupSettings{
					ID:                domains.NewRunnerGroupSettingsID(),
					OrgID:             s.testOrg.ID,
					RunnerGroupID:     runnerGrp2.ID,
					ContainerImageURL: "custom.ecr.aws/runner",
					ContainerImageTag: "v2.0.0",
					RunnerAPIURL:      "https://custom-api.test.com",
					SandboxMode:       false,
				}
				err = s.service.DB.WithContext(ctx).Create(settings2).Error
				require.NoError(s.T(), err)

				runner2 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "runner-2",
					DisplayName:   "Runner 2",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: runnerGrp2.ID,
				}
				err = s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner2)
					s.service.DB.Unscoped().Delete(settings2)
					s.service.DB.Unscoped().Delete(runnerGrp2)
				})

				return runner2.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(settings *app.RunnerGroupSettings) {
				assert.Equal(s.T(), "custom.ecr.aws/runner", settings.ContainerImageURL)
				assert.Equal(s.T(), "v2.0.0", settings.ContainerImageTag)
				assert.Equal(s.T(), "https://custom-api.test.com", settings.RunnerAPIURL)
				assert.False(s.T(), settings.SandboxMode)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/runners/"+runnerID+"/settings")

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else if tc.validateFunc != nil {
				var settings app.RunnerGroupSettings
				err := json.Unmarshal(rr.Body.Bytes(), &settings)
				require.NoError(s.T(), err)
				tc.validateFunc(&settings)
			}
		})
	}
}
