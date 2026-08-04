package service

import (
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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type GetAppBranchesTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	MW              metrics.Writer
	VcsHelpers      *vcshelpers.Helpers
	AppsHelpers     *appshelpers.Helpers
	InstallsHelpers *installshelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	AppsService     *service
	Seeder          *testseed.Seeder
}

type GetAppBranchesTestSuite struct {
	tests.BaseDBTestSuite

	fxApp   *fxtest.App
	service GetAppBranchesTestService
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
	testApp *app.App
}

func TestGetAppBranchesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetAppBranchesTestSuite))
}

func (s *GetAppBranchesTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *GetAppBranchesTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.AppsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetAppBranchesTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *GetAppBranchesTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *GetAppBranchesTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetAppBranchesTestSuite) createBranch(name string) *app.AppBranch {
	branch := &app.AppBranch{
		ID:          domains.NewAppBranchID(),
		OrgID:       s.testOrg.ID,
		AppID:       s.testApp.ID,
		CreatedByID: s.testAcc.ID,
		Name:        name,
		ManagedBy:   app.AppBranchManagedByManually,
	}
	err := s.service.DB.WithContext(s.ctx).Create(branch).Error
	require.NoError(s.T(), err)
	return branch
}

func (s *GetAppBranchesTestSuite) createBranchConfig(branchID string) *app.AppBranchConfig {
	cfg := &app.AppBranchConfig{
		ID:          domains.NewAppBranchConfigID(),
		OrgID:       s.testOrg.ID,
		AppBranchID: branchID,
		CreatedByID: s.testAcc.ID,
	}
	err := s.service.DB.WithContext(s.ctx).Create(cfg).Error
	require.NoError(s.T(), err)
	return cfg
}

func (s *GetAppBranchesTestSuite) createRun(branchID, branchConfigID string, createdAt time.Time) *app.AppBranchRun {
	run := &app.AppBranchRun{
		ID:                domains.NewAppBranchRunID(),
		OrgID:             s.testOrg.ID,
		AppBranchID:       branchID,
		AppBranchConfigID: branchConfigID,
		CreatedByID:       s.testAcc.ID,
		CreatedAt:         createdAt,
		Status:            "pending",
		RunType:           app.AppBranchRunTypeManual,
	}
	err := s.service.DB.WithContext(s.ctx).Create(run).Error
	require.NoError(s.T(), err)
	return run
}

func (s *GetAppBranchesTestSuite) createWorkflowForBranch(branchID string) *app.Workflow {
	wf := &app.Workflow{
		OwnerID:        branchID,
		OwnerType:      "app_branches",
		Type:           app.WorkflowTypeAppBranchesRun,
		Status:         app.NewCompositeStatus(s.ctx, app.StatusPending),
		ApprovalOption: app.InstallApprovalOptionPrompt,
	}
	err := s.service.DB.WithContext(s.ctx).Create(wf).Error
	require.NoError(s.T(), err)
	return wf
}

type rawBranchResponse struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	LatestRun *rawBranchRunInline `json:"latest_run"`
	Configs   []rawBranchConfig   `json:"configs"`
}

type rawBranchRunInline struct {
	ID               string `json:"id"`
	AppBranchID      string `json:"app_branch_id"`
	WorkflowID       string `json:"workflow_id"`
	AwaitingApproval bool   `json:"awaiting_approval"`
}

type rawBranchConfig struct {
	ID            string                  `json:"id"`
	InstallGroups []rawBranchInstallGroup `json:"install_groups"`
}

type rawBranchInstallGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *GetAppBranchesTestSuite) fetchBranches() []rawBranchResponse {
	path := fmt.Sprintf("/v1/apps/%s/branches", s.testApp.ID)
	rr := s.makeRequest(http.MethodGet, path)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var branches []rawBranchResponse
	err := json.Unmarshal(rr.Body.Bytes(), &branches)
	if err != nil {
		s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
	}
	require.NoError(s.T(), err)
	return branches
}

func (s *GetAppBranchesTestSuite) TestBranchWithNoRuns() {
	branch := s.createBranch("no-runs-branch")
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.AppBranch{}, "id = ?", branch.ID)
	})

	branches := s.fetchBranches()

	var found *rawBranchResponse
	for i := range branches {
		if branches[i].ID == branch.ID {
			found = &branches[i]
			break
		}
	}
	require.NotNil(s.T(), found, "branch not found in response")
	require.Nil(s.T(), found.LatestRun, "expected latest_run to be nil when no runs exist")
}

func (s *GetAppBranchesTestSuite) TestBranchLatestRunIsNewest() {
	branch := s.createBranch("multi-run-branch")
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.AppBranch{}, "id = ?", branch.ID)
	})

	cfg := s.createBranchConfig(branch.ID)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.AppBranchConfig{}, "id = ?", cfg.ID)
	})

	now := time.Now().UTC()
	olderRun := s.createRun(branch.ID, cfg.ID, now.Add(-10*time.Minute))
	newerRun := s.createRun(branch.ID, cfg.ID, now)

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.AppBranchRun{}, "id = ?", olderRun.ID)
		s.service.DB.Unscoped().Delete(&app.AppBranchRun{}, "id = ?", newerRun.ID)
	})

	branches := s.fetchBranches()

	var found *rawBranchResponse
	for i := range branches {
		if branches[i].ID == branch.ID {
			found = &branches[i]
			break
		}
	}
	require.NotNil(s.T(), found, "branch not found in response")
	require.NotNil(s.T(), found.LatestRun, "expected latest_run to be present")
	require.Equal(s.T(), newerRun.ID, found.LatestRun.ID, "expected latest_run to be the newest run")

	var dbRun app.AppBranchRun
	err := s.service.DB.Where(app.AppBranchRun{ID: found.LatestRun.ID}).First(&dbRun).Error
	require.NoError(s.T(), err)
	require.Equal(s.T(), newerRun.ID, dbRun.ID)
}

