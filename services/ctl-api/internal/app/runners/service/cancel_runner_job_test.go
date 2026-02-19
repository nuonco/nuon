package service

import (
	"bytes"
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

type CancelRunnerJobTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type CancelRunnerJobTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       CancelRunnerJobTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	testLogStream *app.LogStream
}

func TestCancelRunnerJobSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(CancelRunnerJobTestSuite))
}

func (s *CancelRunnerJobTestSuite) SetupSuite() {
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

func (s *CancelRunnerJobTestSuite) SetupTest() {
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

func (s *CancelRunnerJobTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CancelRunnerJobTestSuite) setupTestData() {
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

func (s *CancelRunnerJobTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *CancelRunnerJobTestSuite) TestCancelRunnerJob() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(string)
		expectedNotFound bool
	}{
		{
			name: "successfully cancel job with no executions",
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

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID
			},
			expectedCode: http.StatusAccepted,
			validateFunc: func(jobID string) {
				// Verify job status is cancelled in database
				var job app.RunnerJob
				err := s.service.DB.First(&job, "id = ?", jobID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.RunnerJobStatusCancelled, job.Status)
			},
		},
		{
			name: "successfully cancel job with running execution",
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

				// Create running execution
				exec := &app.RunnerJobExecution{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					Status:      app.RunnerJobExecutionStatusInProgress,
				}
				err = s.service.DB.WithContext(ctx).Create(exec).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(exec)
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID
			},
			expectedCode: http.StatusAccepted,
			validateFunc: func(jobID string) {
				// Verify job status is cancelled
				var job app.RunnerJob
				err := s.service.DB.First(&job, "id = ?", jobID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.RunnerJobStatusCancelled, job.Status)

				// Verify execution is also cancelled
				var exec app.RunnerJobExecution
				err = s.service.DB.First(&exec, "runner_job_id = ?", jobID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.RunnerJobExecutionStatusCancelled, exec.Status)
			},
		},
		{
			name: "cancel job with multiple executions - only running ones cancelled",
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
					Operation:         app.RunnerJobOperationTypeBuild,
					QueueTimeout:      5 * time.Minute,
					AvailableTimeout:  10 * time.Minute,
					ExecutionTimeout:  30 * time.Minute,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				// Create finished execution (should not be cancelled)
				exec1 := &app.RunnerJobExecution{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					Status:      app.RunnerJobExecutionStatusFinished,
				}
				err = s.service.DB.WithContext(ctx).Create(exec1).Error
				require.NoError(s.T(), err)

				// Create running execution (should be cancelled)
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
			expectedCode: http.StatusAccepted,
			validateFunc: func(jobID string) {
				// Verify only running execution is cancelled
				var executions []app.RunnerJobExecution
				err := s.service.DB.Where("runner_job_id = ?", jobID).Order("created_at asc").Find(&executions).Error
				require.NoError(s.T(), err)
				require.Len(s.T(), executions, 2)

				// First execution should remain finished
				assert.Equal(s.T(), app.RunnerJobExecutionStatusFinished, executions[0].Status)
				// Second execution should be cancelled
				assert.Equal(s.T(), app.RunnerJobExecutionStatusCancelled, executions[1].Status)
			},
		},
		{
			name: "cancel job that's already finished - still succeeds",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				job := &app.RunnerJob{
					ID:                domains.NewRunnerJobID(),
					OrgID:             s.testOrg.ID,
					RunnerID:          s.testRunner.ID,
					LogStreamID:       generics.ToPtr(s.testLogStream.ID),
					Status:            app.RunnerJobStatusFinished,
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

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID
			},
			expectedCode: http.StatusAccepted,
			validateFunc: func(jobID string) {
				// Verify job status is now cancelled
				var job app.RunnerJob
				err := s.service.DB.First(&job, "id = ?", jobID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.RunnerJobStatusCancelled, job.Status)
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
			name: "job in different org not accessible",
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
					Status:            app.RunnerJobStatusInProgress,
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
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			jobID := tc.setupFunc()
			rr := s.makeRequest("POST", "/v1/runner-jobs/"+jobID+"/cancel", CancelRunnerJobRequest{})

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
				assert.Equal(s.T(), app.RunnerJobStatusCancelled, job.Status)
				tc.validateFunc(jobID)
			}
		})
	}
}

