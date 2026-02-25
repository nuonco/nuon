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

// GetAppActionConfigTestService holds all fx-injected dependencies for get action config tests.
type GetAppActionConfigTestService struct {
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

// GetAppActionConfigTestSuite is the testify suite for GetAppActionConfig endpoint.
type GetAppActionConfigTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      GetAppActionConfigTestService
	router       *gin.Engine
	ctx          context.Context
	testOrg      *app.Org
	testAcc      *app.Account
	testApp      *app.App
	mockEvClient *tests.MockEventLoopClient
}

func TestGetAppActionConfigSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetAppActionConfigTestSuite))
}

func (s *GetAppActionConfigTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	s.mockEvClient = tests.NewMockEventLoopClient()

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			Mocks: &tests.TestMocks{MockEv: s.mockEvClient},

			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.service.DB)
}

func (s *GetAppActionConfigTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.mockEvClient.Reset()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.ActionsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetAppActionConfigTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetAppActionConfigTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *GetAppActionConfigTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *GetAppActionConfigTestSuite) TestGetActionConfigSuccess() {
	testCases := []struct {
		name         string
		setupFunc    func() string // Returns configID
		expectedCode int
		validateFunc func(*app.ActionWorkflowConfig)
	}{
		{
			name: "basic config",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "basic-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				config := s.createActionConfig(action.ID, appConfig.ID)
				return config.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(config *app.ActionWorkflowConfig) {
				assert.NotEmpty(s.T(), config.ID)
				assert.Equal(s.T(), s.testOrg.ID, config.OrgID)
				assert.Equal(s.T(), s.testApp.ID, config.AppID)
			},
		},
		{
			name: "config with triggers",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "trigger-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				config := s.createActionConfig(action.ID, appConfig.ID)
				s.createActionTrigger(config.ID, app.ActionWorkflowTriggerTypeManual)
				s.createActionTrigger(config.ID, app.ActionWorkflowTriggerTypeCron)
				return config.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(config *app.ActionWorkflowConfig) {
				assert.Len(s.T(), config.Triggers, 2)
			},
		},
		{
			name: "config with steps",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "steps-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				config := s.createActionConfig(action.ID, appConfig.ID)
				s.createActionStep(config.ID, "step1", 0)
				s.createActionStep(config.ID, "step2", 1)
				s.createActionStep(config.ID, "step3", 2)
				return config.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(config *app.ActionWorkflowConfig) {
				assert.Len(s.T(), config.Steps, 3)
				// Verify steps are ordered by idx
				assert.Equal(s.T(), "step1", config.Steps[0].Name)
				assert.Equal(s.T(), "step2", config.Steps[1].Name)
				assert.Equal(s.T(), "step3", config.Steps[2].Name)
			},
		},
		{
			name: "config with triggers and steps",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "full-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				config := s.createActionConfig(action.ID, appConfig.ID)
				s.createActionTrigger(config.ID, app.ActionWorkflowTriggerTypeManual)
				s.createActionStep(config.ID, "step1", 0)
				s.createActionStep(config.ID, "step2", 1)
				return config.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(config *app.ActionWorkflowConfig) {
				assert.Len(s.T(), config.Triggers, 1)
				assert.Len(s.T(), config.Steps, 2)
				assert.Equal(s.T(), app.ActionWorkflowTriggerTypeManual, config.Triggers[0].Type)
			},
		},
		{
			name: "config with timeout",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "timeout-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				config := s.createActionConfigWithTimeout(action.ID, appConfig.ID, 30*time.Minute)
				return config.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(config *app.ActionWorkflowConfig) {
				assert.Equal(s.T(), 30*time.Minute, config.Timeout)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			configID := tc.setupFunc()

			path := fmt.Sprintf("/v1/apps/%s/actions/configs/%s", s.testApp.ID, configID)
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			var config app.ActionWorkflowConfig
			err := json.Unmarshal(rr.Body.Bytes(), &config)
			require.NoError(s.T(), err)

			if tc.validateFunc != nil {
				tc.validateFunc(&config)
			}
		})
	}
}

func (s *GetAppActionConfigTestSuite) TestGetActionConfigNotFound() {
	testCases := []struct {
		name         string
		configID     string
		expectedCode int
	}{
		{
			name:         "non-existent config",
			configID:     domains.NewActionWorkflowConfigID(),
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := fmt.Sprintf("/v1/apps/%s/actions/configs/%s", s.testApp.ID, tc.configID)
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			assert.Equal(s.T(), tc.expectedCode, rr.Code)
		})
	}
}

