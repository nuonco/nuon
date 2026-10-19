package activities

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type syncInstallConfigDeps struct {
	fx.In
	DB      *gorm.DB `name:"psql"`
	Seed    *testseed.Seeder
	Helpers *helpers.Helpers
	Blob    blobstore.Service
	L       *zap.Logger
}

type syncInstallConfigTestSuite struct {
	tests.BaseDBTestSuite
	app  *fxtest.App
	deps syncInstallConfigDeps
	acts *Activities
}

func TestSyncInstallConfigSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
	}
	suite.Run(t, new(syncInstallConfigTestSuite))
}

func (s *syncInstallConfigTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	tc := temporalclient.NewMockClient(gomock.NewController(s.T()))
	run := mocks.NewWorkflowRun(s.T())
	run.On("GetRunID").Return("test-run-id").Maybe()
	tc.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(run, nil).AnyTimes()
	opts := tests.CtlApiFXOptionsWithMocks(tests.TestOpts{T: s.T(), Mocks: &tests.TestMocks{MockTC: tc}, CustomValidator: true})
	s.app = fxtest.New(s.T(), append(opts, fx.Populate(&s.deps))...)
	s.app.RequireStart()
	s.SetDB(s.deps.DB)
	s.acts = &Activities{db: s.deps.DB, helpers: s.deps.Helpers, blobSvc: s.deps.Blob, l: s.deps.L}
}

func (s *syncInstallConfigTestSuite) TearDownSuite() {
	if s.app != nil {
		s.app.RequireStop()
	}
}

func (s *syncInstallConfigTestSuite) fixture() (context.Context, string, *config.Install) {
	ctx, _ := s.deps.Seed.EnsureAccount(context.Background(), s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())
	a := s.deps.Seed.CreateApp(ctx, s.T())
	cfg := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), a.ID)
	s.deps.Seed.CreateAppSandboxConfig(ctx, s.T(), a.ID, cfg.ID)
	s.deps.Seed.CreateAppRunnerConfig(ctx, s.T(), a.ID, cfg.ID)
	s.deps.Seed.CreateAppInputConfig(ctx, s.T(), a.ID, cfg.ID)
	s.deps.Seed.CreateAppSecretsConfig(ctx, s.T(), a.ID, cfg.ID)
	s.deps.Seed.CreateAppPermissionsConfig(ctx, s.T(), a.ID, cfg.ID)
	s.deps.Seed.CreateAppStackConfig(ctx, s.T(), a.ID, cfg.ID)
	connection := &app.CloudConnection{
		Name: "acme-aws", Platform: app.CloudPlatformAWS, TargetID: "123456789012",
		Principal: "arn:aws:iam::123456789012:role/acme", DefaultRegion: "us-west-2", Status: app.CloudConnectionStatusVerified,
	}
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(connection).Error)
	return ctx, a.ID, &config.Install{
		Name: "acme", CloudConnection: connection.Name,
		AWSAccount:     &config.AWSAccount{Region: "us-west-2", AccountID: "123456789012"},
		ApprovalOption: config.InstallApprovalOptionApproveAll,
		CloudFormation: &config.CloudFormation{Root: config.CloudFormationStack{Parameters: map[string]string{"RunnerInstanceType": "t3.small"}}},
		InputGroups:    []config.InputGroup{{Inputs: map[string]string{"region": "us-west-2"}}},
	}
}

func (s *syncInstallConfigTestSuite) input(ctx context.Context, appID string, cfg *config.Install) *SyncInstallConfigInput {
	sync := &app.InstallConfigSync{Status: app.NewCompositeStatus(ctx, app.StatusPending)}
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(sync).Error)
	return &SyncInstallConfigInput{AppID: appID, InstallConfig: cfg, InstallConfigSyncID: sync.ID, FilePath: "installs/acme.toml"}
}

