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

type GetRunnerJobPublicTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetRunnerJobPublicTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       GetRunnerJobPublicTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	testLogStream *app.LogStream
}

func TestGetRunnerJobPublicSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetRunnerJobPublicTestSuite))
}

func (s *GetRunnerJobPublicTestSuite) SetupSuite() {
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

func (s *GetRunnerJobPublicTestSuite) SetupTest() {
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

func (s *GetRunnerJobPublicTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetRunnerJobPublicTestSuite) setupTestData() {
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

func (s *GetRunnerJobPublicTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetRunnerJobPublicTestSuite) TestGetRunnerJobPublic() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(*app.RunnerJob)
		expectedNotFound bool
	}{
		{
			name: "successfully get runner job",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

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

				// Create execution for the job
				exec := &app.RunnerJobExecution{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					Status:      app.RunnerJobExecutionStatusPending,
				}
				err = s.service.DB.WithContext(ctx).Create(exec).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(exec)
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(job *app.RunnerJob) {
				assert.Equal(s.T(), s.testOrg.ID, job.OrgID)
				assert.Equal(s.T(), s.testRunner.ID, job.RunnerID)
				assert.Equal(s.T(), app.RunnerJobStatusAvailable, job.Status)
				assert.Equal(s.T(), app.RunnerJobGroupBuild, job.Group)
				assert.Equal(s.T(), app.RunnerJobTypeDockerBuild, job.Type)
				// Should preload latest execution
				assert.Len(s.T(), job.Executions, 1)
			},
		},
		{
			name: "runner job not found",
			setupFunc: func() string {
				return "rjbnonexistent123456789012"
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "runner job in different org not accessible",
			setupFunc: func() string {
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

				// Create log stream for org2
				logStream2 := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   org2.ID,
					OwnerID: org2.ID,
					Open:    true,
				}
				err = s.service.DB.WithContext(ctx).Create(logStream2).Error
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

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(job)
					s.service.DB.Unscoped().Delete(logStream2)
					s.service.DB.Unscoped().Delete(runner2)
					s.service.DB.Unscoped().Delete(runnerGrp2)
					s.service.DB.Unscoped().Delete(org2)
				})

				return job.ID
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "job with multiple executions returns latest",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				job := &app.RunnerJob{
					ID:                domains.NewRunnerJobID(),
					OrgID:             s.testOrg.ID,
					RunnerID:          s.testRunner.ID,
					LogStreamID:       generics.ToPtr(s.testLogStream.ID),
					Status:            app.RunnerJobStatusInProgress,
					StatusDescription: "test job",
					Group:             app.RunnerJobGroupDeploy,
					Type:              app.RunnerJobTypeTerraformDeploy,
					Operation:         app.RunnerJobOperationTypeCreateApplyPlan,
					QueueTimeout:      5 * time.Minute,
					AvailableTimeout:  10 * time.Minute,
					ExecutionTimeout:  30 * time.Minute,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				// Create older execution
				exec1 := &app.RunnerJobExecution{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					Status:      app.RunnerJobExecutionStatusFinished,
				}
				err = s.service.DB.WithContext(ctx).Create(exec1).Error
				require.NoError(s.T(), err)

				// Create newer execution (should be the one returned)
				exec2 := &app.RunnerJobExecution{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					Status:      app.RunnerJobExecutionStatusInProgress,
				}
				err = s.service.DB.WithContext(ctx).Create(exec2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(exec2)
					s.service.DB.Unscoped().Delete(exec1)
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(job *app.RunnerJob) {
				require.Len(s.T(), job.Executions, 1, "should only return latest execution")
				assert.Equal(s.T(), app.RunnerJobExecutionStatusInProgress, job.Executions[0].Status)
			},
		},
		{
			name: "job with no executions",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				job := &app.RunnerJob{
					ID:                domains.NewRunnerJobID(),
					OrgID:             s.testOrg.ID,
					RunnerID:          s.testRunner.ID,
					LogStreamID:       generics.ToPtr(s.testLogStream.ID),
					Status:            app.RunnerJobStatusQueued,
					StatusDescription: "test job",
					Group:             app.RunnerJobGroupHealthChecks,
					Type:              app.RunnerJobTypeHealthCheck,
					Operation:         app.RunnerJobOperationTypeExec,
					QueueTimeout:      5 * time.Minute,
					AvailableTimeout:  10 * time.Minute,
					ExecutionTimeout:  30 * time.Minute,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(job *app.RunnerJob) {
				assert.Len(s.T(), job.Executions, 0, "should have no executions")
				assert.Equal(s.T(), app.RunnerJobStatusQueued, job.Status)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			jobID := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/runner-jobs/"+jobID)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else if tc.validateFunc != nil {
				var job app.RunnerJob
				err := json.Unmarshal(rr.Body.Bytes(), &job)
				require.NoError(s.T(), err)
				tc.validateFunc(&job)
			}
		})
	}
}

func (s *GetRunnerJobPublicTestSuite) TestGetRunnerJobWithDifferentStatuses() {
	statuses := []app.RunnerJobStatus{
		app.RunnerJobStatusQueued,
		app.RunnerJobStatusAvailable,
		app.RunnerJobStatusInProgress,
		app.RunnerJobStatusFinished,
		app.RunnerJobStatusFailed,
		app.RunnerJobStatusTimedOut,
		app.RunnerJobStatusCancelled,
	}

	for _, status := range statuses {
		s.Run(string(status), func() {
			ctx := context.Background()
			ctx = cctx.SetAccountContext(ctx, s.testAcc)

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

			rr := s.makeRequest("GET", "/v1/runner-jobs/"+job.ID)
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var result app.RunnerJob
			err = json.Unmarshal(rr.Body.Bytes(), &result)
			require.NoError(s.T(), err)
			assert.Equal(s.T(), status, result.Status)

			s.service.DB.Unscoped().Delete(job)
		})
	}
}
