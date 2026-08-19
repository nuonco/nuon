package helpers_test

import (
	"context"
	"os"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type getConfigGraphDeps struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Seed    *testseed.Seeder
	Helpers *appshelpers.Helpers
}

type GetConfigGraphTestSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps getConfigGraphDeps
}

func TestGetConfigGraphSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetConfigGraphTestSuite))
}

func (s *GetConfigGraphTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(tests.CtlApiFXOptions(s.T()), fx.Populate(&s.deps))
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *GetConfigGraphTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetConfigGraphTestSuite) setComponentIDs(ctx context.Context, cfgID string, ids ...string) {
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Model(&app.AppConfig{}).
		Where(app.AppConfig{ID: cfgID}).
		Update("component_ids", pq.StringArray(ids)).Error)
}

func (s *GetConfigGraphTestSuite) setDependencies(ctx context.Context, cccID string, depIDs ...string) {
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Model(&app.ComponentConfigConnection{}).
		Where(app.ComponentConfigConnection{ID: cccID}).
		Update("component_dependency_ids", pq.StringArray(depIDs)).Error)
}

// an unchanged component must resolve to its config as of this version, not its newest one
func (s *GetConfigGraphTestSuite) TestGraphIgnoresDependenciesAddedAfterConfigVersion() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	compA := s.deps.Seed.CreateComponent(ctx, s.T(), testApp.ID, app.ComponentTypeTerraformModule)
	compB := s.deps.Seed.CreateComponent(ctx, s.T(), testApp.ID, app.ComponentTypeTerraformModule)
	compC := s.deps.Seed.CreateComponent(ctx, s.T(), testApp.ID, app.ComponentTypeTerraformModule)

	cfgV1 := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), testApp.ID)
	s.setComponentIDs(ctx, cfgV1.ID, compA.ID, compB.ID)
	s.deps.Seed.CreateTerraformComponentConfigConnection(ctx, s.T(), compA.ID, cfgV1.ID)
	s.deps.Seed.CreateTerraformComponentConfigConnection(ctx, s.T(), compB.ID, cfgV1.ID)

	// only A changed here, so B carries no row for this version
	cfgV2 := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), testApp.ID)
	s.setComponentIDs(ctx, cfgV2.ID, compA.ID, compB.ID)
	s.deps.Seed.CreateTerraformComponentConfigConnection(ctx, s.T(), compA.ID, cfgV2.ID)

	cfgV3 := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), testApp.ID)
	s.setComponentIDs(ctx, cfgV3.ID, compA.ID, compB.ID, compC.ID)
	s.deps.Seed.CreateTerraformComponentConfigConnection(ctx, s.T(), compC.ID, cfgV3.ID)
	cccBV3 := s.deps.Seed.CreateTerraformComponentConfigConnection(ctx, s.T(), compB.ID, cfgV3.ID)
	s.setDependencies(ctx, cccBV3.ID, compC.ID)

	cfg, err := s.deps.Helpers.GetFullAppConfig(ctx, cfgV2.ID, true)
	s.Require().NoError(err)

	grph, err := s.deps.Helpers.GetConfigGraph(ctx, cfg)
	s.Require().NoError(err, "v2 graph must not fail on a dependency introduced in v3")

	_, err = grph.Vertex(compC.ID)
	s.Error(err, "a component added in v3 must not appear in the v2 graph")

	order, err := s.deps.Helpers.GetConfigDefaultComponentOrder(ctx, cfg)
	s.Require().NoError(err)
	s.ElementsMatch([]string{compA.ID, compB.ID}, order)
}

// one unresolvable dependency must not take down the whole graph
func (s *GetConfigGraphTestSuite) TestGraphSkipsDependencyMissingFromConfig() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	compA := s.deps.Seed.CreateComponent(ctx, s.T(), testApp.ID, app.ComponentTypeTerraformModule)
	compB := s.deps.Seed.CreateComponent(ctx, s.T(), testApp.ID, app.ComponentTypeTerraformModule)

	cfg := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), testApp.ID)
	s.setComponentIDs(ctx, cfg.ID, compA.ID, compB.ID)
	s.deps.Seed.CreateTerraformComponentConfigConnection(ctx, s.T(), compA.ID, cfg.ID)
	cccB := s.deps.Seed.CreateTerraformComponentConfigConnection(ctx, s.T(), compB.ID, cfg.ID)
	s.setDependencies(ctx, cccB.ID, "cmpdoesnotexist00000000000")

	fullCfg, err := s.deps.Helpers.GetFullAppConfig(ctx, cfg.ID, true)
	s.Require().NoError(err)

	_, err = s.deps.Helpers.GetConfigGraph(ctx, fullCfg)
	s.Require().NoError(err, "an unresolvable dependency must not fail the graph build")

	order, err := s.deps.Helpers.GetConfigDefaultComponentOrder(ctx, fullCfg)
	s.Require().NoError(err)
	s.ElementsMatch([]string{compA.ID, compB.ID}, order)
}
