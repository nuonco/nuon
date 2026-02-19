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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AdminCreateInstallRunnerShutdownJobTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminCreateInstallRunnerShutdownJobTestSuite struct {
	tests.BaseDBTestSuite
	app          *fxtest.App
	service      AdminCreateInstallRunnerShutdownJobTestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	mockEvClient *tests.FakeEventLoopClient
}

func TestAdminCreateInstallRunnerShutdownJobSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminCreateInstallRunnerShutdownJobTestSuite))
}

func (s *AdminCreateInstallRunnerShutdownJobTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	s.mockEvClient = tests.NewFakeEventLoopClient()

	options := append(
		tests.CtlApiFXOptions(),
		fx.Provide(New),
		fx.Decorate(func() eventloop.Client {
			return s.mockEvClient
		}),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AdminCreateInstallRunnerShutdownJobTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.mockEvClient.Reset()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.RunnersService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminCreateInstallRunnerShutdownJobTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminCreateInstallRunnerShutdownJobTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
}

func (s *AdminCreateInstallRunnerShutdownJobTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminCreateInstallRunnerShutdownJobTestSuite) TestAdminCreateInstallRunnerShutdownJob() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		requestBody      interface{}
		expectedCode     int
		validateFunc     func(string)
		expectedNotFound bool
	}{
		{
			name: "nonexistent install returns error",
			setupFunc: func() string {
				return "insnonexistent123456789012"
			},
			requestBody:      AdminCreateInstallRunnerShutDownJobRequest{},
			expectedCode:     http.StatusInternalServerError,
			expectedNotFound: true,
		},
		{
			name: "empty body accepted",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				testApp := &app.App{
					ID:    domains.NewAppID(),
					OrgID: s.testOrg.ID,
					Name:  "test-app",
				}
				err := s.service.DB.WithContext(ctx).Create(testApp).Error
				require.NoError(s.T(), err)

				runnerGrp := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   testApp.ID,
					OwnerType: "app",
					Type:      app.RunnerGroupTypeInstall,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err = s.service.DB.WithContext(ctx).Create(runnerGrp).Error
				require.NoError(s.T(), err)

				runner1 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "install-runner-1",
					DisplayName:   "Install Runner 1",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: runnerGrp.ID,
				}
				err = s.service.DB.WithContext(ctx).Create(runner1).Error
				require.NoError(s.T(), err)

				runner2 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "install-runner-2",
					DisplayName:   "Install Runner 2",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: runnerGrp.ID,
				}
				err = s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)

				install := &app.Install{
					ID:    domains.NewInstallID(),
					OrgID: s.testOrg.ID,
					AppID: testApp.ID,
					Name:  "test-install",
				}
				err = s.service.DB.WithContext(ctx).Create(install).Error
				require.NoError(s.T(), err)

				err = s.service.DB.WithContext(ctx).Model(runnerGrp).Update("owner_id", install.ID).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(install)
					s.service.DB.Unscoped().Delete(runner1)
					s.service.DB.Unscoped().Delete(runner2)
					s.service.DB.Unscoped().Delete(runnerGrp)
					s.service.DB.Unscoped().Delete(testApp)
				})

				return install.ID
			},
			requestBody:  nil,
			expectedCode: http.StatusCreated,
			validateFunc: func(installID string) {
				var install app.Install
				err := s.service.DB.
					Preload("RunnerGroup.Runners").
					First(&install, "id = ?", installID).Error
				require.NoError(s.T(), err)

				var jobs []app.RunnerJob
				runnerIDs := []string{}
				for _, runner := range install.RunnerGroup.Runners {
					runnerIDs = append(runnerIDs, runner.ID)
				}

				err = s.service.DB.
					Where("runner_id IN ? AND type = ?", runnerIDs, app.RunnerJobTypeShutDown).
					Find(&jobs).Error
				require.NoError(s.T(), err)
				assert.Len(s.T(), jobs, 2, "should create shutdown jobs for both runners")

				signals := s.mockEvClient.GetSignals()
				assert.Len(s.T(), signals, 2, "should send signals for both runners")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			installID := tc.setupFunc()
			rr := s.makeRequest("POST", "/v1/installs/"+installID+"/runners/shutdown-job", tc.requestBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else if tc.validateFunc != nil {
				tc.validateFunc(installID)
			}
		})
	}
}
