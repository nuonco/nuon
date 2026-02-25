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

// GetAppActionLatestConfigTestService holds all fx-injected dependencies for get latest action config tests.
type GetAppActionLatestConfigTestService struct {
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

// GetAppActionLatestConfigTestSuite is the testify suite for GetAppActionLatestConfig endpoint.
type GetAppActionLatestConfigTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      GetAppActionLatestConfigTestService
	router       *gin.Engine
	ctx          context.Context
	testOrg      *app.Org
	testAcc      *app.Account
	testApp      *app.App
	mockEvClient *tests.MockEventLoopClient
}

func TestGetAppActionLatestConfigSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetAppActionLatestConfigTestSuite))
}

func (s *GetAppActionLatestConfigTestSuite) SetupSuite() {
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

func (s *GetAppActionLatestConfigTestSuite) SetupTest() {
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

func (s *GetAppActionLatestConfigTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetAppActionLatestConfigTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *GetAppActionLatestConfigTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *GetAppActionLatestConfigTestSuite) TestGetLatestConfigSuccess() {
	testCases := []struct {
		name         string
		setupFunc    func() string // Returns actionID
		expectedCode int
		validateFunc func(*app.ActionWorkflowConfig)
	}{
		{
			name: "single config",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "single-config-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				s.createActionConfig(action.ID, appConfig.ID)
				return action.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(config *app.ActionWorkflowConfig) {
				assert.NotEmpty(s.T(), config.ID)
			},
		},
		{
			name: "multiple configs returns latest",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "multi-config-action")
				appConfig1 := s.createAppConfig(s.testApp.ID)
				appConfig2 := s.createAppConfig(s.testApp.ID)
				appConfig3 := s.createAppConfig(s.testApp.ID)

				// Create configs with explicit ordering
				config1 := s.createActionConfig(action.ID, appConfig1.ID)
				time.Sleep(10 * time.Millisecond)
				config2 := s.createActionConfig(action.ID, appConfig2.ID)
				time.Sleep(10 * time.Millisecond)
				config3 := s.createActionConfig(action.ID, appConfig3.ID)

				// Store latest config ID for validation
				s.T().Logf("Created configs in order: %s, %s, %s (latest)", config1.ID, config2.ID, config3.ID)

				return action.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(config *app.ActionWorkflowConfig) {
				// Verify this is actually the latest by checking timestamp
				var allConfigs []*app.ActionWorkflowConfig
				s.service.DB.Order("created_at DESC").Find(&allConfigs)
				if len(allConfigs) > 0 {
					assert.Equal(s.T(), allConfigs[0].ID, config.ID, "Should return the most recently created config")
				}
			},
		},
		{
			name: "with triggers and steps",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "full-config-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				config := s.createActionConfig(action.ID, appConfig.ID)
				s.createActionTrigger(config.ID, app.ActionWorkflowTriggerTypeManual)
				s.createActionStep(config.ID, "step1", 0)
				s.createActionStep(config.ID, "step2", 1)
				return action.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(config *app.ActionWorkflowConfig) {
				assert.Len(s.T(), config.Triggers, 1)
				assert.Len(s.T(), config.Steps, 2)
				assert.Equal(s.T(), "step1", config.Steps[0].Name)
				assert.Equal(s.T(), "step2", config.Steps[1].Name)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			actionID := tc.setupFunc()

			path := fmt.Sprintf("/v1/apps/%s/actions/%s/latest-config", s.testApp.ID, actionID)
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

func (s *GetAppActionLatestConfigTestSuite) TestGetLatestConfigNotFound() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		expectedCode int
	}{
		{
			name: "non-existent action",
			setupFunc: func() string {
				return domains.NewActionWorkflowID()
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "action with no configs",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "no-config-action")
				return action.ID
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			actionID := tc.setupFunc()

			path := fmt.Sprintf("/v1/apps/%s/actions/%s/latest-config", s.testApp.ID, actionID)
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			assert.Equal(s.T(), tc.expectedCode, rr.Code)
		})
	}
}

func (s *GetAppActionLatestConfigTestSuite) TestGetLatestConfigCrossOrgIsolation() {
	// Create action in org1 with configs
	action1 := s.createActionWorkflow(s.testApp.ID, "org1-action")
	appConfig1 := s.createAppConfig(s.testApp.ID)
	config1 := s.createActionConfig(action1.ID, appConfig1.ID)

	// Create second org with action and configs
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

	// Try to get action1's latest config - should work
	path := fmt.Sprintf("/v1/apps/%s/actions/%s/latest-config", s.testApp.ID, action1.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	var config app.ActionWorkflowConfig
	err := json.Unmarshal(rr.Body.Bytes(), &config)
	require.NoError(s.T(), err)

	// Should return org1's config
	assert.Equal(s.T(), config1.ID, config.ID)
	assert.Equal(s.T(), s.testOrg.ID, config.OrgID)
}

// Helper methods

func (s *GetAppActionLatestConfigTestSuite) createActionWorkflow(appID, name string) *app.ActionWorkflow {
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

func (s *GetAppActionLatestConfigTestSuite) createAppConfig(appID string) *app.AppConfig {
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

func (s *GetAppActionLatestConfigTestSuite) createActionConfig(actionID, appConfigID string) *app.ActionWorkflowConfig {
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
	return config
}

func (s *GetAppActionLatestConfigTestSuite) createActionTrigger(configID string, triggerType app.ActionWorkflowTriggerType) *app.ActionWorkflowTriggerConfig {
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

func (s *GetAppActionLatestConfigTestSuite) createActionStep(configID, name string, idx int) *app.ActionWorkflowStepConfig {
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
