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
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type GetRunnerJobCompositePlanTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetRunnerJobCompositePlanTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       GetRunnerJobCompositePlanTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	testLogStream *app.LogStream
}

func TestGetRunnerJobCompositePlanSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetRunnerJobCompositePlanTestSuite))
}

func (s *GetRunnerJobCompositePlanTestSuite) SetupSuite() {
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

func (s *GetRunnerJobCompositePlanTestSuite) SetupTest() {
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

func (s *GetRunnerJobCompositePlanTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetRunnerJobCompositePlanTestSuite) setupTestData() {
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

func (s *GetRunnerJobCompositePlanTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetRunnerJobCompositePlanTestSuite) TestGetRunnerJobCompositePlan() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(*plantypes.CompositePlan)
		expectedNotFound bool
	}{
		{
			name: "successfully get composite plan with stored composite plan",
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

				compositePlan := plantypes.CompositePlan{
					BuildPlan: &plantypes.BuildPlan{
						ComponentBuildID: "cmp123",
					},
				}

				plan := &app.RunnerJobPlan{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					RunnerJobID:   job.ID,
					PlanJSON:      `{}`,
					CompositePlan: compositePlan,
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
			validateFunc: func(cp *plantypes.CompositePlan) {
				assert.NotNil(s.T(), cp.BuildPlan)
				assert.Equal(s.T(), "cmp123", cp.BuildPlan.ComponentBuildID)
			},
		},
		{
			name: "derive composite plan from plan JSON - build job",
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

				buildPlanJSON := `{"component_build_id": "cmp456"}`
				plan := &app.RunnerJobPlan{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					PlanJSON:    buildPlanJSON,
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
			validateFunc: func(cp *plantypes.CompositePlan) {
				assert.NotNil(s.T(), cp.BuildPlan)
				assert.Equal(s.T(), "cmp456", cp.BuildPlan.ComponentBuildID)
			},
		},
		{
			name: "derive composite plan from plan JSON - deploy job",
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
					Operation:         app.RunnerJobOperationTypeBuild,
					QueueTimeout:      5 * time.Minute,
					AvailableTimeout:  10 * time.Minute,
					ExecutionTimeout:  30 * time.Minute,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				deployPlanJSON := `{"install_id": "inst789"}`
				plan := &app.RunnerJobPlan{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					PlanJSON:    deployPlanJSON,
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
			validateFunc: func(cp *plantypes.CompositePlan) {
				assert.NotNil(s.T(), cp.DeployPlan)
				assert.Equal(s.T(), "inst789", cp.DeployPlan.InstallID)
			},
		},
		{
			name: "derive composite plan from plan JSON - sync OCI job",
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
					Group:             app.RunnerJobGroupSync,
					Type:              app.RunnerJobTypeOCISync,
					Operation:         app.RunnerJobOperationTypeBuild,
					QueueTimeout:      5 * time.Minute,
					AvailableTimeout:  10 * time.Minute,
					ExecutionTimeout:  30 * time.Minute,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				syncPlanJSON := `{"src_tag": "sync123"}`
				plan := &app.RunnerJobPlan{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					PlanJSON:    syncPlanJSON,
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
			validateFunc: func(cp *plantypes.CompositePlan) {
				assert.NotNil(s.T(), cp.SyncOCIPlan)
				assert.Equal(s.T(), "sync123", cp.SyncOCIPlan.SrcTag)
			},
		},
		{
			name: "derive composite plan from plan JSON - actions job",
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
					Group:             app.RunnerJobGroupActions,
					Type:              app.RunnerJobTypeNOOPBuild,
					Operation:         app.RunnerJobOperationTypeBuild,
					QueueTimeout:      5 * time.Minute,
					AvailableTimeout:  10 * time.Minute,
					ExecutionTimeout:  30 * time.Minute,
					MaxExecutions:     3,
				}
				err := s.service.DB.WithContext(ctx).Create(job).Error
				require.NoError(s.T(), err)

				actionPlanJSON := `{"id": "act123"}`
				plan := &app.RunnerJobPlan{
					ID:          domains.NewRunnerID(),
					OrgID:       s.testOrg.ID,
					RunnerJobID: job.ID,
					PlanJSON:    actionPlanJSON,
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
			validateFunc: func(cp *plantypes.CompositePlan) {
				assert.NotNil(s.T(), cp.ActionWorkflowRunPlan)
				assert.Equal(s.T(), "act123", cp.ActionWorkflowRunPlan.ID)
			},
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
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			jobID := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/runner-jobs/"+jobID+"/composite-plan")

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else if tc.validateFunc != nil {
				var cp plantypes.CompositePlan
				err := json.Unmarshal(rr.Body.Bytes(), &cp)
				require.NoError(s.T(), err)
				tc.validateFunc(&cp)
			}
		})
	}
}

func (s *GetRunnerJobCompositePlanTestSuite) TestGetRunnerJobCompositePlanAllJobGroups() {
	testCases := []struct {
		group        app.RunnerJobGroup
		typ          app.RunnerJobType
		planJSON     string
		validateFunc func(*plantypes.CompositePlan)
	}{
		{
			group:    app.RunnerJobGroupBuild,
			typ:      app.RunnerJobTypeDockerBuild,
			planJSON: `{"component_build_id": "build123"}`,
			validateFunc: func(cp *plantypes.CompositePlan) {
				assert.NotNil(s.T(), cp.BuildPlan)
			},
		},
		{
			group:    app.RunnerJobGroupDeploy,
			typ:      app.RunnerJobTypeTerraformDeploy,
			planJSON: `{"install_id": "deploy123"}`,
			validateFunc: func(cp *plantypes.CompositePlan) {
				assert.NotNil(s.T(), cp.DeployPlan)
			},
		},
		{
			group:    app.RunnerJobGroupSync,
			typ:      app.RunnerJobTypeOCISync,
			planJSON: `{"src_tag": "sync123"}`,
			validateFunc: func(cp *plantypes.CompositePlan) {
				assert.NotNil(s.T(), cp.SyncOCIPlan)
			},
		},
		{
			group:    app.RunnerJobGroupActions,
			typ:      app.RunnerJobTypeNOOPBuild,
			planJSON: `{"id": "action123"}`,
			validateFunc: func(cp *plantypes.CompositePlan) {
				assert.NotNil(s.T(), cp.ActionWorkflowRunPlan)
			},
		},
		{
			group:    app.RunnerJobGroupSandbox,
			typ:      app.RunnerJobTypeNOOPBuild,
			planJSON: `{"sandbox_id": "sandbox123"}`,
			validateFunc: func(cp *plantypes.CompositePlan) {
				assert.NotNil(s.T(), cp.SandboxRunPlan)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(string(tc.group), func() {
			ctx := context.Background()
			ctx = cctx.SetAccountContext(ctx, s.testAcc)

			job := &app.RunnerJob{
				ID:                domains.NewRunnerJobID(),
				OrgID:             s.testOrg.ID,
				RunnerID:          s.testRunner.ID,
				LogStreamID:       generics.ToPtr(s.testLogStream.ID),
				Status:            app.RunnerJobStatusAvailable,
				StatusDescription: "test job",
				Group:             tc.group,
				Type:              tc.typ,
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
				PlanJSON:    tc.planJSON,
			}
			err = s.service.DB.WithContext(ctx).Create(plan).Error
			require.NoError(s.T(), err)

			rr := s.makeRequest("GET", "/v1/runner-jobs/"+job.ID+"/composite-plan")
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var cp plantypes.CompositePlan
			err = json.Unmarshal(rr.Body.Bytes(), &cp)
			require.NoError(s.T(), err)
			tc.validateFunc(&cp)

			s.service.DB.Unscoped().Delete(plan)
			s.service.DB.Unscoped().Delete(job)
		})
	}
}