func (s *syncInstallConfigTestSuite) workflow(ctx context.Context, result *SyncInstallConfigOutput, wantType app.WorkflowType) {
	s.Require().NotEmpty(result.WorkflowID)
	var wf app.Workflow
	s.Require().NoError(s.deps.DB.WithContext(ctx).Where(app.Workflow{ID: result.WorkflowID}).First(&wf).Error)
	s.Equal(result.InstallID, wf.OwnerID)
	s.Equal(wantType, wf.Type)
	s.Equal(app.InstallApprovalOptionApproveAll, wf.ApprovalOption)
	s.Require().NotNil(wf.Metadata["install_config_version_id"])
	s.Equal(result.InstallConfigVersionID, *wf.Metadata["install_config_version_id"])
	if wantType == app.WorkflowTypeProvision {
		s.Require().NotNil(wf.Metadata[app.WorkflowMetadataKeyStackOnly])
		s.Equal("true", *wf.Metadata[app.WorkflowMetadataKeyStackOnly])
	}
	var signal app.QueueSignal
	s.Require().NoError(s.deps.DB.WithContext(ctx).Where(app.QueueSignal{OwnerID: wf.ID, OwnerType: "install_workflows"}).First(&signal).Error)
	var q app.Queue
	s.Require().NoError(s.deps.DB.WithContext(ctx).Where(app.Queue{ID: signal.QueueID}).First(&q).Error)
	s.Equal(result.InstallID, q.OwnerID)
	s.Equal(queuenames.InstallWorkflowsQueueName, q.Name)
}

func (s *syncInstallConfigTestSuite) counts(ctx context.Context, installID string, wantWorkflows int64) {
	var count int64
	s.Require().NoError(s.deps.DB.WithContext(ctx).Model(&app.Workflow{}).Where(app.Workflow{OwnerID: installID, OwnerType: "installs"}).Count(&count).Error)
	s.Equal(wantWorkflows, count, "workflow count")
	var q app.Queue
	s.Require().NoError(s.deps.DB.WithContext(ctx).Where(app.Queue{OwnerID: installID, Name: queuenames.InstallWorkflowsQueueName}).First(&q).Error)
	s.Require().NoError(s.deps.DB.WithContext(ctx).Model(&app.QueueSignal{}).Where(app.QueueSignal{QueueID: q.ID}).Count(&count).Error)
	s.Equal(wantWorkflows, count, "queued workflow count")
}

func (s *syncInstallConfigTestSuite) TestSchedulingAndNoops() {
	ctx, appID, cfg := s.fixture()
	input := s.input(ctx, appID, cfg)
	input.CreateOnly = true
	created, err := s.acts.SyncInstallConfig(ctx, input)
	s.Require().NoError(err)
	s.True(created.Created)
	s.True(created.Changed)
	s.workflow(ctx, created, app.WorkflowTypeProvision)
	s.counts(ctx, created.InstallID, 1)

	for _, tt := range []struct {
		name       string
		overrides  *config.CloudFormation
		region     string
		changed    bool
		parameters map[string]string
		workflows  int64
	}{
		{"same config", cfg.CloudFormation, "us-west-2", false, map[string]string{"RunnerInstanceType": "t3.small"}, 1},
		{"parameter-only change", &config.CloudFormation{Root: config.CloudFormationStack{Parameters: map[string]string{"RunnerInstanceType": "m6i.large"}}}, "us-west-2", true, map[string]string{"RunnerInstanceType": "m6i.large"}, 2},
		{"input-only change", &config.CloudFormation{Root: config.CloudFormationStack{Parameters: map[string]string{"RunnerInstanceType": "m6i.large"}}}, "eu-central-1", true, map[string]string{"RunnerInstanceType": "m6i.large"}, 3},
		{"omitted parameters preserve saved values", &config.CloudFormation{}, "eu-central-1", false, map[string]string{"RunnerInstanceType": "m6i.large"}, 3},
		{"omitted overrides preserve saved values", nil, "eu-central-1", false, map[string]string{"RunnerInstanceType": "m6i.large"}, 3},
		{"explicit clear", &config.CloudFormation{Root: config.CloudFormationStack{Parameters: map[string]string{}}}, "eu-central-1", true, map[string]string{}, 4},
		{"repeated clear", &config.CloudFormation{Root: config.CloudFormationStack{Parameters: map[string]string{}}}, "eu-central-1", false, map[string]string{}, 4},
	} {
		s.Run(tt.name, func() {
			cfg.CloudFormation = tt.overrides
			cfg.InputGroups[0].Inputs["region"] = tt.region
			input := s.input(ctx, appID, cfg)
			result, err := s.acts.SyncInstallConfig(ctx, input)
			s.Require().NoError(err)
			s.Equal(created.InstallID, result.InstallID)
			s.False(result.Created)
			s.Equal(tt.changed, result.Changed)
			if tt.changed {
				s.workflow(ctx, result, app.WorkflowTypeReprovisionStack)
			} else {
				s.Empty(result.WorkflowID)
				s.Empty(result.InstallConfigVersionID)
			}
			s.counts(ctx, created.InstallID, tt.workflows)
			var versions int64
			s.Require().NoError(s.deps.DB.WithContext(ctx).Model(&app.InstallConfigVersion{}).Where(app.InstallConfigVersion{InstallConfigSyncID: input.InstallConfigSyncID}).Count(&versions).Error)
			if tt.changed {
				s.Equal(int64(1), versions)
			} else {
				s.Zero(versions, "a no-op must not create a config version")
			}
			var saved app.InstallConfig
			s.Require().NoError(s.deps.DB.WithContext(ctx).Where(app.InstallConfig{InstallID: created.InstallID}).Order("created_at DESC, id DESC").First(&saved).Error)
			s.Equal(tt.parameters, saved.AWSStackParameters)
			var inputs app.InstallInputs
			s.Require().NoError(s.deps.DB.WithContext(ctx).Where(app.InstallInputs{InstallID: created.InstallID}).Order("created_at DESC, id DESC").First(&inputs).Error)
			s.Require().NotNil(inputs.Values["region"])
			s.Equal(tt.region, *inputs.Values["region"])
		})
	}
}

