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
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type GetOrgStatsTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	OrgsService     *service
	Seeder          *testseed.Seeder
}

type GetOrgStatsTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service GetOrgStatsTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestGetOrgStatsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetOrgStatsTestSuite))
}

func (s *GetOrgStatsTestSuite) SetupSuite() {
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

func (s *GetOrgStatsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetOrgStatsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetOrgStatsTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *GetOrgStatsTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetOrgStatsTestSuite) TestGetOrgStats() {
	testCases := []struct {
		name              string
		setupFunc         func()
		expectedAppCount  int64
		expectedInstCount int64
		expectedInstNames []string
	}{
		{
			name: "returns zeros when no apps or installs",
			setupFunc: func() {
			},
			expectedAppCount:  0,
			expectedInstCount: 0,
			expectedInstNames: []string{},
		},
		{
			name: "returns correct app count",
			setupFunc: func() {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				app1 := &app.App{
					ID:          domains.NewAppID(),
					Name:        "test-app-1",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusActive,
				}
				app2 := &app.App{
					ID:          domains.NewAppID(),
					Name:        "test-app-2",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusActive,
				}

				err := s.service.DB.WithContext(ctx).Create(app1).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", app1.ID)
				})

				err = s.service.DB.WithContext(ctx).Create(app2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", app2.ID)
				})
			},
			expectedAppCount:  2,
			expectedInstCount: 0,
			expectedInstNames: []string{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setupFunc()

			rr := s.makeRequest(http.MethodGet, "/v1/orgs/current/stats")

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var response OrgStatsResponse
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)

			assert.Equal(s.T(), tc.expectedAppCount, response.AppCount,
				"app_count mismatch")
			assert.Equal(s.T(), tc.expectedInstCount, response.InstallCount,
				"install_count mismatch")

			require.NotNil(s.T(), response.InstallNames, "install_names should not be nil")
			assert.Equal(s.T(), len(tc.expectedInstNames), len(response.InstallNames),
				"install_names length mismatch")

			if len(tc.expectedInstNames) > 0 {
				for _, expectedName := range tc.expectedInstNames {
					assert.Contains(s.T(), response.InstallNames, expectedName,
						"expected install name not found: %s", expectedName)
				}
			}
		})
	}
}