func (s *GetAppBranchesTestSuite) TestAwaitingApproval() {
	testCases := []struct {
		name           string
		addResponse    bool
		expectAwaiting bool
	}{
		{
			name:           "run with unanswered approval awaits approval",
			addResponse:    false,
			expectAwaiting: true,
		},
		{
			name:           "run with answered approval is not awaiting",
			addResponse:    true,
			expectAwaiting: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			branch := s.createBranch(fmt.Sprintf("approval-branch-%v", tc.addResponse))
			s.T().Cleanup(func() {
				s.service.DB.Unscoped().Delete(&app.AppBranch{}, "id = ?", branch.ID)
			})

			cfg := s.createBranchConfig(branch.ID)
			s.T().Cleanup(func() {
				s.service.DB.Unscoped().Delete(&app.AppBranchConfig{}, "id = ?", cfg.ID)
			})

			wf := s.createWorkflowForBranch(branch.ID)
			s.T().Cleanup(func() {
				s.service.DB.Unscoped().Delete(&app.Workflow{}, "id = ?", wf.ID)
			})

			wfID := wf.ID
			run := &app.AppBranchRun{
				ID:                domains.NewAppBranchRunID(),
				OrgID:             s.testOrg.ID,
				AppBranchID:       branch.ID,
				AppBranchConfigID: cfg.ID,
				CreatedByID:       s.testAcc.ID,
				CreatedAt:         time.Now().UTC(),
				Status:            "running",
				RunType:           app.AppBranchRunTypeManual,
				WorkflowID:        &wfID,
			}
			err := s.service.DB.WithContext(s.ctx).Create(run).Error
			require.NoError(s.T(), err)
			s.T().Cleanup(func() {
				s.service.DB.Unscoped().Delete(&app.AppBranchRun{}, "id = ?", run.ID)
			})

			step := s.service.Seeder.CreateWorkflowStep(
				s.ctx,
				s.T(),
				wf.ID,
				testseed.WithStepStatus(app.NewCompositeStatus(s.ctx, app.AwaitingApproval)),
			)
			err = s.service.DB.WithContext(s.ctx).
				Model(&app.WorkflowStep{}).
				Where(app.WorkflowStep{ID: step.ID}).
				Update("execution_type", app.WorkflowStepExecutionTypeApproval).Error
			require.NoError(s.T(), err)
			s.T().Cleanup(func() {
				s.service.DB.Unscoped().Delete(&app.WorkflowStep{}, "id = ?", step.ID)
			})

			approval := s.service.Seeder.CreateWorkflowStepApproval(
				s.ctx,
				s.T(),
				step.ID,
				app.AppBranchPlanApprovalType,
				"plan output",
			)
			s.T().Cleanup(func() {
				s.service.DB.Unscoped().Delete(&app.WorkflowStepApproval{}, "id = ?", approval.ID)
			})

			if tc.addResponse {
				response := &app.WorkflowStepApprovalResponse{
					InstallWorkflowStepApprovalID: approval.ID,
					OrgID:                         s.testOrg.ID,
					CreatedByID:                   s.testAcc.ID,
					Type:                          app.WorkflowStepApprovalResponseTypeApprove,
				}
				err = s.service.DB.WithContext(s.ctx).Create(response).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.WorkflowStepApprovalResponse{}, "id = ?", response.ID)
				})
			}

			branches := s.fetchBranches()

			var found *rawBranchResponse
			for i := range branches {
				if branches[i].ID == branch.ID {
					found = &branches[i]
					break
				}
			}
			require.NotNil(s.T(), found, "branch not found in response")
			require.NotNil(s.T(), found.LatestRun, "expected latest_run to be present")
			require.Equal(s.T(), tc.expectAwaiting, found.LatestRun.AwaitingApproval,
				"awaiting_approval mismatch for case: %s", tc.name)
		})
	}
}

func (s *GetAppBranchesTestSuite) TestInstallGroupsPreloaded() {
	branch := s.createBranch("install-group-branch")
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.AppBranch{}, "id = ?", branch.ID)
	})

	cfg := s.createBranchConfig(branch.ID)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.AppBranchConfig{}, "id = ?", cfg.ID)
	})

	installGroup := &app.AppBranchInstallGroup{
		ID:                domains.NewAppBranchInstallGroupID(),
		OrgID:             s.testOrg.ID,
		AppBranchConfigID: cfg.ID,
		CreatedByID:       s.testAcc.ID,
		Name:              "test-install-group",
		Order:             1,
	}
	err := s.service.DB.WithContext(s.ctx).Create(installGroup).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.AppBranchInstallGroup{}, "id = ?", installGroup.ID)
	})

	branches := s.fetchBranches()

	var found *rawBranchResponse
	for i := range branches {
		if branches[i].ID == branch.ID {
			found = &branches[i]
			break
		}
	}
	require.NotNil(s.T(), found, "branch not found in response")
	require.Len(s.T(), found.Configs, 1, "expected one config on the branch")
	require.Len(s.T(), found.Configs[0].InstallGroups, 1, "expected install_groups to be preloaded")
	require.Equal(s.T(), installGroup.ID, found.Configs[0].InstallGroups[0].ID)
	require.Equal(s.T(), "test-install-group", found.Configs[0].InstallGroups[0].Name)

	var dbGroup app.AppBranchInstallGroup
	err = s.service.DB.Where(app.AppBranchInstallGroup{ID: installGroup.ID}).First(&dbGroup).Error
	require.NoError(s.T(), err)
	require.Equal(s.T(), installGroup.ID, dbGroup.ID)
}
