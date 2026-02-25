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

type AdminOrgInstallsTestService struct {
	fx.In
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	Seeder          *testseed.Seeder
	InstallsService *service
}

type AdminOrgInstallsTestSuite struct {
	tests.BaseDBTestSuite
	app         *fxtest.App
	service     AdminOrgInstallsTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	testApp     *app.App
	testInstall *app.Install
}

func TestAdminOrgInstallsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminOrgInstallsTestSuite))
}

func (s *AdminOrgInstallsTestSuite) SetupSuite() {
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

func (s *AdminOrgInstallsTestSuite) SetupTest() {
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

func (s *AdminOrgInstallsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminOrgInstallsTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	ctx, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(ctx, s.T())
	s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)
	s.testInstall = s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
}

func (s *AdminOrgInstallsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminOrgInstallsTestSuite) TestAdminGetOrgInstalls() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		expectedCode int
		validateFunc func([]*app.Install, string)
	}{
		{
			name: "successfully get org installs",
			setupFunc: func() string {
				return s.testOrg.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(installs []*app.Install, orgID string) {
				assert.GreaterOrEqual(s.T(), len(installs), 1, "should have at least one install")
				found := false
				for _, install := range installs {
					assert.Equal(s.T(), orgID, install.OrgID, "all installs should belong to org")
					if install.ID == s.testInstall.ID {
						found = true
					}
				}
				assert.True(s.T(), found, "test install should be in results")
			},
		},
		{
			name: "get installs for org with multiple installs",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
				ctx = cctx.SetAccountIDContext(ctx, s.testAcc.ID)
				// Create 2 more installs for the same org
				for i := 0; i < 2; i++ {
					install := s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Delete(install)
					})
				}
				return s.testOrg.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(installs []*app.Install, orgID string) {
				assert.GreaterOrEqual(s.T(), len(installs), 3, "should have at least 3 installs")
			},
		},
		{
			name: "org with no installs returns empty array",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org without installs
				org2 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "org-no-installs",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/empty",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(org2)
				})

				return org2.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(installs []*app.Install, orgID string) {
				assert.Empty(s.T(), installs, "org without installs should return empty array")
			},
		},
		{
			name: "installs from different org are not returned",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create another org with installs using seeder helpers
				ctx, org2 := s.service.Seeder.EnsureOrg(ctx, s.T())
				app2 := s.service.Seeder.CreateApp(ctx, s.T())
				s.service.Seeder.CreateAppConfig(ctx, s.T(), app2.ID)
				install2 := s.service.Seeder.CreateInstall(ctx, s.T(), app2)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(install2)
					s.service.DB.Unscoped().Delete(app2)
					s.service.DB.Unscoped().Delete(org2)
				})

				return s.testOrg.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(installs []*app.Install, orgID string) {
				// Should only return installs for testOrg, not org2
				for _, install := range installs {
					assert.Equal(s.T(), orgID, install.OrgID, "should only return installs for requested org")
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			orgID := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/orgs/"+orgID+"/admin-get-installs", nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil && rr.Code == http.StatusOK {
				var installs []*app.Install
				err := json.Unmarshal(rr.Body.Bytes(), &installs)
				require.NoError(s.T(), err)
				tc.validateFunc(installs, orgID)
			}
		})
	}
}
