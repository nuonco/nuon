package syncer

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	pkgsync "github.com/nuonco/nuon/pkg/config/sync"
	temporal "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/terraform"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
	testseedconfig "github.com/nuonco/nuon/services/ctl-api/tests/testseed/config"
)

// syncDeps mirrors the arguments NewDBSyncer takes.
type syncDeps struct {
	fx.In

	DB               *gorm.DB `name:"psql"`
	Seed             *testseed.Seeder
	AppsHelpers      *appshelpers.Helpers
	ComponentHelpers *componenthelpers.Helpers
	ActionsHelpers   *actionshelpers.Helpers
	RunbooksHelpers  *runbookshelpers.Helpers
	InstallHelpers   *installhelpers.Helpers
	VCSHelpers       *vcshelpers.Helpers
	TFClient         terraform.Client
}

// SyncFieldsTestSuite runs the branch-sync path against a real database and
// asserts the config fields it is expected to persist. The CLI path reaches the
// same builders through the HTTP handlers, so a field covered here is covered
// for both.
type SyncFieldsTestSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps syncDeps
}

type syncFieldsWorkflowRun struct{}

func (r *syncFieldsWorkflowRun) GetID() string    { return "test-workflow-id" }
func (r *syncFieldsWorkflowRun) GetRunID() string { return "test-run-id" }
func (r *syncFieldsWorkflowRun) Get(context.Context, interface{}) error {
	return nil
}
func (r *syncFieldsWorkflowRun) GetWithOptions(context.Context, interface{}, tclient.WorkflowRunGetOptions) error {
	return nil
}

func TestSyncFieldsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(SyncFieldsTestSuite))
}

func (s *SyncFieldsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	ctrl := gomock.NewController(s.T())
	mockTC := temporal.NewMockClient(ctrl)
	mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(&syncFieldsWorkflowRun{}, nil).AnyTimes()

	options := append(tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
		T: s.T(),
		Mocks: &tests.TestMocks{
			MockTC: mockTC,
		},
		CustomValidator: true,
	}), fx.Populate(&s.deps))
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.deps.DB)
}

func (s *SyncFieldsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

// sync seeds an org/app/app-config, runs the branch-sync path over cfg, and
// returns the app config it synced into.
func (s *SyncFieldsTestSuite) sync(cfg *config.AppConfig) (context.Context, *app.App, *app.AppConfig) {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	appCfg := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), testApp.ID)

	syncer := NewDBSyncer(
		s.deps.DB,
		s.deps.AppsHelpers,
		s.deps.ComponentHelpers,
		s.deps.ActionsHelpers,
		s.deps.RunbooksHelpers,
		s.deps.InstallHelpers,
		s.deps.VCSHelpers,
		s.deps.TFClient,
		testApp.ID,
		cfg,
		appCfg.ID,
	)
	s.Require().NoError(syncer.Sync(ctx), "sync should succeed")

	return ctx, testApp, appCfg
}

func iamRole(name string) *config.AppAWSIAMRole {
	return &config.AppAWSIAMRole{
		Name:        name,
		DisplayName: name,
		Description: name,
		Policies: []config.AppAWSIAMPolicy{
			{Name: name + "-policy", ManagedPolicyName: "AdministratorAccess"},
		},
	}
}

func (s *SyncFieldsTestSuite) TestPermissionsPersistsCustomRoles() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Permissions = &config.PermissionsConfig{
		ProvisionRole:   iamRole("provision"),
		MaintenanceRole: iamRole("maintenance"),
		DeprovisionRole: iamRole("deprovision"),
		CustomRoles:     []*config.AppAWSIAMRole{iamRole("custom-one"), iamRole("custom-two")},
	}

	_, _, appCfg := s.sync(cfg)

	var roles []app.AppAWSIAMRoleConfig
	s.Require().NoError(s.deps.DB.
		Where("app_config_id = ? AND type = ?", appCfg.ID, app.AWSIAMRoleTypeCustom).
		Find(&roles).Error)

	names := []string{}
	for _, r := range roles {
		names = append(names, r.Name)
	}
	s.ElementsMatch([]string{"custom-one", "custom-two"}, names, "custom roles must be persisted")
}

