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

type GetRunnerJobV2TestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetRunnerJobV2TestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       GetRunnerJobV2TestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	testLogStream *app.LogStream
}

func TestGetRunnerJobV2Suite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetRunnerJobV2TestSuite))
}

func (s *GetRunnerJobV2TestSuite) SetupSuite() {
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

func (s *GetRunnerJobV2TestSuite) SetupTest() {
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

func (s *GetRunnerJobV2TestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetRunnerJobV2TestSuite) setupTestData() {
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

func (s *GetRunnerJobV2TestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetRunnerJobV2TestSuite) TestGetRunnerJobV2() {
	testCases := []struct {
		name             string
		setupFunc        func() (string, string)
		expectedCode     int
		validateFunc     func(*app.RunnerJob)
		expectedNotFound bool
	}{
		{
			name: "success returns job with preloaded execution",
			setupFunc: func() (string, string) {
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
					QueueTimeout:      60,
					AvailableTimeout:  60,
					ExecutionTimeout:  300,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				// Create execution
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

				return s.testRunner.ID, job.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(job *app.RunnerJob) {
				assert.NotNil(s.T(), job)
				assert.Equal(s.T(), app.RunnerJobStatusInProgress, job.Status)
				assert.Len(s.T(), job.Executions, 1, "should preload most recent execution")
				assert.Equal(s.T(), app.RunnerJobExecutionStatusInProgress, job.Executions[0].Status)
			},
		},
		{
			name: "runner not found returns error",
			setupFunc: func() (string, string) {
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
					QueueTimeout:      60,
					AvailableTimeout:  60,
					ExecutionTimeout:  300,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(job)
				})

				return "runnonexistent123456789012", job.ID
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "job not found returns error",
			setupFunc: func() (string, string) {
				return s.testRunner.ID, "rjbnonexistent123456789012"
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "job exists but runner not found returns error",
			setupFunc: func() (string, string) {
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
					QueueTimeout:      60,
					AvailableTimeout:  60,
					ExecutionTimeout:  300,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(job)
				})

				return "runnonexistent123456789012", job.ID
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "job without execution returns successfully",
			setupFunc: func() (string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				job := &app.RunnerJob{
					ID:                domains.NewRunnerJobID(),
					OrgID:             s.testOrg.ID,
					RunnerID:          s.testRunner.ID,
					LogStreamID:       generics.ToPtr(s.testLogStream.ID),
					Status:            app.RunnerJobStatusQueued,
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

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(job)
				})

				return s.testRunner.ID, job.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(job *app.RunnerJob) {
				assert.NotNil(s.T(), job)
				assert.Equal(s.T(), app.RunnerJobStatusQueued, job.Status)
				assert.Empty(s.T(), job.Executions, "should have no executions")
			},
		},
		{
			name: "job with multiple executions returns only most recent",
			setupFunc: func() (string, string) {
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
					QueueTimeout:      60,
					AvailableTimeout:  60,
					ExecutionTimeout:  300,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				// Create older execution (finished)
				exec1 := &app.RunnerJobExecution{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					Status:      app.RunnerJobExecutionStatusFinished,
				}
				err = s.service.DB.WithContext(ctx).Create(exec1).Error
				require.NoError(s.T(), err)

				// Create newer execution (in progress) - this should be preloaded
				exec2 := &app.RunnerJobExecution{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					Status:      app.RunnerJobExecutionStatusInProgress,
				}
				err = s.service.DB.WithContext(ctx).Create(exec2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Where("runner_job_id = ?", job.ID).Delete(&app.RunnerJobExecution{})
					s.service.DB.Unscoped().Delete(job)
				})

				return s.testRunner.ID, job.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(job *app.RunnerJob) {
				assert.NotNil(s.T(), job)
				assert.Len(s.T(), job.Executions, 1, "should preload only most recent execution (LIMIT 1)")
				assert.Equal(s.T(), app.RunnerJobExecutionStatusInProgress, job.Executions[0].Status,
					"should be the most recent (in-progress) execution")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID, jobID := tc.setupFunc()
			path := "/v1/runners/" + runnerID + "/jobs/" + jobID
			rr := s.makeRequest("GET", path)

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
