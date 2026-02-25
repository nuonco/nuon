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
