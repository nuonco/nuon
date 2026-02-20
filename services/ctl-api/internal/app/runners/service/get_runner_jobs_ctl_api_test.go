package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

type GetRunnerJobsCtlAPITestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetRunnerJobsCtlAPITestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       GetRunnerJobsCtlAPITestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	testLogStream *app.LogStream
}

func TestGetRunnerJobsCtlAPISuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetRunnerJobsCtlAPITestSuite))
}

func (s *GetRunnerJobsCtlAPITestSuite) SetupSuite() {
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

func (s *GetRunnerJobsCtlAPITestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with public routes
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetRunnerJobsCtlAPITestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetRunnerJobsCtlAPITestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create log stream for runner jobs
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

func (s *GetRunnerJobsCtlAPITestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetRunnerJobsCtlAPITestSuite) TestGetRunnerJobsCtlAPI() {
	testCases := []struct {
		name             string
		setupFunc        func() []string
		queryParams      string
		expectedCode     int
		expectedCount    int
		validateFunc     func([]*app.RunnerJob)
		expectedNotFound bool
	}{
		{
			name: "successfully get all runner jobs",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				jobIDs := []string{}
				for i := 0; i < 3; i++ {
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
						QueueTimeout:      5 * time.Minute,
						AvailableTimeout:  10 * time.Minute,
						ExecutionTimeout:  30 * time.Minute,
						MaxExecutions:     3,
					}
					err := s.service.DB.WithContext(ctx).Create(job).Error
					require.NoError(s.T(), err)
					jobIDs = append(jobIDs, job.ID)
				}

				s.T().Cleanup(func() {
					for _, jobID := range jobIDs {
						s.service.DB.Unscoped().Delete(&app.RunnerJob{ID: jobID})
					}
				})

				return jobIDs
			},
			queryParams:   "",
			expectedCode:  http.StatusOK,
			expectedCount: 3,
			validateFunc: func(jobs []*app.RunnerJob) {
				for _, job := range jobs {
					assert.Equal(s.T(), s.testRunner.ID, job.RunnerID)
					assert.Equal(s.T(), s.testOrg.ID, job.OrgID)
				}
			},
		},
		{
			name: "filter by single status",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				jobIDs := []string{}
				statuses := []app.RunnerJobStatus{
					app.RunnerJobStatusAvailable,
					app.RunnerJobStatusInProgress,
					app.RunnerJobStatusFinished,
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
						QueueTimeout:      5 * time.Minute,
						AvailableTimeout:  10 * time.Minute,
						ExecutionTimeout:  30 * time.Minute,
						MaxExecutions:     3,
					}
					err := s.service.DB.WithContext(ctx).Create(job).Error
					require.NoError(s.T(), err)
					jobIDs = append(jobIDs, job.ID)
				}

				s.T().Cleanup(func() {
					for _, jobID := range jobIDs {
						s.service.DB.Unscoped().Delete(&app.RunnerJob{ID: jobID})
					}
				})

				return jobIDs
			},
			queryParams:   "?status=in-progress",
			expectedCode:  http.StatusOK,
			expectedCount: 1,
			validateFunc: func(jobs []*app.RunnerJob) {
				require.Len(s.T(), jobs, 1)
				assert.Equal(s.T(), app.RunnerJobStatusInProgress, jobs[0].Status)
			},
		},
		{
			name: "filter by multiple statuses",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				jobIDs := []string{}
				statuses := []app.RunnerJobStatus{
					app.RunnerJobStatusAvailable,
					app.RunnerJobStatusInProgress,
					app.RunnerJobStatusFinished,
					app.RunnerJobStatusFailed,
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
						QueueTimeout:      5 * time.Minute,
						AvailableTimeout:  10 * time.Minute,
						ExecutionTimeout:  30 * time.Minute,
						MaxExecutions:     3,
					}
					err := s.service.DB.WithContext(ctx).Create(job).Error
					require.NoError(s.T(), err)
					jobIDs = append(jobIDs, job.ID)
				}

				s.T().Cleanup(func() {
					for _, jobID := range jobIDs {
						s.service.DB.Unscoped().Delete(&app.RunnerJob{ID: jobID})
					}
				})

				return jobIDs
			},
			queryParams:   "?statuses=available,in-progress",
			expectedCode:  http.StatusOK,
			expectedCount: 2,
			validateFunc: func(jobs []*app.RunnerJob) {
				require.Len(s.T(), jobs, 2)
				statuses := []app.RunnerJobStatus{jobs[0].Status, jobs[1].Status}
				assert.Contains(s.T(), statuses, app.RunnerJobStatusAvailable)
				assert.Contains(s.T(), statuses, app.RunnerJobStatusInProgress)
			},
		},
		{
			name: "filter by single group",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				jobIDs := []string{}
				groups := []app.RunnerJobGroup{
					app.RunnerJobGroupBuild,
					app.RunnerJobGroupDeploy,
					app.RunnerJobGroupSync,
				}

				for _, group := range groups {
					job := &app.RunnerJob{
						ID:                domains.NewRunnerJobID(),
						OrgID:             s.testOrg.ID,
						RunnerID:          s.testRunner.ID,
						LogStreamID:       generics.ToPtr(s.testLogStream.ID),
						Status:            app.RunnerJobStatusAvailable,
						StatusDescription: "test job",
						Group:             group,
						Type:              app.RunnerJobTypeDockerBuild,
						Operation:         app.RunnerJobOperationTypeBuild,
						QueueTimeout:      5 * time.Minute,
						AvailableTimeout:  10 * time.Minute,
						ExecutionTimeout:  30 * time.Minute,
						MaxExecutions:     3,
					}
					err := s.service.DB.WithContext(ctx).Create(job).Error
					require.NoError(s.T(), err)
					jobIDs = append(jobIDs, job.ID)
				}

				s.T().Cleanup(func() {
					for _, jobID := range jobIDs {
						s.service.DB.Unscoped().Delete(&app.RunnerJob{ID: jobID})
					}
				})

				return jobIDs
			},
			queryParams:   "?group=deploy",
			expectedCode:  http.StatusOK,
			expectedCount: 1,
			validateFunc: func(jobs []*app.RunnerJob) {
				require.Len(s.T(), jobs, 1)
				assert.Equal(s.T(), app.RunnerJobGroupDeploy, jobs[0].Group)
			},
		},
		{
			name: "filter by multiple groups",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				jobIDs := []string{}
				groups := []app.RunnerJobGroup{
					app.RunnerJobGroupBuild,
					app.RunnerJobGroupDeploy,
					app.RunnerJobGroupSync,
					app.RunnerJobGroupActions,
				}

				for _, group := range groups {
					job := &app.RunnerJob{
						ID:                domains.NewRunnerJobID(),
						OrgID:             s.testOrg.ID,
						RunnerID:          s.testRunner.ID,
						LogStreamID:       generics.ToPtr(s.testLogStream.ID),
						Status:            app.RunnerJobStatusAvailable,
						StatusDescription: "test job",
						Group:             group,
						Type:              app.RunnerJobTypeDockerBuild,
						Operation:         app.RunnerJobOperationTypeBuild,
						QueueTimeout:      5 * time.Minute,
						AvailableTimeout:  10 * time.Minute,
						ExecutionTimeout:  30 * time.Minute,
						MaxExecutions:     3,
					}
					err := s.service.DB.WithContext(ctx).Create(job).Error
					require.NoError(s.T(), err)
					jobIDs = append(jobIDs, job.ID)
				}

				s.T().Cleanup(func() {
					for _, jobID := range jobIDs {
						s.service.DB.Unscoped().Delete(&app.RunnerJob{ID: jobID})
					}
				})

				return jobIDs
			},
			queryParams:   "?groups=build,deploy",
			expectedCode:  http.StatusOK,
			expectedCount: 2,
			validateFunc: func(jobs []*app.RunnerJob) {
				require.Len(s.T(), jobs, 2)
				groups := []app.RunnerJobGroup{jobs[0].Group, jobs[1].Group}
				assert.Contains(s.T(), groups, app.RunnerJobGroupBuild)
				assert.Contains(s.T(), groups, app.RunnerJobGroupDeploy)
			},
		},
		{
			name: "filter by both status and group",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				jobIDs := []string{}
				testCases := []struct {
					status app.RunnerJobStatus
					group  app.RunnerJobGroup
				}{
					{app.RunnerJobStatusAvailable, app.RunnerJobGroupBuild},
					{app.RunnerJobStatusInProgress, app.RunnerJobGroupBuild},
					{app.RunnerJobStatusAvailable, app.RunnerJobGroupDeploy},
					{app.RunnerJobStatusFinished, app.RunnerJobGroupBuild},
				}

				for _, tc := range testCases {
					job := &app.RunnerJob{
						ID:                domains.NewRunnerJobID(),
						OrgID:             s.testOrg.ID,
						RunnerID:          s.testRunner.ID,
						LogStreamID:       generics.ToPtr(s.testLogStream.ID),
						Status:            tc.status,
						StatusDescription: "test job",
						Group:             tc.group,
						Type:              app.RunnerJobTypeDockerBuild,
						Operation:         app.RunnerJobOperationTypeBuild,
						QueueTimeout:      5 * time.Minute,
						AvailableTimeout:  10 * time.Minute,
						ExecutionTimeout:  30 * time.Minute,
						MaxExecutions:     3,
					}
					err := s.service.DB.WithContext(ctx).Create(job).Error
					require.NoError(s.T(), err)
					jobIDs = append(jobIDs, job.ID)
				}

				s.T().Cleanup(func() {
					for _, jobID := range jobIDs {
						s.service.DB.Unscoped().Delete(&app.RunnerJob{ID: jobID})
					}
				})

				return jobIDs
			},
			queryParams:   "?status=available&group=build",
			expectedCode:  http.StatusOK,
			expectedCount: 1,
			validateFunc: func(jobs []*app.RunnerJob) {
				require.Len(s.T(), jobs, 1)
				assert.Equal(s.T(), app.RunnerJobStatusAvailable, jobs[0].Status)
				assert.Equal(s.T(), app.RunnerJobGroupBuild, jobs[0].Group)
			},
		},
		{
			name: "limit results",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				jobIDs := []string{}
				for i := 0; i < 10; i++ {
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
						QueueTimeout:      5 * time.Minute,
						AvailableTimeout:  10 * time.Minute,
						ExecutionTimeout:  30 * time.Minute,
						MaxExecutions:     3,
					}
					err := s.service.DB.WithContext(ctx).Create(job).Error
					require.NoError(s.T(), err)
					jobIDs = append(jobIDs, job.ID)
				}

				s.T().Cleanup(func() {
					for _, jobID := range jobIDs {
						s.service.DB.Unscoped().Delete(&app.RunnerJob{ID: jobID})
					}
				})

				return jobIDs
			},
			queryParams:   "?limit=5",
			expectedCode:  http.StatusOK,
			expectedCount: 5,
		},
		{
			name: "runner not found",
			setupFunc: func() []string {
				return []string{}
			},
			queryParams:      "",
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "no jobs for runner",
			setupFunc: func() []string {
				// Don't create any jobs
				return []string{}
			},
			queryParams:   "",
			expectedCode:  http.StatusOK,
			expectedCount: 0,
		},
		{
			name: "jobs ordered by created_at desc",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				jobIDs := []string{}
				for i := 0; i < 3; i++ {
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
						QueueTimeout:      5 * time.Minute,
						AvailableTimeout:  10 * time.Minute,
						ExecutionTimeout:  30 * time.Minute,
						MaxExecutions:     3,
					}
					err := s.service.DB.WithContext(ctx).Create(job).Error
					require.NoError(s.T(), err)
					jobIDs = append(jobIDs, job.ID)
				}

				s.T().Cleanup(func() {
					for _, jobID := range jobIDs {
						s.service.DB.Unscoped().Delete(&app.RunnerJob{ID: jobID})
					}
				})

				return jobIDs
			},
			queryParams:   "",
			expectedCode:  http.StatusOK,
			expectedCount: 3,
			validateFunc: func(jobs []*app.RunnerJob) {
				// Verify jobs are ordered by created_at desc (newest first)
				for i := 0; i < len(jobs)-1; i++ {
					assert.True(s.T(), jobs[i].CreatedAt.After(jobs[i+1].CreatedAt) ||
						jobs[i].CreatedAt.Equal(jobs[i+1].CreatedAt),
						"jobs should be ordered by created_at desc")
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_ = tc.setupFunc()

			runnerID := s.testRunner.ID
			if tc.expectedNotFound {
				runnerID = "rnrnonexistent123456789012"
			}

			rr := s.makeRequest("GET", "/v1/runners/"+runnerID+"/jobs"+tc.queryParams)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else {
				var jobs []*app.RunnerJob
				err := json.Unmarshal(rr.Body.Bytes(), &jobs)
				require.NoError(s.T(), err)

				if tc.expectedCount >= 0 {
					assert.Len(s.T(), jobs, tc.expectedCount)
				}

				if tc.validateFunc != nil {
					tc.validateFunc(jobs)
				}
			}
		})
	}
}

