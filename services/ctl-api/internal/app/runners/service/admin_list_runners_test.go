package service

import (
	"context"
	"encoding/json"
	"fmt"
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

type AdminListRunnersTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminListRunnersTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       AdminListRunnersTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunnerGrp *app.RunnerGroup
}

func TestAdminListRunnersSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminListRunnersTestSuite))
}

func (s *AdminListRunnersTestSuite) SetupSuite() {
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

func (s *AdminListRunnersTestSuite) SetupTest() {
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

func (s *AdminListRunnersTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminListRunnersTestSuite) setupTestData() {
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
}

func (s *AdminListRunnersTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminListRunnersTestSuite) TestAdminListRunners() {
	testCases := []struct {
		name          string
		setupFunc     func() []string
		queryParams   string
		expectedCount int
		expectedCode  int
		validateFunc  func([]*app.Runner)
	}{
		{
			name: "empty results when no runners exist",
			setupFunc: func() []string {
				return []string{}
			},
			queryParams:   "",
			expectedCount: 0,
			expectedCode:  http.StatusOK,
		},
		{
			name: "list all org runners",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner1 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "runner-1",
					DisplayName:   "Runner 1",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err := s.service.DB.WithContext(ctx).Create(runner1).Error
				require.NoError(s.T(), err)

				runner2 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "runner-2",
					DisplayName:   "Runner 2",
					Status:        app.RunnerStatusPending,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err = s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner1)
					s.service.DB.Unscoped().Delete(runner2)
				})

				return []string{runner1.ID, runner2.ID}
			},
			queryParams:   "?type=org",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(runners []*app.Runner) {
				assert.Len(s.T(), runners, 2)
				// Verify CreatedByID is set
				for _, runner := range runners {
					assert.NotEmpty(s.T(), runner.CreatedByID)
				}
			},
		},
		{
			name: "filter by type - default to org",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "org-runner",
					DisplayName:   "Org Runner",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return []string{runner.ID}
			},
			queryParams:   "", // No type param, should default to "org"
			expectedCount: 1,
			expectedCode:  http.StatusOK,
		},
		{
			name: "pagination with limit",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				var runnerIDs []string
				for i := 0; i < 5; i++ {
					runner := &app.Runner{
						ID:            domains.NewRunnerID(),
						OrgID:         s.testOrg.ID,
						Name:          fmt.Sprintf("runner-%d", i),
						DisplayName:   "Runner",
						Status:        app.RunnerStatusActive,
						RunnerGroupID: s.testRunnerGrp.ID,
					}
					err := s.service.DB.WithContext(ctx).Create(runner).Error
					require.NoError(s.T(), err)
					runnerIDs = append(runnerIDs, runner.ID)

					runnerID := runner.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Where("id = ?", runnerID).Delete(&app.Runner{})
					})
				}

				return runnerIDs
			},
			queryParams:   "?limit=3",
			expectedCount: 3,
			expectedCode:  http.StatusOK,
		},
		{
			name: "pagination with offset",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				var runnerIDs []string
				for i := 0; i < 5; i++ {
					runner := &app.Runner{
						ID:            domains.NewRunnerID(),
						OrgID:         s.testOrg.ID,
						Name:          fmt.Sprintf("runner-offset-%d", i),
						DisplayName:   "Runner",
						Status:        app.RunnerStatusActive,
						RunnerGroupID: s.testRunnerGrp.ID,
					}
					err := s.service.DB.WithContext(ctx).Create(runner).Error
					require.NoError(s.T(), err)
					runnerIDs = append(runnerIDs, runner.ID)

					runnerID := runner.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Where("id = ?", runnerID).Delete(&app.Runner{})
					})
				}

				return runnerIDs
			},
			queryParams:   "?offset=2&limit=10",
			expectedCount: 3,
			expectedCode:  http.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_ = tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/runners"+tc.queryParams)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			var runners []*app.Runner
			err := json.Unmarshal(rr.Body.Bytes(), &runners)
			require.NoError(s.T(), err)

			assert.Len(s.T(), runners, tc.expectedCount)

			if tc.validateFunc != nil {
				tc.validateFunc(runners)
			}
		})
	}
}
