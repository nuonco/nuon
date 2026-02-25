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
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AdminGetInstallRunnerTestService struct {
	fx.In
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	Seeder          *testseed.Seeder
	InstallsService *service
}

type AdminGetInstallRunnerTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       AdminGetInstallRunnerTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testApp       *app.App
	testInstall   *app.Install
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
}

func TestAdminGetInstallRunnerSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminGetInstallRunnerTestSuite))
}

func (s *AdminGetInstallRunnerTestSuite) SetupSuite() {
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

func (s *AdminGetInstallRunnerTestSuite) SetupTest() {
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

func (s *AdminGetInstallRunnerTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminGetInstallRunnerTestSuite) setupTestData() {
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

	// Create runner in the runner group
	s.testRunner = &app.Runner{
		OrgID:         s.testOrg.ID,
		Name:          "test-runner-" + s.testInstall.ID,
		DisplayName:   "Test Runner",
		Status:        app.RunnerStatusActive,
		RunnerGroupID: s.testRunnerGrp.ID,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunner).Error
	require.NoError(s.T(), err)
}

func (s *AdminGetInstallRunnerTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminGetInstallRunnerTestSuite) TestAdminGetInstallRunner() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(*app.Runner)
		expectedNotFound bool
	}{
		{
			name: "successfully get runner via install",
			setupFunc: func() string {
				return s.testInstall.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(runner *app.Runner) {
				assert.Equal(s.T(), s.testRunner.ID, runner.ID)
				assert.Equal(s.T(), s.testRunner.Name, runner.Name)
				assert.Equal(s.T(), s.testRunner.DisplayName, runner.DisplayName)
				assert.Equal(s.T(), s.testRunner.Status, runner.Status)
			},
		},
		{
			name: "install without runners returns error",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
				ctx = cctx.SetAccountIDContext(ctx, s.testAcc.ID)

				// Create a second install with a runner group that has no runners
				install2 := s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)

				rg2 := &app.RunnerGroup{
					OrgID:     s.testOrg.ID,
					OwnerID:   install2.ID,
					OwnerType: "installs",
					Type:      app.RunnerGroupTypeInstall,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err := s.service.DB.WithContext(ctx).Create(rg2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(rg2)
					s.service.DB.Unscoped().Delete(install2)
				})

				return install2.ID
			},
			expectedCode:     http.StatusInternalServerError,
			expectedNotFound: true,
		},
		{
			name: "nonexistent install returns error",
			setupFunc: func() string {
				return "ins000000000000000000000000"
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			installID := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/installs/"+installID+"/admin-get-runner", nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}

			if tc.validateFunc != nil && rr.Code == http.StatusOK {
				var runner app.Runner
				err := json.Unmarshal(rr.Body.Bytes(), &runner)
				require.NoError(s.T(), err)
				tc.validateFunc(&runner)
			}
		})
	}
}