func (s *GetRunnerJobsCtlAPITestSuite) TestGetRunnerJobsCrossOrgIsolation() {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	// Create second org
	org2 := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "other-org",
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/bar",
		},
	}
	err := s.service.DB.WithContext(ctx).Create(org2).Error
	require.NoError(s.T(), err)

	// Create log stream for org2
	logStream2 := &app.LogStream{
		ID:      domains.NewLogStreamID(),
		OrgID:   org2.ID,
		OwnerID: org2.ID,
		Open:    true,
	}
	err = s.service.DB.WithContext(ctx).Create(logStream2).Error
	require.NoError(s.T(), err)

	// Create runner group for org2
	runnerGrp2 := &app.RunnerGroup{
		ID:        domains.NewRunnerGroupID(),
		OrgID:     org2.ID,
		OwnerID:   org2.ID,
		OwnerType: "org",
		Type:      app.RunnerGroupTypeOrg,
		Platform:  app.AppRunnerTypeAWSEKS,
	}
	err = s.service.DB.WithContext(ctx).Create(runnerGrp2).Error
	require.NoError(s.T(), err)

	// Create runner in org2
	runner2 := &app.Runner{
		ID:            domains.NewRunnerID(),
		OrgID:         org2.ID,
		Name:          "runner-org2",
		DisplayName:   "Runner in Org 2",
		Status:        app.RunnerStatusActive,
		RunnerGroupID: runnerGrp2.ID,
	}
	err = s.service.DB.WithContext(ctx).Create(runner2).Error
	require.NoError(s.T(), err)

	// Create job in org2
	job := &app.RunnerJob{
		ID:                domains.NewRunnerJobID(),
		OrgID:             org2.ID,
		RunnerID:          runner2.ID,
		LogStreamID:       generics.ToPtr(logStream2.ID),
		Status:            app.RunnerJobStatusAvailable,
		StatusDescription: "test job",
		Group:             app.RunnerJobGroupBuild,
		Type:              app.RunnerJobTypeDockerBuild,
		Operation:         app.RunnerJobOperationTypeBuild,
		QueueTimeout:      5 * time.Minute,
		AvailableTimeout:  10 * time.Minute,
		ExecutionTimeout:  30 * time.Minute,
		MaxExecutions:     3,
	}
	err = s.service.DB.WithContext(ctx).Create(job).Error
	require.NoError(s.T(), err)

	// Attempt to get jobs from org2's runner with org1 context
	rr := s.makeRequest("GET", "/v1/runners/"+runner2.ID+"/jobs")
	require.Equal(s.T(), http.StatusNotFound, rr.Code, "should not access runner from different org")

	// Cleanup
	s.service.DB.Unscoped().Delete(job)
	s.service.DB.Unscoped().Delete(logStream2)
	s.service.DB.Unscoped().Delete(runner2)
	s.service.DB.Unscoped().Delete(runnerGrp2)
	s.service.DB.Unscoped().Delete(org2)
}
