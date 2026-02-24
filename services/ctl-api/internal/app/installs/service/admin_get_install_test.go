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

type AdminGetInstallTestService struct {
	fx.In
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	Seeder          *testseed.Seeder
	InstallsService *service
}

type AdminGetInstallTestSuite struct {
	tests.BaseDBTestSuite
	app                   *fxtest.App
	service               AdminGetInstallTestService
	router                *gin.Engine
	testOrg               *app.Org
	testAcc               *app.Account
	testApp               *app.App
	testInstall           *app.Install
	testRunner            *app.Runner
	testRunnerGrp         *app.RunnerGroup
	testRunnerGrpSettings *app.RunnerGroupSettings
}

func TestAdminGetInstallSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminGetInstallTestSuite))
}

func (s *AdminGetInstallTestSuite) SetupSuite() {
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

func (s *AdminGetInstallTestSuite) SetupTest() {
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

func (s *AdminGetInstallTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminGetInstallTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(ctx, s.T())
	s.testInstall = s.service.Seeder.CreateInstall(ctx, s.T())

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
		RunnerGroupID:     s.testRunnerGrp.ID,
		ContainerImageTag: "v1.0.0",
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

	// Update install with runner group
	s.testInstall.RunnerGroupID = s.testRunnerGrp.ID
	err = s.service.DB.WithContext(ctx).Save(s.testInstall).Error
	require.NoError(s.T(), err)
}

func (s *AdminGetInstallTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminGetInstallTestSuite) TestAdminGetInstall() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(*app.Install)
		expectedNotFound bool
	}{
		{
			name: "successfully get install with all preloads",
			setupFunc: func() string {
				return s.testInstall.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(install *app.Install) {
				assert.Equal(s.T(), s.testInstall.ID, install.ID)
				assert.Equal(s.T(), s.testInstall.Name, install.Name)
				assert.NotNil(s.T(), install.App, "App should be preloaded")
				assert.NotNil(s.T(), install.App.Org, "App.Org should be preloaded")
				assert.NotNil(s.T(), install.CreatedBy, "CreatedBy should be preloaded")
				assert.NotNil(s.T(), install.RunnerGroup, "RunnerGroup should be preloaded")
				assert.NotNil(s.T(), install.RunnerGroup.Settings, "RunnerGroup.Settings should be preloaded")
				assert.NotEmpty(s.T(), install.RunnerGroup.Runners, "RunnerGroup.Runners should be preloaded")
				assert.Equal(s.T(), s.testRunner.ID, install.RunnerGroup.Runners[0].ID)
			},
		},
		{
			name: "get install by name",
			setupFunc: func() string {
				return s.testInstall.Name
			},
			expectedCode: http.StatusOK,
			validateFunc: func(install *app.Install) {
				assert.Equal(s.T(), s.testInstall.ID, install.ID)
				assert.Equal(s.T(), s.testInstall.Name, install.Name)
			},
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
			rr := s.makeRequest("GET", "/v1/installs/"+installID+"/admin-get", nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}

			if tc.validateFunc != nil && rr.Code == http.StatusOK {
				var install app.Install
				err := json.Unmarshal(rr.Body.Bytes(), &install)
				require.NoError(s.T(), err)
				tc.validateFunc(&install)
			}
		})
	}
}