func (s *syncInstallConfigTestSuite) TestNestedStackScheduling() {
	ctx, appID, cfg := s.fixture()
	nested := map[string]config.CloudFormationStack{
		"database": {Parameters: map[string]string{"InstanceClass": "db.t4g.small", "OptionalValue": ""}},
		"VPC":      {Parameters: map[string]string{"VpcCIDR": "10.42.0.0/16"}},
	}
	cfg.CloudFormation.NestedStacks = nested
	created, err := s.acts.SyncInstallConfig(ctx, s.input(ctx, appID, cfg))
	s.Require().NoError(err)
	s.workflow(ctx, created, app.WorkflowTypeProvision)
	changed := map[string]config.CloudFormationStack{
		"database": {Parameters: map[string]string{"InstanceClass": "db.r7g.large"}},
	}
	for _, tt := range []struct {
		name           string
		cloudformation *config.CloudFormation
		want           map[string]config.CloudFormationStack
		changed        bool
		workflows      int64
	}{
		{"initial values are a no-op", cfg.CloudFormation, nested, false, 1},
		{"nested-only change", &config.CloudFormation{NestedStacks: changed}, changed, true, 2},
		{"unchanged nested parameters", &config.CloudFormation{NestedStacks: changed}, changed, false, 2},
		{"omitted nested stacks", &config.CloudFormation{}, changed, false, 2},
		{"omitted cloudformation", nil, changed, false, 2},
		{"explicit clear", &config.CloudFormation{NestedStacks: map[string]config.CloudFormationStack{}}, map[string]config.CloudFormationStack{}, true, 3},
		{"repeated clear", &config.CloudFormation{NestedStacks: map[string]config.CloudFormationStack{}}, map[string]config.CloudFormationStack{}, false, 3},
	} {
		s.Run(tt.name, func() {
			cfg.CloudFormation = tt.cloudformation
			result, err := s.acts.SyncInstallConfig(ctx, s.input(ctx, appID, cfg))
			s.Require().NoError(err)
			s.Equal(created.InstallID, result.InstallID)
			s.Equal(tt.changed, result.Changed)
			if tt.changed {
				s.workflow(ctx, result, app.WorkflowTypeReprovisionStack)
			} else {
				s.Empty(result.WorkflowID)
			}
			var saved app.InstallConfig
			s.Require().NoError(s.deps.DB.WithContext(ctx).Where(app.InstallConfig{InstallID: created.InstallID}).First(&saved).Error)
			s.Equal(tt.want, saved.AWSNestedStackParameters)
			s.Equal(map[string]string{"RunnerInstanceType": "t3.small"}, saved.AWSStackParameters)
			s.counts(ctx, created.InstallID, tt.workflows)
		})
	}
}

