package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	pkgconfig "github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	componentsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/components/service"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type appConfigIntermediateDeps struct {
	fx.In

	DB       *gorm.DB `name:"psql"`
	L        *zap.Logger
	Seeder   *testseed.Seeder
	Services []api.Service `group:"services"`
}

type AppConfigIntermediateTestSuite struct {
	tests.BaseDBTestSuite

	fxApp       *fxtest.App
	deps        appConfigIntermediateDeps
	appsService *service
	router      *gin.Engine
	ctx         context.Context
	testOrg     *app.Org
	testAcc     *app.Account
	testApp     *app.App
}

func TestAppConfigIntermediateSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AppConfigIntermediateTestSuite))
}

func (s *AppConfigIntermediateTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Provide(api.AsService(componentsservice.New)),
		fx.Populate(&s.deps, &s.appsService),
	)

	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *AppConfigIntermediateTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *AppConfigIntermediateTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	s.ctx = context.Background()
	s.ctx, s.testAcc = s.deps.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.deps.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.deps.Seeder.CreateApp(s.ctx, s.T())

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.deps.L,
		DB:      s.deps.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	require.NoError(s.T(), s.appsService.RegisterPublicRoutes(s.router))
	for _, svc := range s.deps.Services {
		require.NoError(s.T(), svc.RegisterPublicRoutes(s.router))
	}
}

func (s *AppConfigIntermediateTestSuite) get(path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(http.MethodGet, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AppConfigIntermediateTestSuite) seedBranch(name string) *app.AppBranch {
	branch := &app.AppBranch{AppID: s.testApp.ID, Name: name}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)
	return branch
}

func (s *AppConfigIntermediateTestSuite) seedComponent(name string) *app.Component {
	component := &app.Component{
		AppID: s.testApp.ID,
		Name:  name,
		Type:  app.ComponentTypeTerraformModule,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(component).Error)
	return component
}

func (s *AppConfigIntermediateTestSuite) seedAction(name string) *app.ActionWorkflow {
	action := &app.ActionWorkflow{AppID: s.testApp.ID, Name: name}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(action).Error)
	return action
}

func (s *AppConfigIntermediateTestSuite) seedRunbook(name string) *app.Runbook {
	runbook := &app.Runbook{AppID: s.testApp.ID, Name: name}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(runbook).Error)
	return runbook
}

type seedConfigOpts struct {
	branchID     string
	status       app.AppConfigStatus
	createdAt    time.Time
	blob         string
	noBlob       bool
	componentIDs []string
	actionIDs    []string
	runbookIDs   []string
	orgID        string
	appID        string
}

func (s *AppConfigIntermediateTestSuite) seedConfig(opts seedConfigOpts) *app.AppConfig {
	status := opts.status
	if status == "" {
		status = app.AppConfigStatusActive
	}
	createdAt := opts.createdAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	appID := opts.appID
	if appID == "" {
		appID = s.testApp.ID
	}

	cfg := &app.AppConfig{
		AppID:              appID,
		CreatedAt:          createdAt,
		Status:             status,
		StatusV2:           app.NewCompositeStatus(s.ctx, app.Status(status)),
		StatusDescription:  string(status),
		CLIVersion:         "development",
		Checksum:           "sha256:test-checksum",
		ComponentIDs:       pq.StringArray(opts.componentIDs),
		ActionIDs:          pq.StringArray(opts.actionIDs),
		RunbookIDs:         pq.StringArray(opts.runbookIDs),
		IntermediateConfig: &blobstore.Blob{},
	}
	if opts.orgID != "" {
		cfg.OrgID = opts.orgID
	}
	if opts.branchID != "" {
		require.NoError(s.T(), cfg.AppBranchID.Scan(opts.branchID))
	}
	if !opts.noBlob {
		blob := opts.blob
		if blob == "" {
			blob = s.marshalConfig(&pkgconfig.AppConfig{Version: "v2"})
		}
		cfg.IntermediateConfig.Set(blob)
	}

	blobCtx := blobstore.WithBlobService(s.ctx, s.appsService.blobSvc)
	require.NoError(s.T(), s.deps.DB.WithContext(blobCtx).Create(cfg).Error)
	return cfg
}

