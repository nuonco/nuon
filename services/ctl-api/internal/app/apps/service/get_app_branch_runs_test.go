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
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type GetAppBranchRunsTestService struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	MW          metrics.Writer
	AppsService *service
	Seeder      *testseed.Seeder
}

type GetAppBranchRunsTestSuite struct {
	tests.BaseDBTestSuite

	fxApp   *fxtest.App
	service GetAppBranchRunsTestService
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
	testApp *app.App
}

func TestGetAppBranchRunsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetAppBranchRunsTestSuite))
}

func (s *GetAppBranchRunsTestSuite) SetupSuite() {
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

func (s *GetAppBranchRunsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	require.NoError(s.T(), s.service.AppsService.RegisterPublicRoutes(s.router))
}

func (s *GetAppBranchRunsTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *GetAppBranchRunsTestSuite) createBranch() *app.AppBranch {
	branch := &app.AppBranch{
		ID:          domains.NewAppBranchID(),
		OrgID:       s.testOrg.ID,
		AppID:       s.testApp.ID,
		CreatedByID: s.testAcc.ID,
		Name:        fmt.Sprintf("runs-branch-%d", time.Now().UnixNano()),
		ManagedBy:   app.AppBranchManagedByManually,
	}
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Create(branch).Error)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.AppBranch{}, "id = ?", branch.ID)
	})
	return branch
}

func (s *GetAppBranchRunsTestSuite) seedRun(branchID string, workflowType app.WorkflowType, status app.Status, opts ...testseed.WorkflowOption) *app.Workflow {
	opts = append([]testseed.WorkflowOption{
		testseed.WithWorkflowOwnerType("app_branches"),
		testseed.WithWorkflowStatus(app.NewCompositeStatus(s.ctx, status)),
	}, opts...)

	workflow := s.service.Seeder.CreateWorkflow(s.ctx, s.T(), branchID, workflowType, opts...)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Workflow{}, "id = ?", workflow.ID)
	})
	return workflow
}

func (s *GetAppBranchRunsTestSuite) fetchRuns(branchID, query string) ([]app.Workflow, *httptest.ResponseRecorder) {
	path := fmt.Sprintf("/v1/apps/%s/branches/%s/runs%s", s.testApp.ID, branchID, query)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var runs []app.Workflow
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &runs))
	return runs, rr
}

func runIDs(runs []app.Workflow) []string {
	ids := make([]string, 0, len(runs))
	for _, run := range runs {
		ids = append(ids, run.ID)
	}
	return ids
}

func (s *GetAppBranchRunsTestSuite) TestFilterByStatus() {
	branch := s.createBranch()
	failed := s.seedRun(branch.ID, app.WorkflowTypeAppBranchesRun, app.StatusError)
	succeeded := s.seedRun(branch.ID, app.WorkflowTypeAppBranchesRun, app.StatusSuccess)
	pending := s.seedRun(branch.ID, app.WorkflowTypeAppBranchesConfigRepoUpdate, app.StatusPending)

	s.Run("single status returns only that status", func() {
		runs, _ := s.fetchRuns(branch.ID, "?status=error")
		require.Equal(s.T(), []string{failed.ID}, runIDs(runs))
	})

	s.Run("several statuses return the union", func() {
		runs, _ := s.fetchRuns(branch.ID, "?status=error,success")
		require.ElementsMatch(s.T(), []string{failed.ID, succeeded.ID}, runIDs(runs))
	})

	s.Run("unset status returns everything", func() {
		runs, _ := s.fetchRuns(branch.ID, "")
		require.ElementsMatch(s.T(), []string{failed.ID, succeeded.ID, pending.ID}, runIDs(runs))
	})

	s.Run("empty status is the same as unset", func() {
		runs, _ := s.fetchRuns(branch.ID, "?status=%20,")
		require.ElementsMatch(s.T(), []string{failed.ID, succeeded.ID, pending.ID}, runIDs(runs))
	})

	s.Run("unknown status returns an empty list", func() {
		runs, _ := s.fetchRuns(branch.ID, "?status=not-a-real-status")
		require.Empty(s.T(), runs)
	})
}

func (s *GetAppBranchRunsTestSuite) TestFilterByType() {
	branch := s.createBranch()
	manual := s.seedRun(branch.ID, app.WorkflowTypeAppBranchesRun, app.StatusSuccess)
	configRepo := s.seedRun(branch.ID, app.WorkflowTypeAppBranchesConfigRepoUpdate, app.StatusSuccess)
	componentRepo := s.seedRun(branch.ID, app.WorkflowTypeAppBranchesComponentRepoUpdate, app.StatusSuccess)

	s.Run("single type returns only that type", func() {
		runs, _ := s.fetchRuns(branch.ID, "?type=app_branches_manual_update")
		require.Equal(s.T(), []string{manual.ID}, runIDs(runs))
	})

	s.Run("several types return the union", func() {
		runs, _ := s.fetchRuns(branch.ID, "?type=app_branches_config_repo_update,app_branches_component_repo_update")
		require.ElementsMatch(s.T(), []string{configRepo.ID, componentRepo.ID}, runIDs(runs))
	})

	s.Run("unknown type returns an empty list", func() {
		runs, _ := s.fetchRuns(branch.ID, "?type=not_a_real_type")
		require.Empty(s.T(), runs)
	})
}

