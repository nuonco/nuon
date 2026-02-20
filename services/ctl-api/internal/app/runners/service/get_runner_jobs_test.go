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

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type GetRunnerJobsTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetRunnerJobsTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       GetRunnerJobsTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	testLogStream *app.LogStream
}

func TestGetRunnerJobsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetRunnerJobsTestSuite))
}

func (s *GetRunnerJobsTestSuite) SetupSuite() {
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

func (s *GetRunnerJobsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with runner routes (no TestOrg/TestAcc needed)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.RunnersService.RegisterRunnerRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetRunnerJobsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetRunnerJobsTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create log stream
	s.testLogStream = &app.LogStream{
		ID:      domains.NewLogStreamID(),
		OrgID:   s.testOrg.ID,
		OwnerID: s.testOrg.ID,
		Open:    true,
	}
	err := s.service.DB.WithContext(ctx).Create(s.testLogStream).Error
	require.NoError(s.T(), err)

	// Create runner group
	s.testRunnerGrp = &app.RunnerGroup{
		ID:        domains.NewRunnerGroupID(),
		OrgID:     s.testOrg.ID,
		OwnerID:   s.testOrg.ID,
		OwnerType: "org",
		Type:      app.RunnerGroupTypeOrg,
		Platform:  app.AppRunnerTypeAWSEKS,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunnerGrp).Error
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

func (s *GetRunnerJobsTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetRunnerJobsTestSuite) TestGetRunnerJobs() {
	testCases := []struct {
		name          string
		setupFunc     func() string
		queryParams   string
		expectedCode  int
		expectedCount int
		validateFunc  func([]*app.RunnerJob)
	}{
		{
			name: "returns jobs with default status=available",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create available job
				job1 := &app.RunnerJob{
					ID:                domains.NewRunnerJobID(),
					OrgID:             s.testOrg.ID,
					RunnerID:          s.testRunner.ID,
					LogStreamID:       generics.ToPtr(s.testLogStream.ID),
					Status:            app.RunnerJobStatusAvailable,
					StatusDescription: "test job 1",
					Group:             app.RunnerJobGroupBuild,
					Type:              app.RunnerJobTypeDockerBuild,
					Operation:         app.RunnerJobOperationTypeBuild,
					QueueTimeout:      60,
					AvailableTimeout:  60,
					ExecutionTimeout:  300,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job1).Error
				require.NoError(s.T(), err)

				// Create queued job (should not be returned with default filter)
				job2 := &app.RunnerJob{
					ID:                domains.NewRunnerJobID(),
					OrgID:             s.testOrg.ID,
					RunnerID:          s.testRunner.ID,
					LogStreamID:       generics.ToPtr(s.testLogStream.ID),
					Status:            app.RunnerJobStatusQueued,
					StatusDescription: "test job 2",
					Group:             app.RunnerJobGroupBuild,
					Type:              app.RunnerJobTypeDockerBuild,
					Operation:         app.RunnerJobOperationTypeBuild,
					QueueTimeout:      60,
					AvailableTimeout:  60,
					ExecutionTimeout:  300,
					MaxExecutions:     3,
				}
				err = s.service.DB.WithContext(ctx).Create(job2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(job1)
					s.service.DB.Unscoped().Delete(job2)
				})

				return s.testRunner.ID
			},
			expectedCode:  http.StatusOK,
			expectedCount: 1,
			validateFunc: func(jobs []*app.RunnerJob) {
				assert.Len(s.T(), jobs, 1)
				assert.Equal(s.T(), app.RunnerJobStatusAvailable, jobs[0].Status)
			},
		},
		{
			name: "returns empty array when no matching jobs",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			expectedCode:  http.StatusOK,
			expectedCount: 0,
			validateFunc: func(jobs []*app.RunnerJob) {
				assert.Empty(s.T(), jobs)
			},
		},
		{
			name: "filters by group query param",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create build job
				job1 := &app.RunnerJob{
					ID:                domains.NewRunnerJobID(),
					OrgID:             s.testOrg.ID,
					RunnerID:          s.testRunner.ID,
					LogStreamID:       generics.ToPtr(s.testLogStream.ID),
					Status:            app.RunnerJobStatusAvailable,
					StatusDescription: "build job",
					Group:             app.RunnerJobGroupBuild,
					Type:              app.RunnerJobTypeDockerBuild,
					Operation:         app.RunnerJobOperationTypeBuild,
					QueueTimeout:      60,
					AvailableTimeout:  60,
					ExecutionTimeout:  300,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job1).Error
				require.NoError(s.T(), err)

				// Create deploy job
				job2 := &app.RunnerJob{
					ID:                domains.NewRunnerJobID(),
					OrgID:             s.testOrg.ID,
					RunnerID:          s.testRunner.ID,
					LogStreamID:       generics.ToPtr(s.testLogStream.ID),
					Status:            app.RunnerJobStatusAvailable,
					StatusDescription: "deploy job",
					Group:             app.RunnerJobGroupDeploy,
					Type:              app.RunnerJobTypeTerraformDeploy,
					Operation:         app.RunnerJobOperationTypeCreateApplyPlan,
					QueueTimeout:      60,
					AvailableTimeout:  60,
					ExecutionTimeout:  300,
					MaxExecutions:     3,
				}
				err = s.service.DB.WithContext(ctx).Create(job2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(job1)
					s.service.DB.Unscoped().Delete(job2)
				})

				return s.testRunner.ID
			},
			queryParams:   "?group=build",
			expectedCode:  http.StatusOK,
			expectedCount: 1,
			validateFunc: func(jobs []*app.RunnerJob) {
				assert.Len(s.T(), jobs, 1)
				assert.Equal(s.T(), app.RunnerJobGroupBuild, jobs[0].Group)
			},
		},
		{
			name: "filters by status query param",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create in-progress job
				job1 := &app.RunnerJob{
					ID:                domains.NewRunnerJobID(),
					OrgID:             s.testOrg.ID,
					RunnerID:          s.testRunner.ID,
					LogStreamID:       generics.ToPtr(s.testLogStream.ID),
					Status:            app.RunnerJobStatusInProgress,
					StatusDescription: "in progress job",
					Group:             app.RunnerJobGroupBuild,
					Type:              app.RunnerJobTypeDockerBuild,
					Operation:         app.RunnerJobOperationTypeBuild,
					QueueTimeout:      60,
					AvailableTimeout:  60,
					ExecutionTimeout:  300,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job1).Error
				require.NoError(s.T(), err)

				// Create available job
				job2 := &app.RunnerJob{
					ID:                domains.NewRunnerJobID(),
					OrgID:             s.testOrg.ID,
					RunnerID:          s.testRunner.ID,
					LogStreamID:       generics.ToPtr(s.testLogStream.ID),
					Status:            app.RunnerJobStatusAvailable,
					StatusDescription: "available job",
					Group:             app.RunnerJobGroupBuild,
					Type:              app.RunnerJobTypeDockerBuild,
					Operation:         app.RunnerJobOperationTypeBuild,
					QueueTimeout:      60,
					AvailableTimeout:  60,
					ExecutionTimeout:  300,
					MaxExecutions:     3,
				}
				err = s.service.DB.WithContext(ctx).Create(job2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(job1)
					s.service.DB.Unscoped().Delete(job2)
				})

				return s.testRunner.ID
			},
			queryParams:   "?status=in-progress",
			expectedCode:  http.StatusOK,
			expectedCount: 1,
			validateFunc: func(jobs []*app.RunnerJob) {
				assert.Len(s.T(), jobs, 1)
				assert.Equal(s.T(), app.RunnerJobStatusInProgress, jobs[0].Status)
			},
		},
		{
			name: "runner not found still returns empty array",
			setupFunc: func() string {
				return "runnonexistent123456789012"
			},
			expectedCode:  http.StatusOK,
			expectedCount: 0,
			validateFunc: func(jobs []*app.RunnerJob) {
				assert.Empty(s.T(), jobs)
			},
		},
		{
			name: "invalid limit returns 400",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			queryParams:  "?limit=invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "status=unknown shows all statuses",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create jobs with different statuses
				statuses := []app.RunnerJobStatus{
					app.RunnerJobStatusQueued,
					app.RunnerJobStatusAvailable,
					app.RunnerJobStatusInProgress,
				}

				for _, status := range statuses {
					job := &app.RunnerJob{
						ID:                domains.NewRunnerJobID(),
						OrgID:             s.testOrg.ID,
						RunnerID:          s.testRunner.ID,
						LogStreamID:       generics.ToPtr(s.testLogStream.ID),
						Status:            status,
						StatusDescription: "test job",
						Group:             app.RunnerJobGroupBuild,
						Type:              app.RunnerJobTypeDockerBuild,
						Operation:         app.RunnerJobOperationTypeBuild,
						QueueTimeout:      60,
						AvailableTimeout:  60,
						ExecutionTimeout:  300,
						MaxExecutions:     3,
					}
					err := s.service.DB.WithContext(ctx).Create(job).Error
					require.NoError(s.T(), err)

					jobID := job.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Where("id = ?", jobID).Delete(&app.RunnerJob{})
					})
				}

				return s.testRunner.ID
			},
			queryParams:   "?status=",
			expectedCode:  http.StatusOK,
			expectedCount: 3,
			validateFunc: func(jobs []*app.RunnerJob) {
				assert.Len(s.T(), jobs, 3)
			},
		},
		{
			name: "group=any shows all groups",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create jobs with different groups
				groups := []app.RunnerJobGroup{
					app.RunnerJobGroupBuild,
					app.RunnerJobGroupDeploy,
					app.RunnerJobGroupSync,
				}

				for _, grp := range groups {
					job := &app.RunnerJob{
						ID:                domains.NewRunnerJobID(),
						OrgID:             s.testOrg.ID,
						RunnerID:          s.testRunner.ID,
						LogStreamID:       generics.ToPtr(s.testLogStream.ID),
						Status:            app.RunnerJobStatusAvailable,
						StatusDescription: "test job",
						Group:             grp,
						Type:              app.RunnerJobTypeDockerBuild,
						Operation:         app.RunnerJobOperationTypeBuild,
						QueueTimeout:      60,
						AvailableTimeout:  60,
						ExecutionTimeout:  300,
						MaxExecutions:     3,
					}
					err := s.service.DB.WithContext(ctx).Create(job).Error
					require.NoError(s.T(), err)

					jobID := job.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Where("id = ?", jobID).Delete(&app.RunnerJob{})
					})
				}

				return s.testRunner.ID
			},
			queryParams:   "?group=any",
			expectedCode:  http.StatusOK,
			expectedCount: 3,
			validateFunc: func(jobs []*app.RunnerJob) {
				assert.Len(s.T(), jobs, 3)
			},
		},
		{
			name: "respects limit parameter",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create 5 jobs
				for i := 0; i < 5; i++ {
					job := &app.RunnerJob{
						ID:                domains.NewRunnerJobID(),
						OrgID:             s.testOrg.ID,
						RunnerID:          s.testRunner.ID,
						LogStreamID:       generics.ToPtr(s.testLogStream.ID),
						Status:            app.RunnerJobStatusAvailable,
						StatusDescription: "test job",
						Group:             app.RunnerJobGroupBuild,
						Type:              app.RunnerJobTypeDockerBuild,
						Operation:         app.RunnerJobOperationTypeBuild,
						QueueTimeout:      60,
						AvailableTimeout:  60,
						ExecutionTimeout:  300,
						MaxExecutions:     3,
					}
					err := s.service.DB.WithContext(ctx).Create(job).Error
					require.NoError(s.T(), err)

					jobID := job.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Where("id = ?", jobID).Delete(&app.RunnerJob{})
					})
				}

				return s.testRunner.ID
			},
			queryParams:   "?limit=3",
			expectedCode:  http.StatusOK,
			expectedCount: 3,
			validateFunc: func(jobs []*app.RunnerJob) {
				assert.Len(s.T(), jobs, 3, "returns limit items")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID := tc.setupFunc()
			path := "/v1/runners/" + runnerID + "/jobs"
			if tc.queryParams != "" {
				path += tc.queryParams
			}
			rr := s.makeRequest("GET", path)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var jobs []*app.RunnerJob
				err := json.Unmarshal(rr.Body.Bytes(), &jobs)
				require.NoError(s.T(), err)
				tc.validateFunc(jobs)
			}
		})
	}
}
