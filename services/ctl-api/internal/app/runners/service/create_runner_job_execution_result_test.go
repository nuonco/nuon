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

type CreateRunnerJobExecutionResultTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type CreateRunnerJobExecutionResultTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       CreateRunnerJobExecutionResultTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	testLogStream *app.LogStream
}

func TestCreateRunnerJobExecutionResultSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(CreateRunnerJobExecutionResultTestSuite))
}

func (s *CreateRunnerJobExecutionResultTestSuite) SetupSuite() {
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

func (s *CreateRunnerJobExecutionResultTestSuite) SetupTest() {
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

func (s *CreateRunnerJobExecutionResultTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CreateRunnerJobExecutionResultTestSuite) setupTestData() {
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

func (s *CreateRunnerJobExecutionResultTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *CreateRunnerJobExecutionResultTestSuite) TestCreateRunnerJobExecutionResult() {
	testCases := []struct {
		name             string
		setupFunc        func() (string, string)
		expectedCode     int
		expectedNotFound bool
	}{
		{
			name: "runner job not found",
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

				// Return non-existent job ID
				return "rjbnonexistent123456789012", execution.ID
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			jobID, execID := tc.setupFunc()
			path := "/v1/runner-jobs/" + jobID + "/executions/" + execID + "/result"

			req := CreateRunnerJobExecutionResultRequest{
				Success:  true,
				Contents: "test result contents",
			}
			rr := s.makeRequest("POST", path, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}
		})
	}
}

func (s *CreateRunnerJobExecutionResultTestSuite) TestCreateRunnerJobExecutionResultRequiresAccountContext() {
	// This test documents a known limitation: CreateRunnerJobExecutionResult requires
	// CreatedByID but runner routes don't have account context.
	// The endpoint will fail with a database constraint error when trying to create the result.

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

	path := "/v1/runner-jobs/" + job.ID + "/executions/" + execution.ID + "/result"

	req := CreateRunnerJobExecutionResultRequest{
		Success:  true,
		Contents: "test result contents",
	}
	rr := s.makeRequest("POST", path, req)

	// Expect 500 due to CreatedByID constraint
	// RunnerJobExecutionResult has CreatedByID NOT NULL but runner routes have no account context
	assert.Equal(s.T(), http.StatusInternalServerError, rr.Code)
	s.T().Logf("Expected error due to CreatedByID constraint - Body: %s", rr.Body.String())
	assert.Contains(s.T(), rr.Body.String(), "created_by_id")
}
