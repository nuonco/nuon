package service

import (
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

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type GetInstallCreateTargetTestService struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	MW          metrics.Writer
	AppsService *service
	Seeder      *testseed.Seeder
}

type GetInstallCreateTargetTestSuite struct {
	tests.BaseDBTestSuite

	fxApp         *fxtest.App
	service       GetInstallCreateTargetTestService
	router        *gin.Engine
	ctx           context.Context
	testOrg       *app.Org
	testAcc       *app.Account
	testApp       *app.App
	testAppConfig *app.AppConfig
}

func TestGetInstallCreateTargetSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetInstallCreateTargetTestSuite))
}

func (s *GetInstallCreateTargetTestSuite) SetupSuite() {
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

func (s *GetInstallCreateTargetTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
	s.testAppConfig = s.service.Seeder.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	require.NoError(s.T(), s.service.AppsService.RegisterPublicRoutes(s.router))
}

func (s *GetInstallCreateTargetTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *GetInstallCreateTargetTestSuite) seedBranch(name string, groups []app.AppBranchInstallGroup) (*app.AppBranch, *app.AppConfig) {
	branch := &app.AppBranch{AppID: s.testApp.ID, Name: name}
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Create(branch).Error)

	cfg := s.service.Seeder.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Model(cfg).
		Updates(map[string]any{"app_branch_id": branch.ID, "builds_pinned": true, "builds_locked": true}).Error)

	for i := range groups {
		groups[i].Order = i
	}
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Create(&app.AppBranchConfig{
		AppBranchID:   branch.ID,
		InstallGroups: groups,
	}).Error)

	return branch, cfg
}

func (s *GetInstallCreateTargetTestSuite) getTarget(appID, labelsQuery, appBranchID string) *httptest.ResponseRecorder {
	path := fmt.Sprintf("/v1/apps/%s/install-create-target", appID)
	if labelsQuery != "" || appBranchID != "" {
		path += "?"
		if labelsQuery != "" {
			path += "labels=" + labelsQuery
		}
		if appBranchID != "" {
			if labelsQuery != "" {
				path += "&"
			}
			path += "app_branch_id=" + appBranchID
		}
	}
	req, err := http.NewRequest(http.MethodGet, path, nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetInstallCreateTargetTestSuite) assertNothingCreated() {
	var installCount, connectionCount int64
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Model(&app.Install{}).Where(app.Install{AppID: s.testApp.ID}).Count(&installCount).Error)
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Model(&app.InstallAppBranchConnection{}).Count(&connectionCount).Error)
	assert.Zero(s.T(), installCount)
	assert.Zero(s.T(), connectionCount)
}

func (s *GetInstallCreateTargetTestSuite) TestGetInstallCreateTargetReturnsMatchedBranchAndGroup() {
	branch, cfg := s.seedBranch("prod", []app.AppBranchInstallGroup{{
		Name:          "prod",
		LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
	}})

	rr := s.getTarget(s.testApp.ID, "env:prod", "")
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var resp InstallCreateTargetResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), branch.ID, resp.AppBranchID)
	assert.Equal(s.T(), branch.Name, resp.AppBranchName)
	assert.Equal(s.T(), "prod", resp.AppBranchGroup)
	assert.Equal(s.T(), app.InstallAppBranchGroupAssignmentSourceLabels, resp.AppBranchGroupAssignmentSource)
	assert.Equal(s.T(), cfg.ID, resp.AppConfigID)

	s.assertNothingCreated()
}

func (s *GetInstallCreateTargetTestSuite) TestGetInstallCreateTargetReturnsFallbackConfigWhenNoBranchMatches() {
	rr := s.getTarget(s.testApp.ID, "", "")
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var resp InstallCreateTargetResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), s.testAppConfig.ID, resp.AppConfigID)
	assert.Empty(s.T(), resp.AppBranchID)

	s.assertNothingCreated()
}

func (s *GetInstallCreateTargetTestSuite) TestGetInstallCreateTargetReturnsSameAmbiguityErrorAsCreate() {
	branchA, _ := s.seedBranch("prod", []app.AppBranchInstallGroup{{
		Name:          "prod",
		LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
	}})
	branchB, _ := s.seedBranch("acme", []app.AppBranchInstallGroup{{
		Name:          "acme",
		LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"customer": "acme"}},
	}})

	rr := s.getTarget(s.testApp.ID, "env:prod,customer:acme", "")
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Contains(s.T(), rr.Body.String(), branchA.Name)
	assert.Contains(s.T(), rr.Body.String(), branchB.Name)

	s.assertNothingCreated()
}

func (s *GetInstallCreateTargetTestSuite) TestGetInstallCreateTargetReturnsNoBranchMatchError() {
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Delete(s.testAppConfig).Error)
	s.seedBranch("prod", []app.AppBranchInstallGroup{{
		Name:          "prod",
		LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
	}})

	rr := s.getTarget(s.testApp.ID, "", "")
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Contains(s.T(), rr.Body.String(), "has no app_branch and its labels")

	s.assertNothingCreated()
}

func (s *GetInstallCreateTargetTestSuite) TestGetInstallCreateTargetHonorsExplicitAppBranchID() {
	s.seedBranch("prod", []app.AppBranchInstallGroup{{
		Name:          "prod",
		LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
	}})
	branchB, cfgB := s.seedBranch("target", []app.AppBranchInstallGroup{{Name: "default", Default: true}})

	rr := s.getTarget(s.testApp.ID, "env:prod", branchB.ID)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var resp InstallCreateTargetResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), branchB.ID, resp.AppBranchID)
	assert.Equal(s.T(), cfgB.ID, resp.AppConfigID)

	s.assertNothingCreated()
}

func (s *GetInstallCreateTargetTestSuite) TestGetInstallCreateTargetAppliesAppDefaultLabelsWithoutQueryLabels() {
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Model(s.testApp).Update("default_labels", labels.Labels{"env": "prod"}).Error)
	branch, _ := s.seedBranch("prod", []app.AppBranchInstallGroup{{
		Name:          "prod",
		LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
	}})

	rr := s.getTarget(s.testApp.ID, "", "")
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var resp InstallCreateTargetResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), branch.ID, resp.AppBranchID)

	s.assertNothingCreated()
}

func (s *GetInstallCreateTargetTestSuite) TestGetInstallCreateTargetReturnsNotFoundForAnotherOrgsApp() {
	otherCtx, _ := s.service.Seeder.EnsureOrg(context.Background(), s.T())
	otherApp := s.service.Seeder.CreateApp(otherCtx, s.T())

	rr := s.getTarget(otherApp.ID, "", "")
	assert.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())
}
