package helpers_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type migrateInputsDeps struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Seed    *testseed.Seeder
	Helpers *installhelpers.Helpers
}

type MigrateInstallInputsTestSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps migrateInputsDeps
}

func TestMigrateInstallInputsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(MigrateInstallInputsTestSuite))
}

func (s *MigrateInstallInputsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(tests.CtlApiFXOptions(s.T()), fx.Populate(&s.deps))
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *MigrateInstallInputsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

// seedAppConfigWithInputs creates a further app config carrying an input config
// whose single input is named "region" (what the seeder builds).
func (s *MigrateInstallInputsTestSuite) seedAppConfigWithInputs(ctx context.Context, appID string) *app.AppConfig {
	cfg := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), appID)
	s.deps.Seed.CreateAppInputConfig(ctx, s.T(), appID, cfg.ID)
	return cfg
}

func (s *MigrateInstallInputsTestSuite) latestInstallInputs(ctx context.Context, installID string) app.InstallInputs {
	var got app.InstallInputs
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.InstallInputs{InstallID: installID}).
		Order("created_at DESC").
		First(&got).Error)
	return got
}

func ptrTo[T any](v T) *T { return &v }

// The migration must carry values onto the new config. Dropping them leaves
// templates like {{.nuon.inputs.inputs.region}} resolving to nil at plan time.
func (s *MigrateInstallInputsTestSuite) TestMigratesValuesTiedToOutgoingConfig() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	oldCfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)

	var oldInputCfg app.AppInputConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: oldCfg.ID}).First(&oldInputCfg).Error)
	s.deps.Seed.CreateInstallInputs(ctx, s.T(), install.ID, oldInputCfg.ID,
		map[string]*string{"region": ptrTo("us-west-2")})

	newCfg := s.seedAppConfigWithInputs(ctx, testApp.ID)
	s.Require().NoError(s.deps.Helpers.MigrateInstallInputsToNewConfig(ctx, s.deps.DB,
		map[string]string{install.ID: oldCfg.ID}, newCfg.ID))

	got := s.latestInstallInputs(ctx, install.ID)
	s.Require().NotNil(got.Values["region"])
	s.Equal("us-west-2", *got.Values["region"])
}

// Regression for #857: the lookup uses Find, which reports "no rows" through
// RowsAffected rather than gorm.ErrRecordNotFound. An install whose inputs are
// tied to some other config version took the miss path and had its values
// silently replaced with an empty set, which then persisted through every later
// migration.
func (s *MigrateInstallInputsTestSuite) TestMigratesValuesNotTiedToOutgoingConfig() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	oldCfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)

	// Values live against an unrelated input config, so the primary lookup misses.
	strayCfg := s.seedAppConfigWithInputs(ctx, testApp.ID)
	var strayInputCfg app.AppInputConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: strayCfg.ID}).First(&strayInputCfg).Error)
	s.deps.Seed.CreateInstallInputs(ctx, s.T(), install.ID, strayInputCfg.ID,
		map[string]*string{"region": ptrTo("eu-central-1")})

	newCfg := s.seedAppConfigWithInputs(ctx, testApp.ID)
	s.Require().NoError(s.deps.Helpers.MigrateInstallInputsToNewConfig(ctx, s.deps.DB,
		map[string]string{install.ID: oldCfg.ID}, newCfg.ID))

	got := s.latestInstallInputs(ctx, install.ID)
	s.Require().NotNil(got.Values["region"], "fallback must carry the install's latest values forward")
	s.Equal("eu-central-1", *got.Values["region"])
}

// Inputs dropped from the new config must not be carried forward.
func (s *MigrateInstallInputsTestSuite) TestDropsValuesRemovedFromNewConfig() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	oldCfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)

	var oldInputCfg app.AppInputConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: oldCfg.ID}).First(&oldInputCfg).Error)
	s.deps.Seed.CreateInstallInputs(ctx, s.T(), install.ID, oldInputCfg.ID, map[string]*string{
		"region": ptrTo("us-west-2"),
		"gone":   ptrTo("stale"),
	})

	newCfg := s.seedAppConfigWithInputs(ctx, testApp.ID)
	s.Require().NoError(s.deps.Helpers.MigrateInstallInputsToNewConfig(ctx, s.deps.DB,
		map[string]string{install.ID: oldCfg.ID}, newCfg.ID))

	got := s.latestInstallInputs(ctx, install.ID)
	s.Equal("us-west-2", *got.Values["region"])
	s.NotContains(got.Values, "gone")
}
