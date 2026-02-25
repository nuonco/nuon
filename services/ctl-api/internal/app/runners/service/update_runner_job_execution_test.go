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

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type UpdateRunnerJobExecutionTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type UpdateRunnerJobExecutionTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       UpdateRunnerJobExecutionTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	testLogStream *app.LogStream
}

func TestUpdateRunnerJobExecutionSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(UpdateRunnerJobExecutionTestSuite))
}

func (s *UpdateRunnerJobExecutionTestSuite) SetupSuite() {
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

func (s *UpdateRunnerJobExecutionTestSuite) SetupTest() {
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

func (s *UpdateRunnerJobExecutionTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *UpdateRunnerJobExecutionTestSuite) setupTestData() {
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

func (s *UpdateRunnerJobExecutionTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *UpdateRunnerJobExecutionTestSuite) TestUpdateRunnerJobExecution() {
	testCases := []struct {
		name             string
		setupFunc        func() (string, string)
		newStatus        app.RunnerJobExecutionStatus
		expectedCode     int
		validateFunc     func(string, string)
		expectedNotFound bool
	}{
		{
			name: "successfully update status to in-progress",
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

				execution := &app.RunnerJobExecution{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					Status:      app.RunnerJobExecutionStatusPending,
				}
				err = s.service.DB.WithContext(ctx).Create(execution).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(execution)
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID, execution.ID
			},
			newStatus:    app.RunnerJobExecutionStatusInProgress,
			expectedCode: http.StatusOK,
			validateFunc: func(jobID, execID string) {
				var exec app.RunnerJobExecution
				err := s.service.DB.First(&exec, "id = ?", execID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.RunnerJobExecutionStatusInProgress, exec.Status)
			},
		},
		{
			name: "successfully update status to finished",
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

				execution := &app.RunnerJobExecution{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					Status:      app.RunnerJobExecutionStatusInProgress,
				}
				err = s.service.DB.WithContext(ctx).Create(execution).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(execution)
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID, execution.ID
			},
			newStatus:    app.RunnerJobExecutionStatusFinished,
			expectedCode: http.StatusOK,
			validateFunc: func(jobID, execID string) {
				var exec app.RunnerJobExecution
				err := s.service.DB.First(&exec, "id = ?", execID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.RunnerJobExecutionStatusFinished, exec.Status)
			},
		},
		{
			name: "successfully update status to failed",
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

				execution := &app.RunnerJobExecution{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					Status:      app.RunnerJobExecutionStatusInProgress,
				}
				err = s.service.DB.WithContext(ctx).Create(execution).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(execution)
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID, execution.ID
			},
			newStatus:    app.RunnerJobExecutionStatusFailed,
			expectedCode: http.StatusOK,
			validateFunc: func(jobID, execID string) {
				var exec app.RunnerJobExecution
				err := s.service.DB.First(&exec, "id = ?", execID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.RunnerJobExecutionStatusFailed, exec.Status)
			},
		},
		{
			name: "successfully update status to cancelled",
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

				execution := &app.RunnerJobExecution{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					Status:      app.RunnerJobExecutionStatusInProgress,
				}
				err = s.service.DB.WithContext(ctx).Create(execution).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(execution)
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID, execution.ID
			},
			newStatus:    app.RunnerJobExecutionStatusCancelled,
			expectedCode: http.StatusOK,
			validateFunc: func(jobID, execID string) {
				var exec app.RunnerJobExecution
				err := s.service.DB.First(&exec, "id = ?", execID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.RunnerJobExecutionStatusCancelled, exec.Status)
			},
		},
		{
			name: "successfully update status to initializing",
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

				execution := &app.RunnerJobExecution{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					Status:      app.RunnerJobExecutionStatusPending,
				}
				err = s.service.DB.WithContext(ctx).Create(execution).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(execution)
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID, execution.ID
			},
			newStatus:    app.RunnerJobExecutionStatusInitializing,
			expectedCode: http.StatusOK,
			validateFunc: func(jobID, execID string) {
				var exec app.RunnerJobExecution
				err := s.service.DB.First(&exec, "id = ?", execID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.RunnerJobExecutionStatusInitializing, exec.Status)
			},
		},
		{
			name: "execution not found",
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

				return job.ID, "rjexnonexistent12345678901"
			},
			newStatus:        app.RunnerJobExecutionStatusInProgress,
			expectedCode:     http.StatusOK,
			expectedNotFound: false, // Update returns 200 even when no rows matched
			validateFunc: func(jobID, execID string) {
				// Verify no execution exists with this ID
				var exec app.RunnerJobExecution
				err := s.service.DB.First(&exec, "id = ?", execID).Error
				assert.Error(s.T(), err, "execution should not exist")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			jobID, execID := tc.setupFunc()
			path := "/v1/runner-jobs/" + jobID + "/executions/" + execID

			req := UpdateRunnerJobExecutionRequest{
				Status: tc.newStatus,
			}
			rr := s.makeRequest("PATCH", path, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc(jobID, execID)
			}
		})
	}
}