func (s *AppConfigIntermediateTestSuite) marshalConfig(cfg *pkgconfig.AppConfig) string {
	raw, err := json.Marshal(cfg)
	require.NoError(s.T(), err)
	return string(raw)
}

func (s *AppConfigIntermediateTestSuite) decode(rr *httptest.ResponseRecorder) app.AppConfigIntermediate {
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var response app.AppConfigIntermediate
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &response))
	return response
}

func (s *AppConfigIntermediateTestSuite) resourceIDs(
	resources []app.AppConfigResource,
	kind app.AppConfigResourceKind,
) []string {
	ids := []string{}
	for _, resource := range resources {
		if resource.Kind == kind {
			ids = append(ids, resource.ID)
		}
	}
	return ids
}

func (s *AppConfigIntermediateTestSuite) branchPath(branchID string) string {
	return fmt.Sprintf("/v1/apps/%s/branches/%s/intermediate-config", s.testApp.ID, branchID)
}

func (s *AppConfigIntermediateTestSuite) configPath(configID string) string {
	return fmt.Sprintf("/v1/apps/%s/configs/%s/intermediate", s.testApp.ID, configID)
}

func (s *AppConfigIntermediateTestSuite) TestConfigRouteReturnsParsedConfig() {
	component := s.seedComponent("api")
	cfg := s.seedConfig(seedConfigOpts{
		blob: s.marshalConfig(&pkgconfig.AppConfig{
			Version:     "v2",
			DisplayName: "Example platform",
			Components:  pkgconfig.ComponentList{{Name: "api"}},
		}),
		componentIDs: []string{component.ID},
	})

	rr := s.get(s.configPath(cfg.ID))
	response := s.decode(rr)

	require.Equal(s.T(), cfg.ID, response.ConfigID)
	require.Equal(s.T(), s.testApp.ID, response.AppID)
	require.Equal(s.T(), cfg.Checksum, response.Checksum)
	require.NotNil(s.T(), response.Config)
	require.Equal(s.T(), "Example platform", response.Config.DisplayName)
	require.NotEmpty(s.T(), rr.Header().Get("ETag"))
	require.Contains(s.T(), rr.Header().Get("Cache-Control"), "private, max-age=")
}

func (s *AppConfigIntermediateTestSuite) TestBranchRouteReturnsParsedConfig() {
	branch := s.seedBranch("main")
	component := s.seedComponent("api")
	cfg := s.seedConfig(seedConfigOpts{
		branchID: branch.ID,
		blob: s.marshalConfig(&pkgconfig.AppConfig{
			Version:    "v2",
			Components: pkgconfig.ComponentList{{Name: "api"}},
		}),
		componentIDs: []string{component.ID},
	})

	response := s.decode(s.get(s.branchPath(branch.ID)))

	require.Equal(s.T(), cfg.ID, response.ConfigID)
	require.Equal(s.T(), branch.ID, response.AppBranchID)
	require.Equal(s.T(), []string{component.ID}, s.resourceIDs(response.Resources, app.AppConfigResourceKindComponent))
	require.Empty(s.T(), response.Config.Branches)
}