func (s *GetAppBranchRunsTestSuite) TestSearch() {
	branch := s.createBranch()
	configRepo := s.seedRun(branch.ID, app.WorkflowTypeAppBranchesConfigRepoUpdate, app.StatusSuccess)
	componentRepo := s.seedRun(branch.ID, app.WorkflowTypeAppBranchesComponentRepoUpdate, app.StatusSuccess)

	s.Run("tokens are and-ed against the run title", func() {
		runs, _ := s.fetchRuns(branch.ID, "?q=config+repo")
		require.Equal(s.T(), []string{configRepo.ID}, runIDs(runs))
	})

	s.Run("a run id matches", func() {
		runs, _ := s.fetchRuns(branch.ID, fmt.Sprintf("?q=%s", componentRepo.ID))
		require.Equal(s.T(), []string{componentRepo.ID}, runIDs(runs))
	})

	s.Run("no match returns an empty list", func() {
		runs, _ := s.fetchRuns(branch.ID, "?q=nothing-matches-this")
		require.Empty(s.T(), runs)
	})
}

func (s *GetAppBranchRunsTestSuite) TestFilterByCreatedAt() {
	branch := s.createBranch()
	now := time.Now().UTC()
	old := s.seedRun(branch.ID, app.WorkflowTypeAppBranchesRun, app.StatusSuccess,
		testseed.WithWorkflowCreatedAt(now.Add(-30*24*time.Hour)))
	recent := s.seedRun(branch.ID, app.WorkflowTypeAppBranchesRun, app.StatusSuccess,
		testseed.WithWorkflowCreatedAt(now.Add(-1*time.Hour)))

	s.Run("created_at_gte excludes older runs", func() {
		runs, _ := s.fetchRuns(branch.ID, fmt.Sprintf("?created_at_gte=%s", now.Add(-24*time.Hour).Format(time.RFC3339)))
		require.Equal(s.T(), []string{recent.ID}, runIDs(runs))
	})

	s.Run("created_at_lte excludes newer runs", func() {
		runs, _ := s.fetchRuns(branch.ID, fmt.Sprintf("?created_at_lte=%s", now.Add(-24*time.Hour).Format(time.RFC3339)))
		require.Equal(s.T(), []string{old.ID}, runIDs(runs))
	})

	s.Run("both bounds intersect", func() {
		query := fmt.Sprintf("?created_at_gte=%s&created_at_lte=%s",
			now.Add(-2*time.Hour).Format(time.RFC3339),
			now.Format(time.RFC3339))
		runs, _ := s.fetchRuns(branch.ID, query)
		require.Equal(s.T(), []string{recent.ID}, runIDs(runs))
	})

	s.Run("an unparseable bound is a user error", func() {
		path := fmt.Sprintf("/v1/apps/%s/branches/%s/runs?created_at_gte=yesterday", s.testApp.ID, branch.ID)
		req, err := http.NewRequest(http.MethodGet, path, nil)
		require.NoError(s.T(), err)

		rr := httptest.NewRecorder()
		s.router.ServeHTTP(rr, req)
		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	})
}

func (s *GetAppBranchRunsTestSuite) TestPlanOnlyUnchanged() {
	branch := s.createBranch()
	preview := s.seedRun(branch.ID, app.WorkflowTypeAppBranchesRun, app.StatusSuccess,
		testseed.WithWorkflowPlanOnly(true))
	rollout := s.seedRun(branch.ID, app.WorkflowTypeAppBranchesRun, app.StatusSuccess)

	s.Run("default includes preview runs", func() {
		runs, _ := s.fetchRuns(branch.ID, "")
		require.ElementsMatch(s.T(), []string{preview.ID, rollout.ID}, runIDs(runs))
	})

	s.Run("planonly=false excludes preview runs", func() {
		runs, _ := s.fetchRuns(branch.ID, "?planonly=false")
		require.Equal(s.T(), []string{rollout.ID}, runIDs(runs))
	})

	s.Run("planonly combines with the other filters", func() {
		runs, _ := s.fetchRuns(branch.ID, "?planonly=false&status=success")
		require.Equal(s.T(), []string{rollout.ID}, runIDs(runs))
	})
}

func (s *GetAppBranchRunsTestSuite) TestFiltersIntersect() {
	branch := s.createBranch()
	failedManual := s.seedRun(branch.ID, app.WorkflowTypeAppBranchesRun, app.StatusError)
	s.seedRun(branch.ID, app.WorkflowTypeAppBranchesRun, app.StatusSuccess)
	s.seedRun(branch.ID, app.WorkflowTypeAppBranchesConfigRepoUpdate, app.StatusError)

	runs, _ := s.fetchRuns(branch.ID, "?status=error&type=app_branches_manual_update")
	require.Equal(s.T(), []string{failedManual.ID}, runIDs(runs))
}

func (s *GetAppBranchRunsTestSuite) TestFilteredPagination() {
	branch := s.createBranch()

	failedIDs := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		failedIDs = append(failedIDs, s.seedRun(branch.ID, app.WorkflowTypeAppBranchesRun, app.StatusError).ID)
		s.seedRun(branch.ID, app.WorkflowTypeAppBranchesRun, app.StatusSuccess)
	}

	firstPage, firstRR := s.fetchRuns(branch.ID, "?status=error&limit=2")
	require.Len(s.T(), firstPage, 2)
	require.Equal(s.T(), "true", firstRR.Header().Get("X-Nuon-Page-Next"))

	secondPage, secondRR := s.fetchRuns(branch.ID, "?status=error&limit=2&offset=2")
	require.Len(s.T(), secondPage, 1)
	require.Equal(s.T(), "false", secondRR.Header().Get("X-Nuon-Page-Next"))

	require.ElementsMatch(s.T(), failedIDs, append(runIDs(firstPage), runIDs(secondPage)...))
}

func (s *GetAppBranchRunsTestSuite) TestEmpty() {
	branch := s.createBranch()

	runs, _ := s.fetchRuns(branch.ID, "")
	require.Empty(s.T(), runs)
}
