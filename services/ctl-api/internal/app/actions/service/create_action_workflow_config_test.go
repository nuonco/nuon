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

// CreateAppActionConfigTestService holds all fx-injected dependencies for create action config tests.
type CreateAppActionConfigTestService struct {
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

// CreateAppActionConfigTestSuite is the testify suite for CreateAppActionConfig endpoint.
type CreateAppActionConfigTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      CreateAppActionConfigTestService
	router       *gin.Engine
	ctx          context.Context
	testOrg      *app.Org
	testAcc      *app.Account
	testApp      *app.App
	mockEvClient *tests.MockEventLoopClient
}

func TestCreateAppActionConfigSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(CreateAppActionConfigTestSuite))
}

func (s *CreateAppActionConfigTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	// Create fake event loop client for testing
	s.mockEvClient = tests.NewMockEventLoopClient()

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			Mocks: &tests.TestMocks{MockEv: s.mockEvClient},

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

func (s *CreateAppActionConfigTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Reset mock before each test
	s.mockEvClient.Reset()

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

func (s *CreateAppActionConfigTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CreateAppActionConfigTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *CreateAppActionConfigTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *CreateAppActionConfigTestSuite) TestCreateActionConfigSuccess() {
	testCases := []struct {
		name         string
		setupFunc    func() (string, string, string) // Returns actionID, appConfigID, componentID
		requestFunc  func(actionID, appConfigID, componentID string) CreateActionWorkflowConfigRequest
		expectedCode int
		validateFunc func(*app.ActionWorkflowConfig)
	}{
		{
			name: "manual trigger only",
			setupFunc: func() (string, string, string) {
				action := s.createActionWorkflow(s.testApp.ID, "manual-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				return action.ID, appConfig.ID, ""
			},
			requestFunc: func(actionID, appConfigID, componentID string) CreateActionWorkflowConfigRequest {
				return CreateActionWorkflowConfigRequest{
					AppConfigID: appConfigID,
					Triggers: []CreateActionWorkflowConfigTriggerRequest{
						{Type: app.ActionWorkflowTriggerTypeManual},
					},
					Steps: []CreateActionWorkflowConfigStepRequest{
						{
							Name:           "test-step",
							InlineContents: "echo 'test'",
						},
					},
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(config *app.ActionWorkflowConfig) {
				assert.NotEmpty(s.T(), config.ID)
				assert.Len(s.T(), config.Triggers, 1)
				assert.Len(s.T(), config.Steps, 1)
				assert.Equal(s.T(), app.ActionWorkflowTriggerTypeManual, config.Triggers[0].Type)
			},
		},
		{
			name: "cron trigger",
			setupFunc: func() (string, string, string) {
				action := s.createActionWorkflow(s.testApp.ID, "cron-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				return action.ID, appConfig.ID, ""
			},
			requestFunc: func(actionID, appConfigID, componentID string) CreateActionWorkflowConfigRequest {
				return CreateActionWorkflowConfigRequest{
					AppConfigID: appConfigID,
					Triggers: []CreateActionWorkflowConfigTriggerRequest{
						{
							Type:         app.ActionWorkflowTriggerTypeCron,
							CronSchedule: "0 0 * * *",
						},
					},
					Steps: []CreateActionWorkflowConfigStepRequest{
						{
							Name:    "cron-step",
							Command: "echo 'scheduled'",
						},
					},
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(config *app.ActionWorkflowConfig) {
				assert.Len(s.T(), config.Triggers, 1)
				assert.Equal(s.T(), app.ActionWorkflowTriggerTypeCron, config.Triggers[0].Type)
				assert.Equal(s.T(), "0 0 * * *", config.Triggers[0].CronSchedule)
			},
		},
		{
			name: "component lifecycle trigger",
			setupFunc: func() (string, string, string) {
				action := s.createActionWorkflow(s.testApp.ID, "component-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				component := s.createComponent(s.testApp.ID, "test-comp")
				return action.ID, appConfig.ID, component.Name
			},
			requestFunc: func(actionID, appConfigID, componentName string) CreateActionWorkflowConfigRequest {
				return CreateActionWorkflowConfigRequest{
					AppConfigID: appConfigID,
					Triggers: []CreateActionWorkflowConfigTriggerRequest{
						{
							Type:          app.ActionWorkflowTriggerTypePostDeployComponent,
							ComponentName: componentName,
						},
					},
					Steps: []CreateActionWorkflowConfigStepRequest{
						{
							Name:           "component-step",
							InlineContents: "echo 'after deploy'",
						},
					},
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(config *app.ActionWorkflowConfig) {
				assert.Len(s.T(), config.Triggers, 1)
				assert.Equal(s.T(), app.ActionWorkflowTriggerTypePostDeployComponent, config.Triggers[0].Type)
				assert.NotEmpty(s.T(), config.Triggers[0].ComponentID.String)
			},
		},
		{
			name: "with timeout",
			setupFunc: func() (string, string, string) {
				action := s.createActionWorkflow(s.testApp.ID, "timeout-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				return action.ID, appConfig.ID, ""
			},
			requestFunc: func(actionID, appConfigID, componentID string) CreateActionWorkflowConfigRequest {
				return CreateActionWorkflowConfigRequest{
					AppConfigID: appConfigID,
					Timeout:     30 * time.Minute,
					Triggers: []CreateActionWorkflowConfigTriggerRequest{
						{Type: app.ActionWorkflowTriggerTypeManual},
					},
					Steps: []CreateActionWorkflowConfigStepRequest{
						{
							Name:           "timeout-step",
							InlineContents: "echo 'with timeout'",
						},
					},
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(config *app.ActionWorkflowConfig) {
				assert.Equal(s.T(), 30*time.Minute, config.Timeout)
			},
		},
		{
			name: "multiple steps",
			setupFunc: func() (string, string, string) {
				action := s.createActionWorkflow(s.testApp.ID, "multi-step-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				return action.ID, appConfig.ID, ""
			},
			requestFunc: func(actionID, appConfigID, componentID string) CreateActionWorkflowConfigRequest {
				return CreateActionWorkflowConfigRequest{
					AppConfigID: appConfigID,
					Triggers: []CreateActionWorkflowConfigTriggerRequest{
						{Type: app.ActionWorkflowTriggerTypeManual},
					},
					Steps: []CreateActionWorkflowConfigStepRequest{
						{
							Name:           "step1",
							InlineContents: "echo 'first'",
						},
						{
							Name:           "step2",
							InlineContents: "echo 'second'",
						},
						{
							Name:           "step3",
							InlineContents: "echo 'third'",
						},
					},
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(config *app.ActionWorkflowConfig) {
				assert.Len(s.T(), config.Steps, 3)
				assert.Equal(s.T(), "step1", config.Steps[0].Name)
				assert.Equal(s.T(), "step2", config.Steps[1].Name)
				assert.Equal(s.T(), "step3", config.Steps[2].Name)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock at the start of each test case
			s.mockEvClient.Reset()

			actionID, appConfigID, componentID := tc.setupFunc()

			req := tc.requestFunc(actionID, appConfigID, componentID)
			path := fmt.Sprintf("/v1/apps/%s/actions/%s/configs", s.testApp.ID, actionID)
			rr := s.makeRequest(http.MethodPost, path, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			var config app.ActionWorkflowConfig
			err := json.Unmarshal(rr.Body.Bytes(), &config)
			require.NoError(s.T(), err)

			// Verify signal was sent
			signals := s.mockEvClient.GetSignals()
			require.Len(s.T(), signals, 1)
			assert.Equal(s.T(), actionID, signals[0].ID)

			// Verify database state
			var dbConfig app.ActionWorkflowConfig
			res := s.service.DB.WithContext(s.ctx).
				Preload("Triggers").
				Preload("Steps").
				First(&dbConfig, "id = ?", config.ID)
			require.NoError(s.T(), res.Error)

			if tc.validateFunc != nil {
				tc.validateFunc(&dbConfig)
			}

			s.T().Cleanup(func() {
				s.service.DB.Unscoped().Where("id = ?", config.ID).Delete(&app.ActionWorkflowConfig{})
			})
		})
	}
}

func (s *CreateAppActionConfigTestSuite) TestCreateActionConfigValidation() {
	action := s.createActionWorkflow(s.testApp.ID, "validation-action")
	appConfig := s.createAppConfig(s.testApp.ID)

	testCases := []struct {
		name         string
		requestFunc  func() CreateActionWorkflowConfigRequest
		expectedCode int
	}{
		{
			name: "missing app_config_id",
			requestFunc: func() CreateActionWorkflowConfigRequest {
				return CreateActionWorkflowConfigRequest{
					Triggers: []CreateActionWorkflowConfigTriggerRequest{
						{Type: app.ActionWorkflowTriggerTypeManual},
					},
					Steps: []CreateActionWorkflowConfigStepRequest{
						{Name: "test", InlineContents: "echo 'test'"},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "missing triggers",
			requestFunc: func() CreateActionWorkflowConfigRequest {
				return CreateActionWorkflowConfigRequest{
					AppConfigID: appConfig.ID,
					Steps: []CreateActionWorkflowConfigStepRequest{
						{Name: "test", InlineContents: "echo 'test'"},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "missing steps",
			requestFunc: func() CreateActionWorkflowConfigRequest {
				return CreateActionWorkflowConfigRequest{
					AppConfigID: appConfig.ID,
					Triggers: []CreateActionWorkflowConfigTriggerRequest{
						{Type: app.ActionWorkflowTriggerTypeManual},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "timeout exceeds maximum",
			requestFunc: func() CreateActionWorkflowConfigRequest {
				return CreateActionWorkflowConfigRequest{
					AppConfigID: appConfig.ID,
					Timeout:     2 * time.Hour,
					Triggers: []CreateActionWorkflowConfigTriggerRequest{
						{Type: app.ActionWorkflowTriggerTypeManual},
					},
					Steps: []CreateActionWorkflowConfigStepRequest{
						{Name: "test", InlineContents: "echo 'test'"},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "multiple cron triggers",
			requestFunc: func() CreateActionWorkflowConfigRequest {
				return CreateActionWorkflowConfigRequest{
					AppConfigID: appConfig.ID,
					Triggers: []CreateActionWorkflowConfigTriggerRequest{
						{Type: app.ActionWorkflowTriggerTypeCron, CronSchedule: "0 0 * * *"},
						{Type: app.ActionWorkflowTriggerTypeCron, CronSchedule: "0 12 * * *"},
					},
					Steps: []CreateActionWorkflowConfigStepRequest{
						{Name: "test", InlineContents: "echo 'test'"},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "component trigger without component name",
			requestFunc: func() CreateActionWorkflowConfigRequest {
				return CreateActionWorkflowConfigRequest{
					AppConfigID: appConfig.ID,
					Triggers: []CreateActionWorkflowConfigTriggerRequest{
						{Type: app.ActionWorkflowTriggerTypePostDeployComponent},
					},
					Steps: []CreateActionWorkflowConfigStepRequest{
						{Name: "test", InlineContents: "echo 'test'"},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "step with both inline_contents and command",
			requestFunc: func() CreateActionWorkflowConfigRequest {
				return CreateActionWorkflowConfigRequest{
					AppConfigID: appConfig.ID,
					Triggers: []CreateActionWorkflowConfigTriggerRequest{
						{Type: app.ActionWorkflowTriggerTypeManual},
					},
					Steps: []CreateActionWorkflowConfigStepRequest{
						{
							Name:           "test",
							InlineContents: "echo 'inline'",
							Command:        "echo 'command'",
						},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "step with no execution method",
			requestFunc: func() CreateActionWorkflowConfigRequest {
				return CreateActionWorkflowConfigRequest{
					AppConfigID: appConfig.ID,
					Triggers: []CreateActionWorkflowConfigTriggerRequest{
						{Type: app.ActionWorkflowTriggerTypeManual},
					},
					Steps: []CreateActionWorkflowConfigStepRequest{
						{Name: "test"},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := tc.requestFunc()
			path := fmt.Sprintf("/v1/apps/%s/actions/%s/configs", s.testApp.ID, action.ID)
			rr := s.makeRequest(http.MethodPost, path, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			assert.Equal(s.T(), tc.expectedCode, rr.Code)
		})
	}
}

func (s *CreateAppActionConfigTestSuite) TestCreateActionConfigNotFound() {
	appConfig := s.createAppConfig(s.testApp.ID)

	testCases := []struct {
		name         string
		actionID     string
		expectedCode int
	}{
		{
			name:         "non-existent action",
			actionID:     domains.NewActionWorkflowID(),
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := CreateActionWorkflowConfigRequest{
				AppConfigID: appConfig.ID,
				Triggers: []CreateActionWorkflowConfigTriggerRequest{
					{Type: app.ActionWorkflowTriggerTypeManual},
				},
				Steps: []CreateActionWorkflowConfigStepRequest{
					{Name: "test", InlineContents: "echo 'test'"},
				},
			}

			path := fmt.Sprintf("/v1/apps/%s/actions/%s/configs", s.testApp.ID, tc.actionID)
			rr := s.makeRequest(http.MethodPost, path, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			assert.Equal(s.T(), tc.expectedCode, rr.Code)
		})
	}
}

func (s *CreateAppActionConfigTestSuite) TestCreateActionConfigCrossOrgIsolation() {
	// Create second org and account
	ctx2 := context.Background()
	ctx2, acc2 := s.service.Seeder.EnsureAccount(ctx2, s.T())
	ctx2, org2 := s.service.Seeder.EnsureOrg(ctx2, s.T())

	// Create action in org2
	action2 := &app.ActionWorkflow{
		ID:    domains.NewActionWorkflowID(),
		OrgID: org2.ID,
		AppID: s.testApp.ID,
		Name:  "cross-org-action",
	}
	ctx2 = cctx.SetAccountContext(ctx2, acc2)
	ctx2 = cctx.SetOrgIDContext(ctx2, org2.ID)
	res := s.service.DB.WithContext(ctx2).Create(action2)
	require.NoError(s.T(), res.Error)

	appConfig := s.createAppConfig(s.testApp.ID)

	req := CreateActionWorkflowConfigRequest{
		AppConfigID: appConfig.ID,
		Triggers: []CreateActionWorkflowConfigTriggerRequest{
			{Type: app.ActionWorkflowTriggerTypeManual},
		},
		Steps: []CreateActionWorkflowConfigStepRequest{
			{Name: "test", InlineContents: "echo 'test'"},
		},
	}

	// Try to create config in org1 for action in org2
	path := fmt.Sprintf("/v1/apps/%s/actions/%s/configs", s.testApp.ID, action2.ID)
	rr := s.makeRequest(http.MethodPost, path, req)

	if rr.Code != http.StatusNotFound {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}

// Helper methods

func (s *CreateAppActionConfigTestSuite) createActionWorkflow(appID, name string) *app.ActionWorkflow {
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

func (s *CreateAppActionConfigTestSuite) createAppConfig(appID string) *app.AppConfig {
	config := &app.AppConfig{
		ID:     domains.NewConfigID(),
		OrgID:  s.testOrg.ID,
		AppID:  appID,
		Status: app.AppConfigStatusActive,
	}
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	res := s.service.DB.WithContext(ctx).Create(config)
	require.NoError(s.T(), res.Error)
	return config
}

func (s *CreateAppActionConfigTestSuite) createComponent(appID, name string) *app.Component {
	component := &app.Component{
		ID:    domains.NewComponentID(),
		OrgID: s.testOrg.ID,
		AppID: appID,
		Name:  name,
	}
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	res := s.service.DB.WithContext(ctx).Create(component)
	require.NoError(s.T(), res.Error)
	return component
}
