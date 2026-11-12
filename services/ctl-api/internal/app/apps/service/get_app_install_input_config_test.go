package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// AppInstallInputConfigTestSuite tests the GetAppInstallInputConfig endpoint.
type AppInstallInputConfigTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service TestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
	testApp *app.App
	testCtx context.Context
}

func TestAppInstallInputConfigSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AppInstallInputConfigTestSuite))
}

func (s *AppInstallInputConfigTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AppInstallInputConfigTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	ctx, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
	s.testCtx = ctx
	s.testApp = s.service.Seeder.CreateApp(ctx, s.T())

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.AppsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AppInstallInputConfigTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AppInstallInputConfigTestSuite) makeRequest(path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(http.MethodGet, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AppInstallInputConfigTestSuite) errDescription(rr *httptest.ResponseRecorder) string {
	var resp stderr.ErrResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp), rr.Body.String())
	return resp.Description
}

// TestNoBranchReturnsAppsSyncConfigEvenWithNewerBranchConfig covers the
// customer-reported case: a branch has a newer config with extra inputs, but
// an unbranched install must still be pinned to the apps-sync config.
func (s *AppInstallInputConfigTestSuite) TestNoBranchReturnsAppsSyncConfigEvenWithNewerBranchConfig() {
	ctx := s.testCtx

	syncCfg := s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)

	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(branch).Error)
	branchCfg := s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Model(branchCfg).
		Updates(map[string]any{"app_branch_id": branch.ID, "builds_locked": true}).Error)

	rr := s.makeRequest("/v1/apps/" + s.testApp.ID + "/install-input-config")
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var resp models.AppAppInputConfig
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))

	var expected app.AppInputConfig
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: syncCfg.ID}).First(&expected).Error)
	assert.Equal(s.T(), expected.ID, resp.ID)
}

// TestWithAppBranchIDReturnsBranchConfig covers the branch-pinned path.
func (s *AppInstallInputConfigTestSuite) TestWithAppBranchIDReturnsBranchConfig() {
	ctx := s.testCtx

	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(branch).Error)
	branchCfg := s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Model(branchCfg).
		Updates(map[string]any{"app_branch_id": branch.ID, "builds_locked": true}).Error)

	rr := s.makeRequest("/v1/apps/" + s.testApp.ID + "/install-input-config?app_branch_id=" + branch.ID)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var resp models.AppAppInputConfig
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))

	var expected app.AppInputConfig
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: branchCfg.ID}).First(&expected).Error)
	assert.Equal(s.T(), expected.ID, resp.ID)
}

// TestBranchFromAnotherAppReturnsBadRequest covers an app_branch_id that
// belongs to a different app.
func (s *AppInstallInputConfigTestSuite) TestBranchFromAnotherAppReturnsBadRequest() {
	ctx := s.testCtx

	otherApp := s.service.Seeder.CreateApp(cctx.SetOrgIDContext(ctx, s.testOrg.ID), s.T())
	branch := &app.AppBranch{AppID: otherApp.ID, Name: "prod"}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(branch).Error)

	rr := s.makeRequest("/v1/apps/" + s.testApp.ID + "/install-input-config?app_branch_id=" + branch.ID)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
}

// TestNoAppsSyncConfigWithBranchesReturnsBadRequest covers an app whose only
// configs live on app branches.
func (s *AppInstallInputConfigTestSuite) TestNoAppsSyncConfigWithBranchesReturnsBadRequest() {
	ctx := s.testCtx

	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(branch).Error)
	branchCfg := s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Model(branchCfg).
		Updates(map[string]any{"app_branch_id": branch.ID, "builds_locked": true}).Error)

	rr := s.makeRequest("/v1/apps/" + s.testApp.ID + "/install-input-config")
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Contains(s.T(), s.errDescription(rr), "app branch")
}

// TestNoConfigAndNoBranchesReturnsBadRequest covers an app that has never
// been synced.
func (s *AppInstallInputConfigTestSuite) TestNoConfigAndNoBranchesReturnsBadRequest() {
	rr := s.makeRequest("/v1/apps/" + s.testApp.ID + "/install-input-config")
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), "No active app config found. Please sync your app configuration before creating an install.", s.errDescription(rr))
}

// TestInputLatestConfigStillReturnsNewestConfigOfAnyKind guards that
// input-latest-config keeps its own, unrelated behavior.
func (s *AppInstallInputConfigTestSuite) TestInputLatestConfigStillReturnsNewestConfigOfAnyKind() {
	ctx := s.testCtx

	s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)

	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(branch).Error)
	newestBranchCfg := s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Model(newestBranchCfg).
		Updates(map[string]any{"app_branch_id": branch.ID, "builds_locked": true}).Error)

	rr := s.makeRequest("/v1/apps/" + s.testApp.ID + "/input-latest-config")
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var resp models.AppAppInputConfig
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))

	var expected app.AppInputConfig
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: newestBranchCfg.ID}).First(&expected).Error)
	assert.Equal(s.T(), expected.ID, resp.ID)
}
