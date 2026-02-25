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

// GetAppActionConfigsTestService holds all fx-injected dependencies for get action configs tests.
type GetAppActionConfigsTestService struct {
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

// GetAppActionConfigsTestSuite is the testify suite for GetAppActionConfigs endpoint.
type GetAppActionConfigsTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      GetAppActionConfigsTestService
	router       *gin.Engine
	ctx          context.Context
	testOrg      *app.Org
	testAcc      *app.Account
	testApp      *app.App
	mockEvClient *tests.MockEventLoopClient
}

func TestGetAppActionConfigsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetAppActionConfigsTestSuite))
}

func (s *GetAppActionConfigsTestSuite) SetupSuite() {
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

func (s *GetAppActionConfigsTestSuite) SetupTest() {
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

func (s *GetAppActionConfigsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetAppActionConfigsTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *GetAppActionConfigsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *GetAppActionConfigsTestSuite) TestGetActionConfigsSuccess() {
	testCases := []struct {
		name          string
		setupFunc     func() string // Returns actionID
		queryParams   string
		expectedCount int
		expectedCode  int
		validateFunc  func([]*app.ActionWorkflowConfig)
	}{
		{
			name: "single config",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "single-config-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				s.createActionConfig(action.ID, appConfig.ID)
				return action.ID
			},
			queryParams:   "",
			expectedCount: 1,
			expectedCode:  http.StatusOK,
		},
		{
			name: "multiple configs",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "multi-config-action")
				appConfig1 := s.createAppConfig(s.testApp.ID)
				appConfig2 := s.createAppConfig(s.testApp.ID)
				appConfig3 := s.createAppConfig(s.testApp.ID)
				s.createActionConfig(action.ID, appConfig1.ID)
				time.Sleep(10 * time.Millisecond) // Ensure different timestamps
				s.createActionConfig(action.ID, appConfig2.ID)
				time.Sleep(10 * time.Millisecond)
				s.createActionConfig(action.ID, appConfig3.ID)
				return action.ID
			},
			queryParams:   "",
			expectedCount: 3,
			expectedCode:  http.StatusOK,
		},
		{
			name: "no configs",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "no-config-action")
				return action.ID
			},
			queryParams:   "",
			expectedCount: 0,
			expectedCode:  http.StatusOK,
		},
		{
			name: "with pagination",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "pagination-action")
				appConfig1 := s.createAppConfig(s.testApp.ID)
				appConfig2 := s.createAppConfig(s.testApp.ID)
				appConfig3 := s.createAppConfig(s.testApp.ID)
				s.createActionConfig(action.ID, appConfig1.ID)
				time.Sleep(10 * time.Millisecond)
				s.createActionConfig(action.ID, appConfig2.ID)
				time.Sleep(10 * time.Millisecond)
				s.createActionConfig(action.ID, appConfig3.ID)
				return action.ID
			},
			queryParams:   "?limit=2",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
		},
		{
			name: "configs with triggers and steps",
			setupFunc: func() string {
				action := s.createActionWorkflow(s.testApp.ID, "full-config-action")
				appConfig := s.createAppConfig(s.testApp.ID)
				config := s.createActionConfig(action.ID, appConfig.ID)
				s.createActionTrigger(config.ID, app.ActionWorkflowTriggerTypeManual)
				s.createActionStep(config.ID, "step1", 0)
				s.createActionStep(config.ID, "step2", 1)
				return action.ID
			},
			queryParams:   "",
			expectedCount: 1,
			expectedCode:  http.StatusOK,
			validateFunc: func(configs []*app.ActionWorkflowConfig) {
				assert.Len(s.T(), configs[0].Triggers, 1)
				assert.Len(s.T(), configs[0].Steps, 2)
				assert.Equal(s.T(), "step1", configs[0].Steps[0].Name)
				assert.Equal(s.T(), "step2", configs[0].Steps[1].Name)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			actionID := tc.setupFunc()

			path := fmt.Sprintf("/v1/apps/%s/actions/%s/configs%s", s.testApp.ID, actionID, tc.queryParams)
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			var configs []*app.ActionWorkflowConfig
			err := json.Unmarshal(rr.Body.Bytes(), &configs)
			require.NoError(s.T(), err)

			assert.Len(s.T(), configs, tc.expectedCount)

			if tc.validateFunc != nil {
				tc.validateFunc(configs)
			}
		})
	}
}

func (s *GetAppActionConfigsTestSuite) TestGetActionConfigsNotFound() {
	testCases := []struct {
		name         string
		actionID     string
		expectedCode int
	}{
		{
			name:         "non-existent action",
			actionID:     domains.NewActionWorkflowID(),
			expectedCode: http.StatusOK, // Returns empty array
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := fmt.Sprintf("/v1/apps/%s/actions/%s/configs", s.testApp.ID, tc.actionID)
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			assert.Equal(s.T(), tc.expectedCode, rr.Code)

			var configs []*app.ActionWorkflowConfig
			err := json.Unmarshal(rr.Body.Bytes(), &configs)
			require.NoError(s.T(), err)
			assert.Len(s.T(), configs, 0)
		})
	}
}

func (s *GetAppActionConfigsTestSuite) TestGetActionConfigsCrossOrgIsolation() {
	// Create action in org1
	action1 := s.createActionWorkflow(s.testApp.ID, "org1-action")
	appConfig1 := s.createAppConfig(s.testApp.ID)
	s.createActionConfig(action1.ID, appConfig1.ID)

	// Create second org with action
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

	// Org1 should not see org2's configs
	path := fmt.Sprintf("/v1/apps/%s/actions/%s/configs", s.testApp.ID, action1.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	var configs []*app.ActionWorkflowConfig
	err := json.Unmarshal(rr.Body.Bytes(), &configs)
	require.NoError(s.T(), err)

	// Should only see org1's config
	assert.Len(s.T(), configs, 1)
	assert.Equal(s.T(), s.testOrg.ID, configs[0].OrgID)
}

// Helper methods

func (s *GetAppActionConfigsTestSuite) createActionWorkflow(appID, name string) *app.ActionWorkflow {
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

func (s *GetAppActionConfigsTestSuite) createAppConfig(appID string) *app.AppConfig {
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

func (s *GetAppActionConfigsTestSuite) createActionConfig(actionID, appConfigID string) *app.ActionWorkflowConfig {
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

func (s *GetAppActionConfigsTestSuite) createActionTrigger(configID string, triggerType app.ActionWorkflowTriggerType) *app.ActionWorkflowTriggerConfig {
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

func (s *GetAppActionConfigsTestSuite) createActionStep(configID, name string, idx int) *app.ActionWorkflowStepConfig {
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
