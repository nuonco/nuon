package helpers_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type sdkConfigDeps struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Seed    *testseed.Seeder
	Helpers *installhelpers.Helpers
}

type InstallerSDKConfigTestSuite struct {
	tests.BaseDBTestSuite

	fxApp *fxtest.App
	deps  sdkConfigDeps

	ctx      context.Context
	testApp  *app.App
	appCfg   *app.AppConfig
	inputCfg *app.AppInputConfig
	group    *app.AppInputGroup
}

func TestInstallerSDKConfigSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(InstallerSDKConfigTestSuite))
}

func (s *InstallerSDKConfigTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(tests.CtlApiFXOptions(s.T()), fx.Populate(&s.deps))
	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *InstallerSDKConfigTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *InstallerSDKConfigTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	s.ctx = context.Background()
	s.ctx, _ = s.deps.Seed.EnsureAccount(s.ctx, s.T())
	s.ctx, _ = s.deps.Seed.EnsureOrg(s.ctx, s.T())
	s.testApp = s.deps.Seed.CreateApp(s.ctx, s.T())
	s.appCfg = s.deps.Seed.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)

	s.inputCfg = &app.AppInputConfig{}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where("app_config_id = ?", s.appCfg.ID).First(s.inputCfg).Error)
	s.group = &app.AppInputGroup{}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where("app_input_config_id = ?", s.inputCfg.ID).First(s.group).Error)
}

// customerInput adds a customer-source app input to the install's pinned config.
func (s *InstallerSDKConfigTestSuite) customerInput(name, def string, sensitive bool) {
	in := &app.AppInput{
		AppInputConfigID: s.inputCfg.ID,
		AppInputGroupID:  s.group.ID,
		Name:             name,
		Description:      name,
		Type:             app.AppInputTypeString,
		Source:           app.AppInputSourceCustomer,
		Default:          def,
		Sensitive:        sensitive,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(in).Error)
}

// seedInstallWithRunner gives the install the runner group, runner, and runner API
// URL BuildInstallerSDKConfig requires before it will render anything.
func (s *InstallerSDKConfigTestSuite) seedInstallWithRunner() *app.Install {
	t := s.T()
	install := s.deps.Seed.CreateInstall(s.ctx, t, s.testApp)

	group := &app.RunnerGroup{
		OwnerID:   install.ID,
		OwnerType: "installs",
		Type:      app.RunnerGroupTypeInstall,
		Platform:  app.AppRunnerTypeAWS,
	}
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Create(group).Error)
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Create(&app.RunnerGroupSettings{
		RunnerGroupID: group.ID,
		RunnerAPIURL:  "https://runner.example.com",
		// the settings BeforeCreate hook tags the group id onto Metadata in place
		Metadata: pgtype.Hstore{},
	}).Error)
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Create(&app.Runner{
		RunnerGroupID:     group.ID,
		Name:              "runner-" + install.ID,
		DisplayName:       "runner",
		Status:            app.RunnerStatusActive,
		StatusDescription: "active",
	}).Error)

	return install
}

func (s *InstallerSDKConfigTestSuite) build(installID string) *app.InstallerSDKConfig {
	cfg, err := s.deps.Helpers.BuildInstallerSDKConfig(s.ctx, installID)
	require.NoError(s.T(), err)
	return cfg
}

// The config read is authenticated, so it serves the install's real current input
// values — the names-only contract belonged to the unauthenticated tfvars flow.
func (s *InstallerSDKConfigTestSuite) TestServesCurrentInputValues() {
	t := s.T()

	s.customerInput("domain", "default.example.com", false)
	s.customerInput("bucket", "", false)
	install := s.seedInstallWithRunner()

	s.deps.Seed.CreateInstallInputs(s.ctx, t, install.ID, s.inputCfg.ID, map[string]*string{
		"domain": generics.ToPtr("set.example.com"),
	})

	cfg := s.build(install.ID)
	assert.Equal(t, "set.example.com", cfg.InstallInputs["domain"])
	assert.Equal(t, "", cfg.InstallInputs["bucket"])
	// Vendor-source inputs stay out: install_inputs is the customer's surface.
	assert.NotContains(t, cfg.InstallInputs, "region")
}

// With no value stored, the app input's default is what the stack should apply.
func (s *InstallerSDKConfigTestSuite) TestFallsBackToAppInputDefault() {
	t := s.T()

	s.customerInput("domain", "default.example.com", false)
	s.customerInput("bucket", "", false)
	install := s.seedInstallWithRunner()

	cfg := s.build(install.ID)
	assert.Equal(t, "default.example.com", cfg.InstallInputs["domain"])
	assert.Equal(t, "", cfg.InstallInputs["bucket"])
}

// install_inputs is a plain map, so sensitivity has to travel beside it — without
// this list the provider cannot know which values to mark sensitive.
func (s *InstallerSDKConfigTestSuite) TestReportsSensitiveInputNames() {
	t := s.T()

	s.customerInput("domain", "", false)
	s.customerInput("api_key", "", true)
	install := s.seedInstallWithRunner()

	cfg := s.build(install.ID)
	assert.Equal(t, []string{"api_key"}, cfg.SensitiveInputs)
}

// cluster_name resolves from the install's current inputs, which only works if the
// install's inputs are actually loaded.
func (s *InstallerSDKConfigTestSuite) TestClusterNameFromCurrentInputs() {
	t := s.T()

	s.customerInput("cluster_name", "", false)
	install := s.seedInstallWithRunner()
	s.deps.Seed.CreateInstallInputs(s.ctx, t, install.ID, s.inputCfg.ID, map[string]*string{
		"cluster_name": generics.ToPtr("my-cluster"),
	})

	cfg := s.build(install.ID)
	require.NotNil(t, cfg.AWS)
	assert.Equal(t, "my-cluster", cfg.AWS.ClusterName)
}