func (s *AppConfigIntermediateTestSuite) TestBranchRouteMatchesComponentsEndpoint() {
	branch := s.seedBranch("main")
	other := s.seedBranch("staging")

	mine := s.seedComponent("api")
	alsoMine := s.seedComponent("worker")
	theirs := s.seedComponent("experimental")

	s.seedConfig(seedConfigOpts{
		branchID: branch.ID,
		blob: s.marshalConfig(&pkgconfig.AppConfig{
			Version:    "v2",
			Components: pkgconfig.ComponentList{{Name: "api"}, {Name: "worker"}},
		}),
		componentIDs: []string{mine.ID, alsoMine.ID},
	})
	s.seedConfig(seedConfigOpts{
		branchID:     other.ID,
		createdAt:    time.Now().UTC().Add(time.Hour),
		blob:         s.marshalConfig(&pkgconfig.AppConfig{Version: "v2", Components: pkgconfig.ComponentList{{Name: "experimental"}}}),
		componentIDs: []string{theirs.ID},
	})

	response := s.decode(s.get(s.branchPath(branch.ID)))

	componentsRR := s.get(fmt.Sprintf("/v1/apps/%s/components?branch_id=%s", s.testApp.ID, branch.ID))
	require.Equal(s.T(), http.StatusOK, componentsRR.Code, componentsRR.Body.String())

	var components []struct {
		ID string `json:"id"`
	}
	require.NoError(s.T(), json.Unmarshal(componentsRR.Body.Bytes(), &components))

	componentIDs := []string{}
	for _, component := range components {
		componentIDs = append(componentIDs, component.ID)
	}

	require.ElementsMatch(
		s.T(),
		componentIDs,
		s.resourceIDs(response.Resources, app.AppConfigResourceKindComponent),
	)
	require.NotContains(s.T(), componentIDs, theirs.ID)
}

func (s *AppConfigIntermediateTestSuite) TestBranchRouteIgnoresNewerNonActiveConfigs() {
	branch := s.seedBranch("main")
	active := s.seedConfig(seedConfigOpts{
		branchID:  branch.ID,
		createdAt: time.Now().UTC().Add(-time.Hour),
	})

	for _, status := range []app.AppConfigStatus{
		app.AppConfigStatusPending,
		app.AppConfigStatusSyncing,
		app.AppConfigStatusError,
	} {
		s.seedConfig(seedConfigOpts{
			branchID:  branch.ID,
			status:    status,
			createdAt: time.Now().UTC(),
		})
	}

	response := s.decode(s.get(s.branchPath(branch.ID)))
	require.Equal(s.T(), active.ID, response.ConfigID)
}

func (s *AppConfigIntermediateTestSuite) TestBranchRouteNotFoundWhenBranchHasNeverSynced() {
	branch := s.seedBranch("never-synced")
	s.seedConfig(seedConfigOpts{})

	rr := s.get(s.branchPath(branch.ID))
	require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())
	require.Contains(s.T(), rr.Body.String(), "has not synced a config yet")
}

func (s *AppConfigIntermediateTestSuite) TestConfigWithoutBlobIsNotFound() {
	cfg := s.seedConfig(seedConfigOpts{noBlob: true})

	rr := s.get(s.configPath(cfg.ID))
	require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())
	require.Contains(s.T(), rr.Body.String(), "without an intermediate config")
}

func (s *AppConfigIntermediateTestSuite) TestResourcesCoverEveryIDAndName() {
	component := s.seedComponent("api")
	action := s.seedAction("rotate-keys")
	runbook := s.seedRunbook("release")

	cfg := s.seedConfig(seedConfigOpts{
		blob: s.marshalConfig(&pkgconfig.AppConfig{
			Version:    "v2",
			Components: pkgconfig.ComponentList{{Name: "api"}},
			Actions:    []*pkgconfig.ActionConfig{{Name: "rotate-keys"}},
			Runbooks:   []*pkgconfig.RunbookConfig{{Name: "release"}},
		}),
		componentIDs: []string{component.ID},
		actionIDs:    []string{action.ID},
		runbookIDs:   []string{runbook.ID},
	})

	response := s.decode(s.get(s.configPath(cfg.ID)))

	byName := map[string]app.AppConfigResource{}
	for _, resource := range response.Resources {
		byName[resource.Name] = resource
	}

	require.Len(s.T(), response.Resources, 3)
	require.Equal(s.T(), component.ID, byName["api"].ID)
	require.Equal(s.T(), app.AppConfigResourceKindComponent, byName["api"].Kind)
	require.Equal(s.T(), action.ID, byName["rotate-keys"].ID)
	require.Equal(s.T(), app.AppConfigResourceKindAction, byName["rotate-keys"].Kind)
	require.Equal(s.T(), runbook.ID, byName["release"].ID)
	require.Equal(s.T(), app.AppConfigResourceKindRunbook, byName["release"].Kind)

	for _, component := range response.Config.Components {
		require.Contains(s.T(), byName, component.Name)
	}
	for _, action := range response.Config.Actions {
		require.Contains(s.T(), byName, action.Name)
	}
	for _, runbook := range response.Config.Runbooks {
		require.Contains(s.T(), byName, runbook.Name)
	}
}

