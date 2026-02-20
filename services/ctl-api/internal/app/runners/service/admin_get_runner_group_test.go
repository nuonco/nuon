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

type AdminGetRunnerGroupTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminGetRunnerGroupTestSuite struct {
	tests.BaseDBTestSuite
	app                   *fxtest.App
	service               AdminGetRunnerGroupTestService
	router                *gin.Engine
	testOrg               *app.Org
	testAcc               *app.Account
	testRunnerGrp         *app.RunnerGroup
	testRunnerGrpSettings *app.RunnerGroupSettings
	testRunner            *app.Runner
}

func TestAdminGetRunnerGroupSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminGetRunnerGroupTestSuite))
}

func (s *AdminGetRunnerGroupTestSuite) SetupSuite() {
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

func (s *AdminGetRunnerGroupTestSuite) SetupTest() {
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

func (s *AdminGetRunnerGroupTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminGetRunnerGroupTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create runner group (requires account context for created_by_id)
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

	// Create runner group settings (requires account context for created_by_id)
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

	// Create runner (requires account context for created_by_id)
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

func (s *AdminGetRunnerGroupTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminGetRunnerGroupTestSuite) TestAdminGetRunnerGroup() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(*app.RunnerGroup)
		expectedNotFound bool
	}{
		{
			name: "successfully get runner group with all associations",
			setupFunc: func() string {
				return s.testRunnerGrp.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(rg *app.RunnerGroup) {
				assert.Equal(s.T(), s.testRunnerGrp.ID, rg.ID)
				assert.Equal(s.T(), s.testOrg.ID, rg.OrgID)
				assert.Equal(s.T(), app.RunnerGroupTypeOrg, rg.Type)
				assert.Equal(s.T(), app.AppRunnerTypeAWSEKS, rg.Platform)

				// Verify CreatedByID is set (CreatedBy preload may be zero-value)
				assert.Equal(s.T(), s.testAcc.ID, rg.CreatedByID)

				// Verify Settings is preloaded
				assert.NotNil(s.T(), rg.Settings)
				assert.Equal(s.T(), s.testRunnerGrpSettings.ID, rg.Settings.ID)
				assert.Equal(s.T(), "test.ecr.aws/runner", rg.Settings.ContainerImageURL)

				// Verify Runners is preloaded
				assert.NotEmpty(s.T(), rg.Runners)
				assert.Len(s.T(), rg.Runners, 1)
				assert.Equal(s.T(), s.testRunner.ID, rg.Runners[0].ID)
			},
		},
		{
			name: "runner group not found returns error",
			setupFunc: func() string {
				return "rngnonexistent123456789012"
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "runner group with multiple runners",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create additional runner in same group
				runner2 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "test-runner-2",
					DisplayName:   "Test Runner 2",
					Status:        app.RunnerStatusPending,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err := s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner2)
				})

				return s.testRunnerGrp.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(rg *app.RunnerGroup) {
				// Should have 2 runners now
				assert.Len(s.T(), rg.Runners, 2)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerGroupID := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/runner-groups/"+runnerGroupID)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else if tc.validateFunc != nil {
				var rg app.RunnerGroup
				err := json.Unmarshal(rr.Body.Bytes(), &rg)
				require.NoError(s.T(), err)
				tc.validateFunc(&rg)
			}
		})
	}
}
