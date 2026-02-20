package service

import (
	"context"
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

type GetRunnerJobPlanPublicTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetRunnerJobPlanPublicTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       GetRunnerJobPlanPublicTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	testLogStream *app.LogStream
}

func TestGetRunnerJobPlanPublicSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetRunnerJobPlanPublicTestSuite))
}

func (s *GetRunnerJobPlanPublicTestSuite) SetupSuite() {
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

func (s *GetRunnerJobPlanPublicTestSuite) SetupTest() {
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

func (s *GetRunnerJobPlanPublicTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetRunnerJobPlanPublicTestSuite) setupTestData() {
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

func (s *GetRunnerJobPlanPublicTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetRunnerJobPlanPublicTestSuite) TestGetRunnerJobPlanPublic() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		expectedPlan     string
		expectedNotFound bool
	}{
		{
			name: "successfully get runner job plan",
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

				plan := &app.RunnerJobPlan{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					PlanJSON:    `{"buildConfig": {"dockerfile": "Dockerfile"}}`,
				}
				err = s.service.DB.WithContext(ctx).Create(plan).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(plan)
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID
			},
			expectedCode: http.StatusOK,
			expectedPlan: `{"buildConfig": {"dockerfile": "Dockerfile"}}`,
		},
		{
			name: "runner job plan not found",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create job without plan
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
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "plan in different org not accessible",
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

				plan := &app.RunnerJobPlan{
					ID:          domains.NewRunnerID(),
					OrgID:       org2.ID,
					RunnerJobID: job.ID,
					PlanJSON:    `{"secret": "data"}`,
				}
				err = s.service.DB.WithContext(ctx).Create(plan).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(plan)
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
			name: "plan with complex JSON structure",
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

				complexPlan := `{
					"resources": [
						{"type": "aws_instance", "name": "web"},
						{"type": "aws_security_group", "name": "web_sg"}
					],
					"variables": {
						"instance_type": "t3.micro",
						"region": "us-west-2"
					}
				}`

				plan := &app.RunnerJobPlan{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					PlanJSON:    complexPlan,
				}
				err = s.service.DB.WithContext(ctx).Create(plan).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(plan)
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "empty plan JSON",
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
					Group:             app.RunnerJobGroupBuild,
					Type:              app.RunnerJobTypeNOOPBuild,
					Operation:         app.RunnerJobOperationTypeBuild,
					QueueTimeout:      5 * time.Minute,
					AvailableTimeout:  10 * time.Minute,
					ExecutionTimeout:  30 * time.Minute,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				plan := &app.RunnerJobPlan{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					PlanJSON:    "",
				}
				err = s.service.DB.WithContext(ctx).Create(plan).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(plan)
					s.service.DB.Unscoped().Delete(job)
				})

				return job.ID
			},
			expectedCode: http.StatusOK,
			expectedPlan: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			jobID := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/runner-jobs/"+jobID+"/plan")

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else {
				planJSON := rr.Body.String()
				if tc.expectedPlan != "" {
					assert.Equal(s.T(), tc.expectedPlan, planJSON)
				}
			}
		})
	}
}

func (s *GetRunnerJobPlanPublicTestSuite) TestGetRunnerJobPlanDifferentJobTypes() {
	jobTypes := []struct {
		group app.RunnerJobGroup
		typ   app.RunnerJobType
		plan  string
	}{
		{app.RunnerJobGroupBuild, app.RunnerJobTypeDockerBuild, `{"dockerfile": "Dockerfile"}`},
		{app.RunnerJobGroupBuild, app.RunnerJobTypeHelmChartBuild, `{"chart": "mychart"}`},
		{app.RunnerJobGroupBuild, app.RunnerJobTypeTerraformModuleBuild, `{"module": "mymodule"}`},
		{app.RunnerJobGroupDeploy, app.RunnerJobTypeTerraformDeploy, `{"tf_vars": {}}`},
		{app.RunnerJobGroupDeploy, app.RunnerJobTypeHelmChartDeploy, `{"values": {}}`},
		{app.RunnerJobGroupSync, app.RunnerJobTypeOCISync, `{"registry": "ecr"}`},
	}

	for _, jt := range jobTypes {
		s.Run(string(jt.typ), func() {
			ctx := context.Background()
			ctx = cctx.SetAccountContext(ctx, s.testAcc)

			job := &app.RunnerJob{
				ID:                domains.NewRunnerJobID(),
				OrgID:             s.testOrg.ID,
				RunnerID:          s.testRunner.ID,
				LogStreamID:       generics.ToPtr(s.testLogStream.ID),
				Status:            app.RunnerJobStatusAvailable,
				StatusDescription: "test job",
				Group:             jt.group,
				Type:              jt.typ,
				Operation:         app.RunnerJobOperationTypeBuild,
				QueueTimeout:      5 * time.Minute,
				AvailableTimeout:  10 * time.Minute,
				ExecutionTimeout:  30 * time.Minute,
				MaxExecutions:     3,
			}
			err := s.service.DB.WithContext(ctx).Create(job).Error
			require.NoError(s.T(), err)

			plan := &app.RunnerJobPlan{
				ID:          domains.NewRunnerID(),
				OrgID:       s.testOrg.ID,
				RunnerJobID: job.ID,
				PlanJSON:    jt.plan,
			}
			err = s.service.DB.WithContext(ctx).Create(plan).Error
			require.NoError(s.T(), err)

			rr := s.makeRequest("GET", "/v1/runner-jobs/"+job.ID+"/plan")
			require.Equal(s.T(), http.StatusOK, rr.Code)
			assert.Equal(s.T(), jt.plan, rr.Body.String())

			s.service.DB.Unscoped().Delete(plan)
			s.service.DB.Unscoped().Delete(job)
		})
	}
}

