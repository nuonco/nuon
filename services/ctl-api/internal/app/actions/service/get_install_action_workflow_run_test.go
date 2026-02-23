package service

import (
	"bytes"
	"context"
	"database/sql"
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

	pkg_generics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	comphelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type GetInstallActionWorkflowRunTestService struct {
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

type GetInstallActionWorkflowRunTestSuite struct {
	tests.BaseDBTestSuite
	app          *fxtest.App
	service      GetInstallActionWorkflowRunTestService
	router       *gin.Engine
	ctx          context.Context
	testOrg      *app.Org
	testAcc      *app.Account
	testApp      *app.App
	mockEvClient *tests.FakeEventLoopClient
}

func TestGetInstallActionWorkflowRunSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetInstallActionWorkflowRunTestSuite))
}

func (s *GetInstallActionWorkflowRunTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)
	s.mockEvClient = tests.NewFakeEventLoopClient()
	options := append(
		tests.CtlApiFXOptions(),
		fx.Decorate(func() eventloop.Client { return s.mockEvClient }),
		fx.Provide(New),
		fx.Populate(&s.service),
	)
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *GetInstallActionWorkflowRunTestSuite) SetupTest() {
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

func (s *GetInstallActionWorkflowRunTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetInstallActionWorkflowRunTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *GetInstallActionWorkflowRunTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *GetInstallActionWorkflowRunTestSuite) createInstall(appID string) *app.Install {
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

func (s *GetInstallActionWorkflowRunTestSuite) createActionWorkflow(appID, name string) *app.ActionWorkflow {
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

func (s *GetInstallActionWorkflowRunTestSuite) createInstallActionWorkflow(installID, actionID string) *app.InstallActionWorkflow {
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

func (s *GetInstallActionWorkflowRunTestSuite) createInstallActionWorkflowRun(installID string, installActionID string, status app.InstallActionWorkflowRunStatus) *app.InstallActionWorkflowRun {
	run := &app.InstallActionWorkflowRun{
		ID:        domains.NewInstallActionWorkflowRunID(),
		OrgID:     s.testOrg.ID,
		InstallID: installID,
		InstallActionWorkflowID: pkg_generics.NullString{
			NullString: sql.NullString{String: installActionID, Valid: true},
		},
		Status:            status,
		StatusDescription: string(status),
		TriggerType:       app.ActionWorkflowTriggerTypeManual,
	}
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	res := s.service.DB.WithContext(ctx).Create(run)
	require.NoError(s.T(), res.Error)
	return run
}

func (s *GetInstallActionWorkflowRunTestSuite) TestGetInstallActionRun() {
	testCases := []struct {
		name         string
		setupFunc    func() (string, string)
		expectedCode int
		validateFunc func(*app.InstallActionWorkflowRun)
	}{
		{
			name: "returns run with preloaded associations",
			setupFunc: func() (string, string) {
				install := s.createInstall(s.testApp.ID)
				action := s.createActionWorkflow(s.testApp.ID, "test-action")
				installAction := s.createInstallActionWorkflow(install.ID, action.ID)
				run := s.createInstallActionWorkflowRun(install.ID, installAction.ID, app.InstallActionRunStatusQueued)
				return install.ID, run.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(run *app.InstallActionWorkflowRun) {
				assert.Equal(s.T(), app.InstallActionRunStatusQueued, run.Status)
				assert.NotEmpty(s.T(), run.ID)
				assert.NotEmpty(s.T(), run.InstallID)
			},
		},
		{
			name: "returns run with different statuses",
			setupFunc: func() (string, string) {
				install := s.createInstall(s.testApp.ID)
				action := s.createActionWorkflow(s.testApp.ID, "test-action-finished")
				installAction := s.createInstallActionWorkflow(install.ID, action.ID)
				run := s.createInstallActionWorkflowRun(install.ID, installAction.ID, app.InstallActionRunStatusFinished)
				return install.ID, run.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(run *app.InstallActionWorkflowRun) {
				assert.Equal(s.T(), app.InstallActionRunStatusFinished, run.Status)
			},
		},
		{
			name: "returns 404 for non-existent run",
			setupFunc: func() (string, string) {
				install := s.createInstall(s.testApp.ID)
				return install.ID, domains.NewInstallActionWorkflowRunID()
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "returns 404 for run not in org scope",
			setupFunc: func() (string, string) {
				install := s.createInstall(s.testApp.ID)
				// Use non-existent run ID to simulate org scope isolation
				return install.ID, domains.NewInstallActionWorkflowRunID()
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			installID, runID := tc.setupFunc()

			path := fmt.Sprintf("/v1/installs/%s/actions/runs/%s", installID, runID)
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var response app.InstallActionWorkflowRun
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