// Config sync matches components by name, so a rename creates a new record and an existing id
// keeps the name it was synced under.
func (s *AppConfigIntermediateTestSuite) TestRenamedComponentResolvesToTheNewRecord() {
	original := s.seedComponent("api")
	renamed := s.seedComponent("api-v2")

	old := s.seedConfig(seedConfigOpts{
		createdAt:    time.Now().UTC().Add(-time.Hour),
		blob:         s.marshalConfig(&pkgconfig.AppConfig{Version: "v2", Components: pkgconfig.ComponentList{{Name: "api"}}}),
		componentIDs: []string{original.ID},
	})
	current := s.seedConfig(seedConfigOpts{
		blob:         s.marshalConfig(&pkgconfig.AppConfig{Version: "v2", Components: pkgconfig.ComponentList{{Name: "api-v2"}}}),
		componentIDs: []string{renamed.ID},
	})

	oldResponse := s.decode(s.get(s.configPath(old.ID)))
	require.Equal(s.T(), "api", oldResponse.Resources[0].Name)
	require.Equal(s.T(), original.ID, oldResponse.Resources[0].ID)

	currentResponse := s.decode(s.get(s.configPath(current.ID)))
	require.Equal(s.T(), "api-v2", currentResponse.Resources[0].Name)
	require.Equal(s.T(), renamed.ID, currentResponse.Resources[0].ID)
}

func (s *AppConfigIntermediateTestSuite) TestAnotherOrgsConfigIsNotFound() {
	otherCtx := context.Background()
	otherCtx, _ = s.deps.Seeder.EnsureAccount(otherCtx, s.T())
	otherCtx, otherOrg := s.deps.Seeder.EnsureOrg(otherCtx, s.T())
	otherApp := s.deps.Seeder.CreateApp(otherCtx, s.T())

	otherBranch := &app.AppBranch{AppID: otherApp.ID, Name: "main"}
	require.NoError(s.T(), s.deps.DB.WithContext(otherCtx).Create(otherBranch).Error)

	otherCfg := &app.AppConfig{
		OrgID:              otherOrg.ID,
		AppID:              otherApp.ID,
		Status:             app.AppConfigStatusActive,
		StatusV2:           app.NewCompositeStatus(otherCtx, app.Status(app.AppConfigStatusActive)),
		StatusDescription:  "active",
		IntermediateConfig: &blobstore.Blob{},
	}
	require.NoError(s.T(), otherCfg.AppBranchID.Scan(otherBranch.ID))
	otherCfg.IntermediateConfig.Set(s.marshalConfig(&pkgconfig.AppConfig{Version: "v2"}))
	blobCtx := blobstore.WithBlobService(otherCtx, s.appsService.blobSvc)
	require.NoError(s.T(), s.deps.DB.WithContext(blobCtx).Create(otherCfg).Error)

	require.Equal(s.T(), http.StatusNotFound, s.get(s.configPath(otherCfg.ID)).Code)
	require.Equal(s.T(), http.StatusNotFound, s.get(s.branchPath(otherBranch.ID)).Code)
}

func (s *AppConfigIntermediateTestSuite) TestSecretValuesAreNotInThePayload() {
	secret := &app.AppSecret{
		AppID: s.testApp.ID,
		Name:  "database-password",
		Value: "s3cr3t-value-that-must-not-leak",
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(secret).Error)

	cfg := s.seedConfig(seedConfigOpts{
		blob: s.marshalConfig(&pkgconfig.AppConfig{
			Version: "v2",
			Secrets: &pkgconfig.SecretsConfig{
				Secrets: []*pkgconfig.AppSecret{{
					Name:        "database-password",
					Description: "password for the primary database",
				}},
			},
		}),
	})

	rr := s.get(s.configPath(cfg.ID))
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())
	require.False(s.T(), strings.Contains(rr.Body.String(), secret.Value))
}
