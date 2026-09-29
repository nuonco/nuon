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

	"github.com/nuonco/nuon/pkg/config"
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
		Metadata:      pgtype.Hstore{},
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

func (s *InstallerSDKConfigTestSuite) setCustomNestedStacks(stacks []config.CustomNestedStack) {
	var stackConfig app.AppStackConfig
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where("app_config_id = ?", s.appCfg.ID).
		First(&stackConfig).Error)
	stackConfig.CustomNestedStacks = stacks
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Save(&stackConfig).Error)
}

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
	assert.NotContains(t, cfg.InstallInputs, "region")
}

func (s *InstallerSDKConfigTestSuite) TestFallsBackToAppInputDefault() {
	t := s.T()

	s.customerInput("domain", "default.example.com", false)
	s.customerInput("bucket", "", false)
	install := s.seedInstallWithRunner()

	cfg := s.build(install.ID)
	assert.Equal(t, "default.example.com", cfg.InstallInputs["domain"])
	assert.Equal(t, "", cfg.InstallInputs["bucket"])
}

func (s *InstallerSDKConfigTestSuite) TestReportsSensitiveInputNames() {
	t := s.T()

	s.customerInput("domain", "", false)
	s.customerInput("api_key", "", true)
	install := s.seedInstallWithRunner()

	cfg := s.build(install.ID)
	assert.Equal(t, []string{"api_key"}, cfg.SensitiveInputs)
}

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

func (s *InstallerSDKConfigTestSuite) TestCustomStacksEmptyWhenAbsent() {
	install := s.seedInstallWithRunner()

	cfg := s.build(install.ID)
	assert.Empty(s.T(), cfg.CustomStacks)
}

func (s *InstallerSDKConfigTestSuite) TestCustomStacksSortedByIndex() {
	t := s.T()

	s.setCustomNestedStacks([]config.CustomNestedStack{
		{Name: "second", Index: 2, Parameters: map[string]string{"a": "1"}},
		{Name: "first", Index: 0, Parameters: map[string]string{"b": "2"}},
		{Name: "third", Index: 1},
	})

	install := s.seedInstallWithRunner()

	cfg := s.build(install.ID)
	require.Len(t, cfg.CustomStacks, 3)
	assert.Equal(t, []string{"first", "third", "second"}, []string{
		cfg.CustomStacks[0].Name, cfg.CustomStacks[1].Name, cfg.CustomStacks[2].Name,
	})
	assert.Equal(t, map[string]string{"b": "2"}, cfg.CustomStacks[0].Parameters)
}

func (s *InstallerSDKConfigTestSuite) TestCustomStacksInstallOverrideMerge() {
	t := s.T()

	s.setCustomNestedStacks([]config.CustomNestedStack{
		{Name: "shared", Index: 0, Parameters: map[string]string{"env": "app-default"}},
	})

	install := s.seedInstallWithRunner()
	installConfig := &app.InstallConfig{
		InstallID: install.ID,
		CustomNestedStacks: []config.CustomNestedStack{
			{Name: "shared", Index: 0, Parameters: map[string]string{"env": "install-override"}},
			{Name: "extra", Index: 1, Parameters: map[string]string{"env": "install-only"}},
		},
	}
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Create(installConfig).Error)

	cfg := s.build(install.ID)
	require.Len(t, cfg.CustomStacks, 2)
	assert.Equal(t, "shared", cfg.CustomStacks[0].Name)
	assert.Equal(t, "install-override", cfg.CustomStacks[0].Parameters["env"])
	assert.Equal(t, "extra", cfg.CustomStacks[1].Name)
}

func (s *InstallerSDKConfigTestSuite) TestCustomStacksInputParametersFromLatestStackVersion() {
	t := s.T()
	s.customerInput("namespaces", "", false)
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Create(&app.AppInput{
		AppInputConfigID: s.inputCfg.ID,
		AppInputGroupID:  s.group.ID,
		Name:             "vendor_setting",
		Description:      "vendor_setting",
		Type:             app.AppInputTypeString,
		Source:           app.AppInputSourceVendor,
		Default:          "default",
	}).Error)

	s.setCustomNestedStacks([]config.CustomNestedStack{
		{Name: "k8s-namespaces", Index: 0, TemplateURL: "https://nuon-artifacts.s3.us-west-2.amazonaws.com/does-not-exist.yaml"},
	})

	install := s.seedInstallWithRunner()
	installStack := &app.InstallStack{InstallID: install.ID}
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Create(installStack).Error)
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Create(&app.InstallStackVersion{
		InstallID:      install.ID,
		InstallStackID: installStack.ID,
		AppConfigID:    s.appCfg.ID,
		CustomStacksInputParametersMap: map[string]map[string]string{
			"k8s-namespaces": {
				"K8SNamespacesNamespaces": "namespaces",
				"VendorSetting":           "vendor_setting",
			},
		},
	}).Error)

	cfg := s.build(install.ID)
	require.Len(t, cfg.CustomStacks, 1)
	assert.Equal(t, map[string]string{"K8SNamespacesNamespaces": "namespaces"}, cfg.CustomStacks[0].InputParameters)
}

func (s *InstallerSDKConfigTestSuite) TestCustomStacksGCPModuleName() {
	t := s.T()
	s.customerInput("bucket_location", "", false)
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Create(&app.AppInput{
		AppInputConfigID: s.inputCfg.ID,
		AppInputGroupID:  s.group.ID,
		Name:             "bucket_versioning",
		Description:      "bucket_versioning",
		Type:             app.AppInputTypeBool,
		Source:           app.AppInputSourceVendor,
		Default:          "false",
	}).Error)

	s.setCustomNestedStacks([]config.CustomNestedStack{
		{
			Name:        "bucket",
			Index:       0,
			TemplateURL: "github.com/nuonco/install-stacks//gcp/modules/bucket",
			Parameters: map[string]string{
				"location":   "{{.nuon.install.inputs.bucket_location}}",
				"versioning": "{{.nuon.install.inputs.bucket_versioning}}",
			},
		},
		{
			Name:        "cf-nested",
			Index:       1,
			TemplateURL: "https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/nested.yaml",
		},
	})

	install := s.seedInstallWithRunner()
	s.deps.Seed.CreateInstallInputs(s.ctx, t, install.ID, s.inputCfg.ID, map[string]*string{
		"bucket_location":   generics.ToPtr("US"),
		"bucket_versioning": generics.ToPtr("true"),
	})

	cfg := s.build(install.ID)
	require.Len(t, cfg.CustomStacks, 2)
	assert.Equal(t, "bucket", cfg.CustomStacks[0].Module)
	assert.Equal(t, map[string]string{"location": "US", "versioning": "true"}, cfg.CustomStacks[0].Parameters)
	assert.Equal(t, map[string]string{"location": "bucket_location"}, cfg.CustomStacks[0].InputParameters)
	assert.Equal(t, "", cfg.CustomStacks[1].Module)
}