// V2 Test Suite for GetRunnerJobPlanV2 (runner routes)

type GetRunnerJobPlanV2TestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetRunnerJobPlanV2TestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       GetRunnerJobPlanV2TestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	testLogStream *app.LogStream
}

func TestGetRunnerJobPlanV2Suite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetRunnerJobPlanV2TestSuite))
}

func (s *GetRunnerJobPlanV2TestSuite) SetupSuite() {
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

func (s *GetRunnerJobPlanV2TestSuite) SetupTest() {
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

func (s *GetRunnerJobPlanV2TestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetRunnerJobPlanV2TestSuite) setupTestData() {
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

func (s *GetRunnerJobPlanV2TestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetRunnerJobPlanV2TestSuite) TestGetRunnerJobPlanV2() {
	testCases := []struct {
		name             string
		setupFunc        func() (string, string)
		expectedCode     int
		expectedPlan     string
		expectedNotFound bool
	}{
		{
			name: "successfully get runner job plan",
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
					QueueTimeout:      5 * time.Minute,
					AvailableTimeout:  10 * time.Minute,
					ExecutionTimeout:  30 * time.Minute,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				plan := &app.RunnerJobPlan{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					PlanJSON:    `{"buildConfig": {"dockerfile": "Dockerfile"}}`,
				}
				err = s.service.DB.WithContext(ctx).Create(plan).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(plan)
					s.service.DB.Unscoped().Delete(job)
				})

				return s.testRunner.ID, job.ID
			},
			expectedCode: http.StatusOK,
			expectedPlan: `{"buildConfig": {"dockerfile": "Dockerfile"}}`,
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
					QueueTimeout:      5 * time.Minute,
					AvailableTimeout:  10 * time.Minute,
					ExecutionTimeout:  30 * time.Minute,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				plan := &app.RunnerJobPlan{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					PlanJSON:    `{"buildConfig": {"dockerfile": "Dockerfile"}}`,
				}
				err = s.service.DB.WithContext(ctx).Create(plan).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(plan)
					s.service.DB.Unscoped().Delete(job)
				})

				return "runnonexistent123456789012", job.ID
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "runner job plan not found",
			setupFunc: func() (string, string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create job without plan
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

				return s.testRunner.ID, job.ID
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "empty plan JSON returns successfully",
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
					Type:              app.RunnerJobTypeNOOPBuild,
					Operation:         app.RunnerJobOperationTypeBuild,
					QueueTimeout:      5 * time.Minute,
					AvailableTimeout:  10 * time.Minute,
					ExecutionTimeout:  30 * time.Minute,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				plan := &app.RunnerJobPlan{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					PlanJSON:    "",
				}
				err = s.service.DB.WithContext(ctx).Create(plan).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(plan)
					s.service.DB.Unscoped().Delete(job)
				})

				return s.testRunner.ID, job.ID
			},
			expectedCode: http.StatusOK,
			expectedPlan: "",
		},
		{
			name: "plan with complex JSON structure",
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

				complexPlan := `{
					"resources": [
						{"type": "aws_instance", "name": "web"},
						{"type": "aws_security_group", "name": "web_sg"}
					],
					"variables": {
						"instance_type": "t3.micro",
						"region": "us-west-2"
					}
				}`

				plan := &app.RunnerJobPlan{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					PlanJSON:    complexPlan,
				}
				err = s.service.DB.WithContext(ctx).Create(plan).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(plan)
					s.service.DB.Unscoped().Delete(job)
				})

				return s.testRunner.ID, job.ID
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID, jobID := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/runners/"+runnerID+"/jobs/"+jobID+"/plan")

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else {
				planJSON := rr.Body.String()
				if tc.expectedPlan != "" {
					assert.Equal(s.T(), tc.expectedPlan, planJSON)
				}
			}
		})
	}
}
