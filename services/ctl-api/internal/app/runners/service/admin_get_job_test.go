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

type AdminGetJobTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminGetJobTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       AdminGetJobTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	testLogStream *app.LogStream
}

func TestAdminGetJobSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminGetJobTestSuite))
}

func (s *AdminGetJobTestSuite) SetupSuite() {
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

func (s *AdminGetJobTestSuite) SetupTest() {
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

func (s *AdminGetJobTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminGetJobTestSuite) setupTestData() {
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

func (s *AdminGetJobTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminGetJobTestSuite) TestAdminGetRunnerJob() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(*app.RunnerJob)
		expectedNotFound bool
	}{
		{
			name: "successfully get runner job by ID",
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
			expectedCode: http.StatusOK,
			validateFunc: func(job *app.RunnerJob) {
				assert.Equal(s.T(), s.testOrg.ID, job.OrgID)
				assert.Equal(s.T(), s.testRunner.ID, job.RunnerID)
				assert.Equal(s.T(), app.RunnerJobStatusAvailable, job.Status)
				assert.Equal(s.T(), "test job", job.StatusDescription)
			},
		},
		{
			name: "successfully get runner job by owner_id",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				ownerID := domains.NewBuildID()
				job := &app.RunnerJob{
					ID:                domains.NewRunnerJobID(),
					OrgID:             s.testOrg.ID,
					RunnerID:          s.testRunner.ID,
					LogStreamID:       generics.ToPtr(s.testLogStream.ID),
					OwnerID:           ownerID,
					Status:            app.RunnerJobStatusInProgress,
					StatusDescription: "job with owner",
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

				return ownerID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(job *app.RunnerJob) {
				assert.Equal(s.T(), app.RunnerJobStatusInProgress, job.Status)
				assert.Equal(s.T(), "job with owner", job.StatusDescription)
			},
		},
		{
			name: "job with execution preloaded",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				job := &app.RunnerJob{
					ID:                domains.NewRunnerJobID(),
					OrgID:             s.testOrg.ID,
					RunnerID:          s.testRunner.ID,
					LogStreamID:       generics.ToPtr(s.testLogStream.ID),
					Status:            app.RunnerJobStatusInProgress,
					StatusDescription: "job with execution",
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
			expectedCode: http.StatusOK,
			validateFunc: func(job *app.RunnerJob) {
				// Verify last execution is preloaded
				assert.NotEmpty(s.T(), job.Executions)
				assert.Len(s.T(), job.Executions, 1)
				assert.Equal(s.T(), app.RunnerJobExecutionStatusInProgress, job.Executions[0].Status)
			},
		},
		{
			name: "runner job not found returns error",
			setupFunc: func() string {
				return "rjbnonexistent123456789012"
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
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