func (s *SyncFieldsTestSuite) TestBreakGlassConfigRowIsCreated() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Permissions = &config.PermissionsConfig{
		ProvisionRole:   iamRole("provision"),
		MaintenanceRole: iamRole("maintenance"),
		DeprovisionRole: iamRole("deprovision"),
	}
	cfg.BreakGlass = &config.BreakGlass{Roles: []*config.AppAWSIAMRole{iamRole("break-glass")}}

	_, _, appCfg := s.sync(cfg)

	// Every consumer reads AppConfig.BreakGlassConfig.Roles, so the row must
	// exist and carry the roles.
	var bg app.AppBreakGlassConfig
	s.Require().NoError(s.deps.DB.
		Preload("Roles").
		Where("app_config_id = ?", appCfg.ID).
		First(&bg).Error, "break glass config row must exist")

	s.Require().Len(bg.Roles, 1)
	s.Equal("break-glass", bg.Roles[0].Name)
}

func (s *SyncFieldsTestSuite) TestSandboxPersistsPulumiAndOperationRoles() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Sandbox.Type = config.AppSandboxTypePulumi
	cfg.Sandbox.Runtime = "go"
	cfg.Sandbox.PulumiVersion = "3.100.0"
	cfg.Sandbox.PulumiConfig = map[string]string{"region": "us-west-2"}
	cfg.Sandbox.SkipNoops = ptrTo(true)
	cfg.Sandbox.AutoApproveOnPoliciesPassing = ptrTo(true)
	cfg.Sandbox.OperationRoles = []config.EntityOperationRole{
		{Operation: config.OperationType(app.OperationReprovision), RoleName: "maintenance"},
	}

	// The pulumi sandbox is feature gated; enable it for this org.
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, org := s.deps.Seed.EnsureOrg(ctx, s.T())
	s.Require().NoError(s.deps.DB.Model(&app.Org{}).Where("id = ?", org.ID).
		Update("features", types.StringBoolMap{string(app.OrgFeaturePulumiSandbox): true}).Error)

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	appCfg := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), testApp.ID)

	syncer := NewDBSyncer(s.deps.DB, s.deps.AppsHelpers, s.deps.ComponentHelpers,
		s.deps.ActionsHelpers, s.deps.RunbooksHelpers, s.deps.InstallHelpers,
		s.deps.VCSHelpers, s.deps.TFClient, testApp.ID, cfg, appCfg.ID)
	s.Require().NoError(syncer.Sync(ctx))

	var sandbox app.AppSandboxConfig
	s.Require().NoError(s.deps.DB.Where("app_config_id = ?", appCfg.ID).First(&sandbox).Error)

	s.Equal(config.AppSandboxTypePulumi, sandbox.Type, "pulumi sandbox must not fall back to terraform")
	s.Equal("go", sandbox.Runtime)
	s.Equal("3.100.0", sandbox.PulumiVersion)
	s.Require().Contains(sandbox.PulumiConfig, "region")
	s.Equal("us-west-2", *sandbox.PulumiConfig["region"])
	s.Require().NotNil(sandbox.SkipNoops)
	s.True(*sandbox.SkipNoops)
	s.Require().NotNil(sandbox.AutoApproveOnPoliciesPassing)
	s.True(*sandbox.AutoApproveOnPoliciesPassing)
	s.Require().Contains(sandbox.OperationRoles, string(app.OperationReprovision))
	s.Equal("maintenance", *sandbox.OperationRoles[string(app.OperationReprovision)])
}

func (s *SyncFieldsTestSuite) TestJobComponentGetsAConfig() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Components = config.ComponentList{
		{
			Type: config.JobComponentType,
			Name: "migrate",
			Job: &config.JobComponentConfig{
				ImageURL: "ghcr.io/acme/migrate",
				Tag:      "v1",
				Cmd:      []string{"/bin/migrate"},
			},
		},
	}

	_, _, appCfg := s.sync(cfg)

	var ccc app.ComponentConfigConnection
	s.Require().NoError(s.deps.DB.
		Preload("JobComponentConfig").
		Where("app_config_id = ?", appCfg.ID).
		First(&ccc).Error, "job component must get a config connection")

	s.Require().NotNil(ccc.JobComponentConfig, "job component config must be persisted")
	s.Equal("ghcr.io/acme/migrate", ccc.JobComponentConfig.ImageURL)
	s.Equal("v1", ccc.JobComponentConfig.Tag)
}

