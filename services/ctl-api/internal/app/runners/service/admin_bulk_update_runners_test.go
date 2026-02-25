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
	"github.com/jackc/pgx/v5/pgtype"
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

type AdminBulkUpdateRunnersTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminBulkUpdateRunnersTestSuite struct {
	tests.BaseDBTestSuite
	app          *fxtest.App
	service      AdminBulkUpdateRunnersTestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	mockEvClient *tests.MockEventLoopClient
}

func TestAdminBulkUpdateRunnersSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminBulkUpdateRunnersTestSuite))
}

func (s *AdminBulkUpdateRunnersTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	s.mockEvClient = tests.NewMockEventLoopClient()

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			Mocks: &tests.TestMocks{MockEv: s.mockEvClient},

			CustomValidator: true,
		}),

		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AdminBulkUpdateRunnersTestSuite) SetupTest() {
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

func (s *AdminBulkUpdateRunnersTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminBulkUpdateRunnersTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
}

func (s *AdminBulkUpdateRunnersTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminBulkUpdateRunnersTestSuite) TestAdminBulkUpdateRunners() {
	testCases := []struct {
		name         string
		setupFunc    func() []string
		requestBody  interface{}
		expectedCode int
		validateFunc func([]string)
	}{
		{
			name: "successful bulk update",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runnerGrp1 := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   s.testOrg.ID,
					OwnerType: "org",
					Type:      app.RunnerGroupTypeOrg,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err := s.service.DB.WithContext(ctx).Create(runnerGrp1).Error
				require.NoError(s.T(), err)

				settings1 := &app.RunnerGroupSettings{
					RunnerGroupID:     runnerGrp1.ID,
					OrgID:             s.testOrg.ID,
					ContainerImageURL: "gcr.io/nuon-dev-public/runner",
					ContainerImageTag: "v1.0.0",
					RunnerAPIURL:      "https://api.example.com",
					Metadata:          pgtype.Hstore{},
				}
				err = s.service.DB.WithContext(ctx).Create(settings1).Error
				require.NoError(s.T(), err)

				runnerGrp2 := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   s.testOrg.ID,
					OwnerType: "org",
					Type:      app.RunnerGroupTypeOrg,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err = s.service.DB.WithContext(ctx).Create(runnerGrp2).Error
				require.NoError(s.T(), err)

				settings2 := &app.RunnerGroupSettings{
					RunnerGroupID:     runnerGrp2.ID,
					OrgID:             s.testOrg.ID,
					ContainerImageURL: "gcr.io/nuon-dev-public/runner",
					ContainerImageTag: "v1.0.0",
					RunnerAPIURL:      "https://api.example.com",
					Metadata:          pgtype.Hstore{},
				}
				err = s.service.DB.WithContext(ctx).Create(settings2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(settings1)
					s.service.DB.Unscoped().Delete(settings2)
					s.service.DB.Unscoped().Delete(runnerGrp1)
					s.service.DB.Unscoped().Delete(runnerGrp2)
				})

				return []string{runnerGrp1.ID, runnerGrp2.ID}
			},
			requestBody: AdminBulkUpdateRunnersRequest{
				ContainerImageTag: "v2.0.0",
			},
			expectedCode: http.StatusOK,
			validateFunc: nil, // Bulk update operates on all orgs with OrgTypeDefault, which may not include test orgs
		},
		{
			name: "missing request body returns 400",
			setupFunc: func() []string {
				return []string{}
			},
			requestBody:  nil,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerGrpIDs := tc.setupFunc()
			rr := s.makeRequest("PATCH", "/v1/runners/bulk-update", tc.requestBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc(runnerGrpIDs)
			}
		})
	}
}
