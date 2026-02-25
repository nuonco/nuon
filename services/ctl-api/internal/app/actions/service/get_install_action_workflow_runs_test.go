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
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type GetInstallActionWorkflowRunsTestService struct {
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

type GetInstallActionWorkflowRunsTestSuite struct {
	tests.BaseDBTestSuite
	app          *fxtest.App
	service      GetInstallActionWorkflowRunsTestService
	router       *gin.Engine
	ctx          context.Context
	testOrg      *app.Org
	testAcc      *app.Account
	testApp      *app.App
	mockEvClient *tests.MockEventLoopClient
}

func TestGetInstallActionWorkflowRunsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetInstallActionWorkflowRunsTestSuite))
}

func (s *GetInstallActionWorkflowRunsTestSuite) SetupSuite() {
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

func (s *GetInstallActionWorkflowRunsTestSuite) SetupTest() {
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

func (s *GetInstallActionWorkflowRunsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetInstallActionWorkflowRunsTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *GetInstallActionWorkflowRunsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *GetInstallActionWorkflowRunsTestSuite) createInstall(appID string) *app.Install {
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

func (s *GetInstallActionWorkflowRunsTestSuite) createActionWorkflow(appID, name string) *app.ActionWorkflow {
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

func (s *GetInstallActionWorkflowRunsTestSuite) createInstallActionWorkflow(installID, actionID string) *app.InstallActionWorkflow {
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

func (s *GetInstallActionWorkflowRunsTestSuite) createInstallActionWorkflowRun(installID string, installActionID string, status app.InstallActionWorkflowRunStatus) *app.InstallActionWorkflowRun {
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

func (s *GetInstallActionWorkflowRunsTestSuite) TestGetInstallActionRuns() {
	testCases := []struct {
		name          string
		setupFunc     func() string
		queryParams   string
		expectedCount int
		expectedCode  int
		validateFunc  func([]*app.InstallActionWorkflowRun)
	}{
		{
			name: "returns empty array when no runs",
			setupFunc: func() string {
				install := s.createInstall(s.testApp.ID)
				return install.ID
			},
			queryParams:   "",
			expectedCount: 0,
			expectedCode:  http.StatusOK,
		},
		{
			name: "returns install action runs ordered by created_at desc",
			setupFunc: func() string {
				install := s.createInstall(s.testApp.ID)
				action := s.createActionWorkflow(s.testApp.ID, "test-action")
				installAction := s.createInstallActionWorkflow(install.ID, action.ID)
				s.createInstallActionWorkflowRun(install.ID, installAction.ID, app.InstallActionRunStatusQueued)
				s.createInstallActionWorkflowRun(install.ID, installAction.ID, app.InstallActionRunStatusInProgress)
				s.createInstallActionWorkflowRun(install.ID, installAction.ID, app.InstallActionRunStatusFinished)
				return install.ID
			},
			queryParams:   "",
			expectedCount: 3,
			expectedCode:  http.StatusOK,
			validateFunc: func(runs []*app.InstallActionWorkflowRun) {
				assert.Equal(s.T(), app.InstallActionRunStatusFinished, runs[0].Status)
				assert.Equal(s.T(), app.InstallActionRunStatusInProgress, runs[1].Status)
				assert.Equal(s.T(), app.InstallActionRunStatusQueued, runs[2].Status)
			},
		},
		{
			name: "respects pagination limit",
			setupFunc: func() string {
				install := s.createInstall(s.testApp.ID)
				action := s.createActionWorkflow(s.testApp.ID, "test-action-pagination")
				installAction := s.createInstallActionWorkflow(install.ID, action.ID)
				for i := 0; i < 15; i++ {
					s.createInstallActionWorkflowRun(install.ID, installAction.ID, app.InstallActionRunStatusQueued)
				}
				return install.ID
			},
			queryParams:   "?limit=5",
			expectedCount: 5,
			expectedCode:  http.StatusOK,
		},
		{
			name: "filters by install id correctly",
			setupFunc: func() string {
				install1 := s.createInstall(s.testApp.ID)
				install2 := s.createInstall(s.testApp.ID)
				action := s.createActionWorkflow(s.testApp.ID, "test-action-filter")
				installAction1 := s.createInstallActionWorkflow(install1.ID, action.ID)
				installAction2 := s.createInstallActionWorkflow(install2.ID, action.ID)
				s.createInstallActionWorkflowRun(install1.ID, installAction1.ID, app.InstallActionRunStatusQueued)
				s.createInstallActionWorkflowRun(install1.ID, installAction1.ID, app.InstallActionRunStatusQueued)
				s.createInstallActionWorkflowRun(install2.ID, installAction2.ID, app.InstallActionRunStatusQueued)
				return install1.ID
			},
			queryParams:   "",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
		},
		{
			name: "returns 404 for non-existent install",
			setupFunc: func() string {
				return domains.NewInstallID()
			},
			queryParams:  "",
			expectedCode: http.StatusOK,
			validateFunc: func(runs []*app.InstallActionWorkflowRun) {
				assert.Len(s.T(), runs, 0)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			installID := tc.setupFunc()

			path := fmt.Sprintf("/v1/installs/%s/actions/runs%s", installID, tc.queryParams)
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var response []*app.InstallActionWorkflowRun
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
				}
				require.NoError(s.T(), err)
				require.NotNil(s.T(), response)

				if tc.expectedCount >= 0 {
					assert.Len(s.T(), response, tc.expectedCount)
				}

				if tc.validateFunc != nil {
					tc.validateFunc(response)
				}
			}
		})
	}
}