func (s *SyncFieldsTestSuite) TestKubernetesManifestKeepsInlineManifestAndToggles() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Components = config.ComponentList{
		{
			Type:           config.KubernetesManifestComponentType,
			Name:           "manifests",
			Toggleable:     ptrTo(true),
			DefaultEnabled: ptrTo(false),
			KubernetesManifest: &config.KubernetesManifestComponentConfig{
				Namespace: "default",
				Manifest:  "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: demo\n",
			},
		},
	}

	_, _, appCfg := s.sync(cfg)

	var ccc app.ComponentConfigConnection
	s.Require().NoError(s.deps.DB.
		Preload("KubernetesManifestComponentConfig").
		Where("app_config_id = ?", appCfg.ID).
		First(&ccc).Error)

	s.Require().NotNil(ccc.KubernetesManifestComponentConfig)
	s.Contains(ccc.KubernetesManifestComponentConfig.Manifest, "kind: Namespace", "inline manifest must be persisted")

	s.Require().NotNil(ccc.Toggleable, "toggleable must be persisted")
	s.True(*ccc.Toggleable)
	s.Require().NotNil(ccc.DefaultEnabled, "default_enabled must be persisted")
	s.False(*ccc.DefaultEnabled)
}

func (s *SyncFieldsTestSuite) TestActionStepsAreOrderedAndKeepRole() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Actions = []*config.ActionConfig{
		{
			Name:             "deploy",
			Role:             "maintenance",
			EnableKubeConfig: ptrTo(false),
			Triggers:         []*config.ActionTriggerConfig{{Type: "manual"}},
			Steps: []*config.ActionStepConfig{
				{Name: "first", InlineContents: "echo first"},
				{Name: "second", InlineContents: "echo second"},
				{Name: "third", InlineContents: "echo third"},
			},
		},
	}

	_, _, appCfg := s.sync(cfg)

	var awc app.ActionWorkflowConfig
	s.Require().NoError(s.deps.DB.
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("action_workflow_step_configs.idx ASC")
		}).
		Where("app_config_id = ?", appCfg.ID).
		First(&awc).Error)

	s.Equal("maintenance", awc.Role, "action role must be persisted")
	s.True(awc.EnableKubeConfig.Valid)
	s.False(awc.EnableKubeConfig.Bool, "enable_kube_config=false must be honoured")

	// Steps load ordered by idx, so unset idx values would scramble execution.
	s.Require().Len(awc.Steps, 3)
	s.Equal([]string{"first", "second", "third"},
		[]string{awc.Steps[0].Name, awc.Steps[1].Name, awc.Steps[2].Name},
		"action steps must keep their declared order")
	s.Equal([]int{0, 1, 2}, []int{awc.Steps[0].Idx, awc.Steps[1].Idx, awc.Steps[2].Idx})
}

func (s *SyncFieldsTestSuite) TestPolicyNameIsPersisted() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Policies = &config.PoliciesConfig{
		Policies: []config.AppPolicy{
			{Type: config.AppPolicyTypeSandbox, Name: "no-public-buckets", Contents: "package nuon"},
		},
	}

	_, _, appCfg := s.sync(cfg)

	var policy app.AppPolicyConfig
	s.Require().NoError(s.deps.DB.Where("app_config_id = ?", appCfg.ID).First(&policy).Error)
	s.Equal("no-public-buckets", policy.Name, "policy name must be persisted")
}

func (s *SyncFieldsTestSuite) TestSyncStateIsPersisted() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	terraformComponent := testseedconfig.BuildTerraformComponent("tf")
	terraformComponent.TerraformModule.TerraformVersion = "1.9.0"
	cfg.Components = config.ComponentList{terraformComponent}

	_, _, appCfg := s.sync(cfg)

	var reloaded app.AppConfig
	s.Require().NoError(s.deps.DB.Select("id", "state").Where("id = ?", appCfg.ID).First(&reloaded).Error)
	s.Require().NotEmpty(reloaded.State, "sync state must be persisted for the next CLI sync")

	var state pkgsync.State
	s.Require().NoError(json.Unmarshal([]byte(reloaded.State), &state))
	s.Require().Len(state.Components, 1)
	s.Equal("tf", state.Components[0].Name)
}

func ptrTo[T any](v T) *T { return &v }