func (s *GetAppActionConfigTestSuite) TestGetActionConfigCrossOrgIsolation() {
	// Create config in org1
	action1 := s.createActionWorkflow(s.testApp.ID, "org1-action")
	appConfig1 := s.createAppConfig(s.testApp.ID)
	config1 := s.createActionConfig(action1.ID, appConfig1.ID)

	// Create second org with config
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
	ctx2 = cctx.SetOrgIDContext(ctx2, org2.ID)
	res := s.service.DB.WithContext(ctx2).Create(app2)
	require.NoError(s.T(), res.Error)

	action2 := &app.ActionWorkflow{
		ID:    domains.NewActionWorkflowID(),
		OrgID: org2.ID,
		AppID: app2.ID,
		Name:  "org2-action",
	}
	res = s.service.DB.WithContext(ctx2).Create(action2)
	require.NoError(s.T(), res.Error)

	appConfig2 := &app.AppConfig{
		ID:     domains.NewConfigID(),
		OrgID:  org2.ID,
		AppID:  app2.ID,
		Status: app.AppConfigStatusActive,
	}
	res = s.service.DB.WithContext(ctx2).Create(appConfig2)
	require.NoError(s.T(), res.Error)

	config2 := &app.ActionWorkflowConfig{
		ID:               domains.NewActionWorkflowConfigID(),
		OrgID:            org2.ID,
		AppID:            app2.ID,
		AppConfigID:      appConfig2.ID,
		ActionWorkflowID: action2.ID,
		Timeout:          5 * time.Minute,
	}
	res = s.service.DB.WithContext(ctx2).Create(config2)
	require.NoError(s.T(), res.Error)

	// Org1 should be able to get config1
	path1 := fmt.Sprintf("/v1/apps/%s/actions/configs/%s", s.testApp.ID, config1.ID)
	rr1 := s.makeRequest(http.MethodGet, path1, nil)
	require.Equal(s.T(), http.StatusOK, rr1.Code)

	var retrievedConfig app.ActionWorkflowConfig
	err := json.Unmarshal(rr1.Body.Bytes(), &retrievedConfig)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), config1.ID, retrievedConfig.ID)

	// Org1 should NOT be able to get config2 (different org)
	path2 := fmt.Sprintf("/v1/apps/%s/actions/configs/%s", s.testApp.ID, config2.ID)
	rr2 := s.makeRequest(http.MethodGet, path2, nil)
	assert.Equal(s.T(), http.StatusNotFound, rr2.Code)
}

// Helper methods

func (s *GetAppActionConfigTestSuite) createActionWorkflow(appID, name string) *app.ActionWorkflow {
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

func (s *GetAppActionConfigTestSuite) createAppConfig(appID string) *app.AppConfig {
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

func (s *GetAppActionConfigTestSuite) createActionConfig(actionID, appConfigID string) *app.ActionWorkflowConfig {
	return s.createActionConfigWithTimeout(actionID, appConfigID, 5*time.Minute)
}

func (s *GetAppActionConfigTestSuite) createActionConfigWithTimeout(actionID, appConfigID string, timeout time.Duration) *app.ActionWorkflowConfig {
	config := &app.ActionWorkflowConfig{
		ID:               domains.NewActionWorkflowConfigID(),
		OrgID:            s.testOrg.ID,
		AppID:            s.testApp.ID,
		AppConfigID:      appConfigID,
		ActionWorkflowID: actionID,
		Timeout:          timeout,
	}
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	res := s.service.DB.WithContext(ctx).Create(config)
	require.NoError(s.T(), res.Error)
	return config
}

func (s *GetAppActionConfigTestSuite) createActionTrigger(configID string, triggerType app.ActionWorkflowTriggerType) *app.ActionWorkflowTriggerConfig {
	// Get the parent config to extract AppConfigID
	var config app.ActionWorkflowConfig
	res := s.service.DB.First(&config, "id = ?", configID)
	require.NoError(s.T(), res.Error)

	trigger := &app.ActionWorkflowTriggerConfig{
		ID:                     domains.NewActionWorkflowTriggerConfigID(),
		OrgID:                  s.testOrg.ID,
		AppID:                  s.testApp.ID,
		AppConfigID:            config.AppConfigID,
		ActionWorkflowConfigID: configID,
		Type:                   triggerType,
	}
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	res = s.service.DB.WithContext(ctx).Create(trigger)
	require.NoError(s.T(), res.Error)
	return trigger
}

func (s *GetAppActionConfigTestSuite) createActionStep(configID, name string, idx int) *app.ActionWorkflowStepConfig {
	// Get the parent config to extract AppConfigID
	var config app.ActionWorkflowConfig
	res := s.service.DB.First(&config, "id = ?", configID)
	require.NoError(s.T(), res.Error)

	step := &app.ActionWorkflowStepConfig{
		ID:                     domains.NewActionWorkflowStepConfigID(),
		OrgID:                  s.testOrg.ID,
		AppID:                  s.testApp.ID,
		AppConfigID:            config.AppConfigID,
		ActionWorkflowConfigID: configID,
		Name:                   name,
		Idx:                    idx,
		InlineContents:         "echo 'test'",
	}
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	res = s.service.DB.WithContext(ctx).Create(step)
	require.NoError(s.T(), res.Error)
	return step
}
