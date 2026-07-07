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
	cctx "github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type GetInstallActionWorkflowsTestService struct {
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

type GetInstallActionWorkflowsTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service GetInstallActionWorkflowsTestService
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
	testApp *app.App
}

func TestGetInstallActionWorkflowsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetInstallActionWorkflowsTestSuite))
}

func (s *GetInstallActionWorkflowsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)
	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T:               s.T(),
			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *GetInstallActionWorkflowsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.ActionsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetInstallActionWorkflowsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetInstallActionWorkflowsTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *GetInstallActionWorkflowsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

// createInstall creates a full app config and then an install with app_config_id set.
// This is required so the currentAppConfigActionFilter in the handler has a config ID to match against.
func (s *GetInstallActionWorkflowsTestSuite) createInstall(appID string) (*app.Install, *app.AppConfig) {
	appCfg := s.service.Seeder.CreateAppConfig(s.ctx, s.T(), appID)
	install := s.service.Seeder.CreateInstall(s.ctx, s.T(), s.testApp)
	return install, appCfg
}

func (s *GetInstallActionWorkflowsTestSuite) createActionWorkflow(appID, name string) *app.ActionWorkflow {
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

// createActionWorkflowConfig ties an action to an app config, making it visible to the filter.
func (s *GetInstallActionWorkflowsTestSuite) createActionWorkflowConfig(appConfigID, actionID string) *app.ActionWorkflowConfig {
	return s.service.Seeder.CreateActionWorkflowConfig(s.ctx, s.T(), s.testApp.ID, appConfigID, actionID)
}

func (s *GetInstallActionWorkflowsTestSuite) createInstallActionWorkflow(installID, actionID string) *app.InstallActionWorkflow {
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

func (s *GetInstallActionWorkflowsTestSuite) TestGetInstallActions() {
	testCases := []struct {
		name          string
		setupFunc     func() string
		queryParams   string
		expectedCount int
		expectedCode  int
		validateFunc  func([]app.InstallActionWorkflow)
	}{
		{
			name: "returns empty array when no actions",
			setupFunc: func() string {
				install, _ := s.createInstall(s.testApp.ID)
				return install.ID
			},
			queryParams:   "",
			expectedCount: 0,
			expectedCode:  http.StatusOK,
		},
		{
			name: "returns install actions with current app config",
			setupFunc: func() string {
				install, appCfg := s.createInstall(s.testApp.ID)
				action1 := s.createActionWorkflow(s.testApp.ID, "action-1")
				action2 := s.createActionWorkflow(s.testApp.ID, "action-2")
				s.createActionWorkflowConfig(appCfg.ID, action1.ID)
				s.createActionWorkflowConfig(appCfg.ID, action2.ID)
				s.createInstallActionWorkflow(install.ID, action1.ID)
				s.createInstallActionWorkflow(install.ID, action2.ID)
				return install.ID
			},
			queryParams:   "",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(iaws []app.InstallActionWorkflow) {
				require.NotNil(s.T(), iaws[0].ActionWorkflow)
				require.NotNil(s.T(), iaws[1].ActionWorkflow)
			},
		},
		{
			name: "respects pagination",
			setupFunc: func() string {
				install, appCfg := s.createInstall(s.testApp.ID)
				for i := 0; i < 15; i++ {
					action := s.createActionWorkflow(s.testApp.ID, fmt.Sprintf("action-pagination-%d", i))
					s.createActionWorkflowConfig(appCfg.ID, action.ID)
					s.createInstallActionWorkflow(install.ID, action.ID)
				}
				return install.ID
			},
			queryParams:   "?limit=5",
			expectedCount: 5,
			expectedCode:  http.StatusOK,
		},
		{
			name: "hides action whose config belongs to a different app config",
			setupFunc: func() string {
				install, currentAppCfg := s.createInstall(s.testApp.ID)

				// Action visible in current app config.
				visibleAction := s.createActionWorkflow(s.testApp.ID, "current-action")
				s.createActionWorkflowConfig(currentAppCfg.ID, visibleAction.ID)
				s.createInstallActionWorkflow(install.ID, visibleAction.ID)

				// Action that was removed in a later sync: its ActionWorkflowConfig
				// points to a different (older) app config, not the install's current one.
				olderAppCfg := s.service.Seeder.CreateBareAppConfig(s.ctx, s.T(), s.testApp.ID)
				hiddenAction := s.createActionWorkflow(s.testApp.ID, "old-action")
				s.createActionWorkflowConfig(olderAppCfg.ID, hiddenAction.ID)
				s.createInstallActionWorkflow(install.ID, hiddenAction.ID)

				return install.ID
			},
			queryParams:   "",
			expectedCount: 1,
			expectedCode:  http.StatusOK,
			validateFunc: func(iaws []app.InstallActionWorkflow) {
				require.Equal(s.T(), "current-action", iaws[0].ActionWorkflow.Name)
			},
		},
		{
			name: "not found for non-existent install",
			setupFunc: func() string {
				return "inst-nonexistent123456789"
			},
			queryParams:  "",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			installID := tc.setupFunc()

			path := fmt.Sprintf("/v1/installs/%s/actions%s", installID, tc.queryParams)
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var response []app.InstallActionWorkflow
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
				}
				require.NoError(s.T(), err)
				require.NotNil(s.T(), response)

				if tc.expectedCount >= 0 {
					assert.Len(s.T(), response, tc.expectedCount)
				}

				if tc.validateFunc != nil && len(response) > 0 {
					tc.validateFunc(response)
				}
			}
		})
	}
}