// TestCancelRunnerJobOnlyMostRecentExecution verifies that cancel only affects the
// most recent execution, since getRunnerJob preloads only the latest one (LIMIT 1).
func (s *CancelRunnerJobTestSuite) TestCancelRunnerJobOnlyMostRecentExecution() {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	job := &app.RunnerJob{
		ID:                domains.NewRunnerJobID(),
		OrgID:             s.testOrg.ID,
		RunnerID:          s.testRunner.ID,
		LogStreamID:       generics.ToPtr(s.testLogStream.ID),
		Status:            app.RunnerJobStatusInProgress,
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

	// Create an older finished execution
	olderExec := &app.RunnerJobExecution{
		ID:          domains.NewRunnerID(),
		OrgID:       s.testOrg.ID,
		RunnerJobID: job.ID,
		Status:      app.RunnerJobExecutionStatusFinished,
	}
	err = s.service.DB.WithContext(ctx).Create(olderExec).Error
	require.NoError(s.T(), err)

	// Small sleep to ensure distinct created_at timestamps
	time.Sleep(10 * time.Millisecond)

	// Create a newer running execution (this is the most recent)
	newerExec := &app.RunnerJobExecution{
		ID:          domains.NewRunnerID(),
		OrgID:       s.testOrg.ID,
		RunnerJobID: job.ID,
		Status:      app.RunnerJobExecutionStatusInProgress,
	}
	err = s.service.DB.WithContext(ctx).Create(newerExec).Error
	require.NoError(s.T(), err)

	// Cancel the job
	rr := s.makeRequest("POST", "/v1/runner-jobs/"+job.ID+"/cancel", CancelRunnerJobRequest{})
	require.Equal(s.T(), http.StatusAccepted, rr.Code)

	// Verify: the most recent (in-progress) execution should be cancelled
	var newerResult app.RunnerJobExecution
	err = s.service.DB.First(&newerResult, "id = ?", newerExec.ID).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), app.RunnerJobExecutionStatusCancelled, newerResult.Status,
		"most recent running execution should be cancelled")

	// Verify: the older (finished) execution should remain unchanged
	var olderResult app.RunnerJobExecution
	err = s.service.DB.First(&olderResult, "id = ?", olderExec.ID).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), app.RunnerJobExecutionStatusFinished, olderResult.Status,
		"older finished execution should remain unchanged")

	// Cleanup
	s.service.DB.Unscoped().Where("runner_job_id = ?", job.ID).Delete(&app.RunnerJobExecution{})
	s.service.DB.Unscoped().Delete(job)
}

// TestCancelRunnerJobTerminalMostRecentExecution verifies that when the most recent
// execution is already terminal, no executions are modified.
func (s *CancelRunnerJobTestSuite) TestCancelRunnerJobTerminalMostRecentExecution() {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	job := &app.RunnerJob{
		ID:                domains.NewRunnerJobID(),
		OrgID:             s.testOrg.ID,
		RunnerID:          s.testRunner.ID,
		LogStreamID:       generics.ToPtr(s.testLogStream.ID),
		Status:            app.RunnerJobStatusInProgress,
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

	// Create a running execution (older)
	olderExec := &app.RunnerJobExecution{
		ID:          domains.NewRunnerID(),
		OrgID:       s.testOrg.ID,
		RunnerJobID: job.ID,
		Status:      app.RunnerJobExecutionStatusPending,
	}
	err = s.service.DB.WithContext(ctx).Create(olderExec).Error
	require.NoError(s.T(), err)

	time.Sleep(10 * time.Millisecond)

	// Create a terminal execution (newer - most recent)
	newerExec := &app.RunnerJobExecution{
		ID:          domains.NewRunnerID(),
		OrgID:       s.testOrg.ID,
		RunnerJobID: job.ID,
		Status:      app.RunnerJobExecutionStatusFailed,
	}
	err = s.service.DB.WithContext(ctx).Create(newerExec).Error
	require.NoError(s.T(), err)

	// Cancel the job
	rr := s.makeRequest("POST", "/v1/runner-jobs/"+job.ID+"/cancel", CancelRunnerJobRequest{})
	require.Equal(s.T(), http.StatusAccepted, rr.Code)

	// Verify: the most recent (failed) execution stays failed
	var newerResult app.RunnerJobExecution
	err = s.service.DB.First(&newerResult, "id = ?", newerExec.ID).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), app.RunnerJobExecutionStatusFailed, newerResult.Status,
		"most recent terminal execution should remain unchanged")

	// Verify: the older (pending) execution also stays pending - it's not loaded by getRunnerJob
	var olderResult app.RunnerJobExecution
	err = s.service.DB.First(&olderResult, "id = ?", olderExec.ID).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), app.RunnerJobExecutionStatusPending, olderResult.Status,
		"older execution should remain unchanged since only the most recent is loaded")

	// Cleanup
	s.service.DB.Unscoped().Where("runner_job_id = ?", job.ID).Delete(&app.RunnerJobExecution{})
	s.service.DB.Unscoped().Delete(job)
}