func (s *syncInstallConfigTestSuite) TestSchedulingFailureCanBeRetriedWithoutDuplicates() {
	for name, existing := range map[string]bool{"new approved install": false, "existing install": true} {
		s.Run(name, func() {
			ctx, appID, cfg := s.fixture()
			cfg.CloudFormation.NestedStacks = map[string]config.CloudFormationStack{
				"database": {Parameters: map[string]string{"InstanceClass": "db.r7g.large"}},
			}
			baseline := int64(0)
			if existing {
				_, err := s.acts.SyncInstallConfig(ctx, s.input(ctx, appID, cfg))
				s.Require().NoError(err)
				baseline = 1
			}
			cfg.CloudFormation.Root.Parameters["RunnerInstanceType"] = "m6i.large"
			input := s.input(ctx, appID, cfg)
			input.CreateOnly = !existing
			failing := true
			callbackName := "test:stack-parameters:" + input.InstallConfigSyncID
			s.Require().NoError(s.deps.DB.Callback().Create().Before("gorm:create").Register(callbackName, func(tx *gorm.DB) {
				if signal, ok := tx.Statement.Dest.(*app.QueueSignal); failing && ok && signal.OwnerType == "install_workflows" {
					tx.AddError(errors.New("test queue insert failure"))
				}
			}))
			s.T().Cleanup(func() { s.deps.DB.Callback().Create().Remove(callbackName) })
			_, err := s.acts.SyncInstallConfig(ctx, input)
			s.Require().ErrorContains(err, "test queue insert failure")
			failing = false

			var pending app.InstallConfigVersion
			s.Require().NoError(s.deps.DB.WithContext(ctx).Where(app.InstallConfigVersion{InstallConfigSyncID: input.InstallConfigSyncID}).First(&pending).Error)
			s.Equal(app.StatusPending, pending.Status.Status)
			s.Equal(!existing, pending.Created)
			s.counts(ctx, pending.InstallID, baseline)

			result, err := s.acts.SyncInstallConfig(ctx, input)
			s.Require().NoError(err)
			s.Equal(pending.InstallID, result.InstallID)
			s.Equal(pending.ID, result.InstallConfigVersionID)
			s.Equal(!existing, result.Created)
			wantType := app.WorkflowTypeReprovisionStack
			if !existing {
				wantType = app.WorkflowTypeProvision
			}
			s.workflow(ctx, result, wantType)
			repeated, err := s.acts.SyncInstallConfig(ctx, input)
			s.Require().NoError(err)
			s.Equal(result, repeated)
			s.counts(ctx, result.InstallID, baseline+1)
			var count int64
			s.Require().NoError(s.deps.DB.WithContext(ctx).Model(&app.Install{}).Where(app.Install{AppID: appID, Name: cfg.Name}).Count(&count).Error)
			s.Equal(int64(1), count, "install count")
			s.Require().NoError(s.deps.DB.WithContext(ctx).Model(&app.InstallConfigVersion{}).Where(app.InstallConfigVersion{InstallConfigSyncID: input.InstallConfigSyncID}).Count(&count).Error)
			s.Equal(int64(1), count, "config version count")
			var saved app.InstallConfig
			s.Require().NoError(s.deps.DB.WithContext(ctx).Where(app.InstallConfig{InstallID: result.InstallID}).First(&saved).Error)
			s.Equal(cfg.CloudFormation.NestedStacks, saved.AWSNestedStackParameters)
			cfg.CloudFormation.Root.Parameters["RunnerInstanceType"] = "m6i.xlarge"
			_, err = s.acts.SyncInstallConfig(ctx, input)
			s.Require().ErrorContains(err, "different config")
			s.counts(ctx, result.InstallID, baseline+1)
		})
	}
}

func (s *syncInstallConfigTestSuite) TestCloudFormationRequiresAWSConnection() {
	ctx, appID, cfg := s.fixture()
	cfg.AWSAccount = nil
	s.Require().NoError(s.deps.DB.WithContext(ctx).Model(&app.CloudConnection{}).
		Where(app.CloudConnection{Name: cfg.CloudConnection}).Update("platform", app.CloudPlatformAzure).Error)
	_, err := s.acts.SyncInstallConfig(ctx, s.input(ctx, appID, cfg))
	s.Require().ErrorContains(err, "cloudformation requires an AWS cloud_connection")
	var count int64
	s.Require().NoError(s.deps.DB.WithContext(ctx).Model(&app.Install{}).Where(app.Install{AppID: appID}).Count(&count).Error)
	s.Zero(count)
}
