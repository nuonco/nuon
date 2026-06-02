package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	comphelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// CreateInstallActionRunTestService holds all fx-injected dependencies for create install action run tests.
type CreateInstallActionRunTestService struct {
	fx.In

	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	VcsHelpers     *vcshelpers.Helpers
	CompHelpers    *comphelpers.Helpers
	ActionsHelpers *actionshelpers.Helpers
	InstallHelpers *installhelpers.Helpers
	ActionsService *service
	Seeder         *testseed.Seeder
}

// CreateInstallActionRunTestSuite is the testify suite for CreateInstallActionRun endpoint.
type CreateInstallActionRunTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service CreateInstallActionRunTestService
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
	testApp *app.App
}

func TestCreateInstallActionRunSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(CreateInstallActionRunTestSuite))
}

func (s *CreateInstallActionRunTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			CustomValidator: true,
		}),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *CreateInstallActionRunTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Reset mock before each test

	// Create test router with standard middlewares using helper
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.ActionsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *CreateInstallActionRunTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CreateInstallActionRunTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *CreateInstallActionRunTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *CreateInstallActionRunTestSuite) TestCreateActionRunSuccess() {
	testCases := []struct {
		name         string
		setupFunc    func() (string, string) // Returns installID, actionConfigID
		requestFunc  func(configID string) CreateInstallActionWorkflowRunRequest
		expectedCode int
		validateFunc func(installID string)
	}{
		{
			name: "manual trigger without env vars",
			setupFunc: func() (string, string) {
				action, appConfig, install := s.setupActionWorkflowWithInstall()
				config := s.createActionConfigWithManualTrigger(action.ID, appConfig.ID)
				installAction := s.createInstallActionWorkflow(install.ID, action.ID)
				s.T().Logf("Created install action workflow: %s", installAction.ID)
				return install.ID, config.ID
			},
			requestFunc: func(configID string) CreateInstallActionWorkflowRunRequest {
				return CreateInstallActionWorkflowRunRequest{
					ActionWorkFlowConfigID: configID,
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(installID string) {
				// Verify workflow was created
				var workflows []app.Workflow
				res := s.service.DB.Where("owner_id = ? AND owner_type = ?", installID, "installs").Find(&workflows)
				require.NoError(s.T(), res.Error)
				assert.Len(s.T(), workflows, 1)
				assert.Equal(s.T(), app.WorkflowTypeActionWorkflowRun, workflows[0].Type)

				// Verify signal was sent
				queueSignals := tests.GetQueueSignals(s.T(), s.service.DB)
				require.Len(s.T(), queueSignals, 1)
				assert.Equal(s.T(), installID, queueSignals[0].OwnerID)
			},
		},
		{
			name: "manual trigger with env vars",
			setupFunc: func() (string, string) {
				action, appConfig, install := s.setupActionWorkflowWithInstall()
				config := s.createActionConfigWithManualTrigger(action.ID, appConfig.ID)
				s.createInstallActionWorkflow(install.ID, action.ID)
				return install.ID, config.ID
			},
			requestFunc: func(configID string) CreateInstallActionWorkflowRunRequest {
				return CreateInstallActionWorkflowRunRequest{
					ActionWorkFlowConfigID: configID,
					RunEnvVars: map[string]string{
						"TEST_VAR": "test_value",
						"FOO":      "bar",
					},
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(installID string) {
				// Verify workflow has env vars with RUNENV_ prefix
				var workflow app.Workflow
				res := s.service.DB.Where("owner_id = ? AND owner_type = ?", installID, "installs").First(&workflow)
				require.NoError(s.T(), res.Error)

				metadata := workflow.Metadata
				assert.Contains(s.T(), metadata, "RUNENV_TEST_VAR")
				assert.Contains(s.T(), metadata, "RUNENV_FOO")
				// HSTORE values are *string
				require.NotNil(s.T(), metadata["RUNENV_TEST_VAR"])
				require.NotNil(s.T(), metadata["RUNENV_FOO"])
				assert.Equal(s.T(), "test_value", *metadata["RUNENV_TEST_VAR"])
				assert.Equal(s.T(), "bar", *metadata["RUNENV_FOO"])
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock at the start of each test case

			installID, configID := tc.setupFunc()

			req := tc.requestFunc(configID)
			path := fmt.Sprintf("/v1/installs/%s/actions/runs", installID)
			rr := s.makeRequest(http.MethodPost, path, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc(installID)
			}

			s.T().Cleanup(func() {
				s.service.DB.Unscoped().Where("id = ?", installID).Delete(&app.Install{})
			})
		})
	}
}

func (s *CreateInstallActionRunTestSuite) TestCreateActionRunValidation() {
	action, appConfig, install := s.setupActionWorkflowWithInstall()
	// Create install action workflow once — it has a unique constraint on (install_id, action_workflow_id)
	s.createInstallActionWorkflow(install.ID, action.ID)

	testCases := []struct {
		name         string
		setupFunc    func() string // Returns configID
		requestFunc  func(configID string) CreateInstallActionWorkflowRunRequest
		expectedCode int
	}{
		{
			name: "missing action_workflow_config_id",
			setupFunc: func() string {
				config := s.createActionConfigWithManualTrigger(action.ID, appConfig.ID)
				return config.ID
			},
			requestFunc: func(configID string) CreateInstallActionWorkflowRunRequest {
				return CreateInstallActionWorkflowRunRequest{}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "config belongs to different app config",
			setupFunc: func() string {
				// Create config with different app config than install
				otherAppConfig := s.createAppConfig(s.testApp.ID)
				config := s.createActionConfigWithManualTrigger(action.ID, otherAppConfig.ID)
				return config.ID
			},
			requestFunc: func(configID string) CreateInstallActionWorkflowRunRequest {
				return CreateInstallActionWorkflowRunRequest{
					ActionWorkFlowConfigID: configID,
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "config without manual trigger",
			setupFunc: func() string {
				// Use a separate appConfig to avoid unique constraint on (action_workflow_id, app_config_id)
				otherAppConfig := s.createAppConfig(s.testApp.ID)
				config := &app.ActionWorkflowConfig{
					ID:               domains.NewActionWorkflowConfigID(),
					OrgID:            s.testOrg.ID,
					AppID:            s.testApp.ID,
					AppConfigID:      otherAppConfig.ID,
					ActionWorkflowID: action.ID,
					Timeout:          5 * time.Minute,
				}
				ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
				ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
				res := s.service.DB.WithContext(ctx).Create(config)
				require.NoError(s.T(), res.Error)

				// Create trigger that is NOT manual
				trigger := &app.ActionWorkflowTriggerConfig{
					ID:                     domains.NewActionWorkflowTriggerConfigID(),
					OrgID:                  s.testOrg.ID,
					AppID:                  s.testApp.ID,
					AppConfigID:            otherAppConfig.ID,
					ActionWorkflowConfigID: config.ID,
					Type:                   app.ActionWorkflowTriggerTypeCron,
					CronSchedule:           "0 0 * * *",
				}
				res = s.service.DB.WithContext(ctx).Create(trigger)
				require.NoError(s.T(), res.Error)

				return config.ID
			},
			requestFunc: func(configID string) CreateInstallActionWorkflowRunRequest {
				return CreateInstallActionWorkflowRunRequest{
					ActionWorkFlowConfigID: configID,
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "non-existent config",
			setupFunc: func() string {
				return domains.NewActionWorkflowConfigID()
			},
			requestFunc: func(configID string) CreateInstallActionWorkflowRunRequest {
				return CreateInstallActionWorkflowRunRequest{
					ActionWorkFlowConfigID: configID,
				}
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			configID := tc.setupFunc()

			req := tc.requestFunc(configID)
			path := fmt.Sprintf("/v1/installs/%s/actions/runs", install.ID)
			rr := s.makeRequest(http.MethodPost, path, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			assert.Equal(s.T(), tc.expectedCode, rr.Code)
		})
	}
}

func (s *CreateInstallActionRunTestSuite) TestCreateActionRunNotFound() {
	testCases := []struct {
		name         string
		installID    string
		expectedCode int
	}{
		{
			name:         "non-existent install",
			installID:    domains.NewInstallID(),
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			action, appConfig, _ := s.setupActionWorkflowWithInstall()
			config := s.createActionConfigWithManualTrigger(action.ID, appConfig.ID)

			req := CreateInstallActionWorkflowRunRequest{
				ActionWorkFlowConfigID: config.ID,
			}

			path := fmt.Sprintf("/v1/installs/%s/actions/runs", tc.installID)
			rr := s.makeRequest(http.MethodPost, path, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			assert.Equal(s.T(), tc.expectedCode, rr.Code)
		})
	}
}

func (s *CreateInstallActionRunTestSuite) TestCreateActionRunCrossOrgIsolation() {
	// Create install in org1
	action1, appConfig1, install1 := s.setupActionWorkflowWithInstall()
	config1 := s.createActionConfigWithManualTrigger(action1.ID, appConfig1.ID)
	s.createInstallActionWorkflow(install1.ID, action1.ID)

	// Create second org with install
	ctx2 := context.Background()
	ctx2, acc2 := s.service.Seeder.EnsureAccount(ctx2, s.T())
	ctx2, org2 := s.service.Seeder.EnsureOrg(ctx2, s.T())

	app2 := &app.App{
		ID:          domains.NewAppID(),
		OrgID:       org2.ID,
		Name:        "org2-app",
		CreatedByID: acc2.ID,
	}
	ctx2 = cctx.SetAccountContext(ctx2, acc2)
	ctx2 = cctx.SetOrgContext(ctx2, org2)
	res := s.service.DB.WithContext(ctx2).Create(app2)
	require.NoError(s.T(), res.Error)

	install2 := &app.Install{
		ID:    domains.NewInstallID(),
		Name:  fmt.Sprintf("org2-install-%s", domains.NewInstallID()),
		AppID: app2.ID,
	}
	res = s.service.DB.WithContext(ctx2).
		Omit("app_config_id", "app_sandbox_config_id", "app_runner_config_id").
		Create(install2)
	require.NoError(s.T(), res.Error)

	// Try to create action run in org1 for install in org2
	req := CreateInstallActionWorkflowRunRequest{
		ActionWorkFlowConfigID: config1.ID,
	}

	path := fmt.Sprintf("/v1/installs/%s/actions/runs", install2.ID)
	rr := s.makeRequest(http.MethodPost, path, req)

	// Cross-org request returns 400 because the config's app_config_id
	// doesn't match the other org's install's app_config_id
	if rr.Code != http.StatusBadRequest {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

// Helper methods

func (s *CreateInstallActionRunTestSuite) setupActionWorkflowWithInstall() (*app.ActionWorkflow, *app.AppConfig, *app.Install) {
	action := s.createActionWorkflow(s.testApp.ID, fmt.Sprintf("test-action-%s", domains.NewActionWorkflowID()))
	appConfig := s.createAppConfig(s.testApp.ID)
	install := s.createInstall(s.testApp.ID, appConfig.ID)
	return action, appConfig, install
}

func (s *CreateInstallActionRunTestSuite) createActionWorkflow(appID, name string) *app.ActionWorkflow {
	action := &app.ActionWorkflow{
		ID:    domains.NewActionWorkflowID(),
		OrgID: s.testOrg.ID,
		AppID: appID,
		Name:  name,
	}
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	res := s.service.DB.WithContext(ctx).Create(action)
	require.NoError(s.T(), res.Error)
	return action
}

func (s *CreateInstallActionRunTestSuite) createAppConfig(appID string) *app.AppConfig {
	config := &app.AppConfig{
		ID:          domains.NewConfigID(),
		OrgID:       s.testOrg.ID,
		AppID:       appID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppConfigStatusActive,
	}
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	res := s.service.DB.WithContext(ctx).Create(config)
	require.NoError(s.T(), res.Error)
	return config
}

func (s *CreateInstallActionRunTestSuite) createInstall(appID, appConfigID string) *app.Install {
	install := &app.Install{
		ID:          domains.NewInstallID(),
		Name:        fmt.Sprintf("test-install-%s", domains.NewInstallID()),
		AppID:       appID,
		AppConfigID: appConfigID,
	}
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	ctx = cctx.SetOrgContext(ctx, s.testOrg)
	res := s.service.DB.WithContext(ctx).
		Omit("app_sandbox_config_id", "app_runner_config_id").
		Create(install)
	if res.Error != nil {
		s.T().Logf("FULL createInstall error: %+v", res.Error)
	}
	require.NoError(s.T(), res.Error)
	return install
}

func (s *CreateInstallActionRunTestSuite) createActionConfigWithManualTrigger(actionID, appConfigID string) *app.ActionWorkflowConfig {
	config := &app.ActionWorkflowConfig{
		ID:               domains.NewActionWorkflowConfigID(),
		OrgID:            s.testOrg.ID,
		AppID:            s.testApp.ID,
		AppConfigID:      appConfigID,
		ActionWorkflowID: actionID,
		Timeout:          5 * time.Minute,
	}
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	res := s.service.DB.WithContext(ctx).Create(config)
	require.NoError(s.T(), res.Error)

	// Create manual trigger
	trigger := &app.ActionWorkflowTriggerConfig{
		ID:                     domains.NewActionWorkflowTriggerConfigID(),
		OrgID:                  s.testOrg.ID,
		AppID:                  s.testApp.ID,
		AppConfigID:            appConfigID,
		ActionWorkflowConfigID: config.ID,
		Type:                   app.ActionWorkflowTriggerTypeManual,
	}
	res = s.service.DB.WithContext(ctx).Create(trigger)
	require.NoError(s.T(), res.Error)

	return config
}

func (s *CreateInstallActionRunTestSuite) createInstallActionWorkflow(installID, actionID string) *app.InstallActionWorkflow {
	installAction := &app.InstallActionWorkflow{
		ID:               domains.NewInstallActionWorkflowConfigID(),
		OrgID:            s.testOrg.ID,
		InstallID:        installID,
		ActionWorkflowID: actionID,
	}
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	res := s.service.DB.WithContext(ctx).Create(installAction)
	require.NoError(s.T(), res.Error)
	return installAction
}
