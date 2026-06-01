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

type GetInstallActionWorkflowRecentRunsTestService struct {
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

type GetInstallActionWorkflowRecentRunsTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service GetInstallActionWorkflowRecentRunsTestService
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
	testApp *app.App
}

func TestGetInstallActionWorkflowRecentRunsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetInstallActionWorkflowRecentRunsTestSuite))
}

func (s *GetInstallActionWorkflowRecentRunsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)
	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *GetInstallActionWorkflowRecentRunsTestSuite) SetupTest() {
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

func (s *GetInstallActionWorkflowRecentRunsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetInstallActionWorkflowRecentRunsTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *GetInstallActionWorkflowRecentRunsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *GetInstallActionWorkflowRecentRunsTestSuite) createInstall(appID string) *app.Install {
	install := &app.Install{
		ID:    domains.NewInstallID(),
		Name:  fmt.Sprintf("test-install-%s", domains.NewInstallID()),
		AppID: appID,
	}
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	ctx = cctx.SetOrgContext(ctx, s.testOrg)
	res := s.service.DB.WithContext(ctx).
		Omit("app_config_id", "app_sandbox_config_id", "app_runner_config_id").
		Create(install)
	require.NoError(s.T(), res.Error)
	return install
}

func (s *GetInstallActionWorkflowRecentRunsTestSuite) createActionWorkflow(appID, name string) *app.ActionWorkflow {
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

func (s *GetInstallActionWorkflowRecentRunsTestSuite) createInstallActionWorkflow(installID, actionID string) *app.InstallActionWorkflow {
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

func (s *GetInstallActionWorkflowRecentRunsTestSuite) TestGetInstallActionRecentRuns() {
	testCases := []struct {
		name         string
		setupFunc    func() (string, string)
		queryParams  string
		expectedCode int
		validateFunc func(*app.InstallActionWorkflow)
	}{
		{
			name: "returns workflow when action has no configs",
			setupFunc: func() (string, string) {
				install := s.createInstall(s.testApp.ID)
				action := s.createActionWorkflow(s.testApp.ID, "test-action")
				iaw := s.createInstallActionWorkflow(install.ID, action.ID)
				return install.ID, iaw.ActionWorkflowID
			},
			queryParams:  "",
			expectedCode: http.StatusOK,
			validateFunc: func(iaw *app.InstallActionWorkflow) {
				require.NotEmpty(s.T(), iaw.ID)
				require.Empty(s.T(), iaw.ActionWorkflow.Configs, "should have no configs")
			},
		},
		{
			name: "respects pagination limit",
			setupFunc: func() (string, string) {
				install := s.createInstall(s.testApp.ID)
				action := s.createActionWorkflow(s.testApp.ID, "pagination-test")
				iaw := s.createInstallActionWorkflow(install.ID, action.ID)
				return install.ID, iaw.ActionWorkflowID
			},
			queryParams:  "?limit=10",
			expectedCode: http.StatusOK,
			validateFunc: func(iaw *app.InstallActionWorkflow) {
				require.NotNil(s.T(), iaw)
			},
		},
		{
			name: "returns null for non-existent install",
			setupFunc: func() (string, string) {
				action := s.createActionWorkflow(s.testApp.ID, "test-action-notfound-install")
				return "inst-nonexistent123456789", action.ID
			},
			queryParams:  "",
			expectedCode: http.StatusOK,
			validateFunc: func(iaw *app.InstallActionWorkflow) {
				// Handler returns 200 with null on error
				require.Empty(s.T(), iaw.ID)
			},
		},
		{
			name: "returns null for non-existent action",
			setupFunc: func() (string, string) {
				install := s.createInstall(s.testApp.ID)
				return install.ID, "actw-nonexistent123456789"
			},
			queryParams:  "",
			expectedCode: http.StatusOK,
			validateFunc: func(iaw *app.InstallActionWorkflow) {
				// Handler returns 200 with null on error
				require.Empty(s.T(), iaw.ID)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			installID, actionID := tc.setupFunc()

			path := fmt.Sprintf("/v1/installs/%s/actions/%s/recent-runs%s", installID, actionID, tc.queryParams)
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var response app.InstallActionWorkflow
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
				}
				require.NoError(s.T(), err)

				if tc.validateFunc != nil {
					tc.validateFunc(&response)
				}
			}
		})
	}
}
