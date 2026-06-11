package testseed

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// BuildAppConfig creates an app.AppConfig with fake defaults for the given app.
func BuildAppConfig(appID string) *app.AppConfig {
	acct := BuildAccount()
	return &app.AppConfig{
		AppID:       appID,
		CreatedByID: acct.ID,
		Status:      app.AppConfigStatusActive,
		StatusV2:    app.NewCompositeStatus(context.Background(), app.Status(app.AppConfigStatusActive)),
		CLIVersion:  "development",
	}
}

// CreateBareAppConfig persists a minimal AppConfig shell for the given app to the database.
// OrgID and CreatedByID are populated by the BeforeCreate hook from context.
// Use CreateAppConfig for a fully-populated config.
func (s *Seeder) CreateBareAppConfig(ctx context.Context, t *testing.T, appID string) *app.AppConfig {
	cfg := &app.AppConfig{
		AppID:      appID,
		Status:     app.AppConfigStatusActive,
		StatusV2:   app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusActive)),
		CLIVersion: "development",
	}
	res := s.db.WithContext(ctx).Create(cfg)
	require.NoError(t, res.Error)
	return cfg
}

// CreateAppSandboxConfig persists an AppSandboxConfig with a nested PublicGitVCSConfig.
func (s *Seeder) CreateAppSandboxConfig(ctx context.Context, t *testing.T, appID, appConfigID string) *app.AppSandboxConfig {
	sandbox := &app.AppSandboxConfig{
		AppID:            appID,
		AppConfigID:      appConfigID,
		TerraformVersion: "latest",
		PublicGitVCSConfig: &app.PublicGitVCSConfig{
			Repo:      "https://github.com/fake/terraform-component",
			Branch:    "main",
			Directory: "/",
		},
	}
	require.NoError(t, s.db.WithContext(ctx).Create(sandbox).Error)
	return sandbox
}

// CreateAppRunnerConfig persists an AppRunnerConfig with type aws.
func (s *Seeder) CreateAppRunnerConfig(ctx context.Context, t *testing.T, appID, appConfigID string) *app.AppRunnerConfig {
	runner := &app.AppRunnerConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
		Type:        app.AppRunnerTypeAWS,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(runner).Error)
	return runner
}

// CreateAppInputConfig persists an AppInputConfig with one AppInputGroup and one AppInput.
// Three sequential creates are required due to the FK chain: AppInputConfig -> AppInputGroup -> AppInput.
func (s *Seeder) CreateAppInputConfig(ctx context.Context, t *testing.T, appID, appConfigID string) *app.AppInputConfig {
	inputCfg := &app.AppInputConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(inputCfg).Error)

	group := &app.AppInputGroup{
		AppInputConfigID: inputCfg.ID,
		Name:             "general",
		DisplayName:      "General",
		Description:      "General configuration",
	}
	require.NoError(t, s.db.WithContext(ctx).Create(group).Error)

	input := &app.AppInput{
		AppInputConfigID: inputCfg.ID,
		AppInputGroupID:  group.ID,
		Name:             "region",
		DisplayName:      "AWS Region",
		Description:      "The AWS region to deploy into",
		Type:             app.AppInputTypeString,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(input).Error)

	inputCfg.AppInputGroups = []app.AppInputGroup{*group}
	inputCfg.AppInputs = []app.AppInput{*input}
	return inputCfg
}

// CreateAppSecretsConfig persists an empty AppSecretsConfig.
func (s *Seeder) CreateAppSecretsConfig(ctx context.Context, t *testing.T, appID, appConfigID string) *app.AppSecretsConfig {
	cfg := &app.AppSecretsConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(cfg).Error)
	return cfg
}

// CreateAppPermissionsConfig persists an empty AppPermissionsConfig.
func (s *Seeder) CreateAppPermissionsConfig(ctx context.Context, t *testing.T, appID, appConfigID string) *app.AppPermissionsConfig {
	cfg := &app.AppPermissionsConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(cfg).Error)
	return cfg
}

// CreateAppPoliciesConfig persists an empty AppPoliciesConfig.
func (s *Seeder) CreateAppPoliciesConfig(ctx context.Context, t *testing.T, appID, appConfigID string) *app.AppPoliciesConfig {
	cfg := &app.AppPoliciesConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(cfg).Error)
	return cfg
}

// CreateAppBreakGlassConfig persists an empty AppBreakGlassConfig.
func (s *Seeder) CreateAppBreakGlassConfig(ctx context.Context, t *testing.T, appID, appConfigID string) *app.AppBreakGlassConfig {
	cfg := &app.AppBreakGlassConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(cfg).Error)
	return cfg
}

// CreateAppStackConfig persists an AppStackConfig of type aws-cloudformation.
func (s *Seeder) CreateAppStackConfig(ctx context.Context, t *testing.T, appID, appConfigID string) *app.AppStackConfig {
	cfg := &app.AppStackConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
		Type:        app.StackTypeAWS,
		Name:        fmt.Sprintf("stack-%s", domains.NewAppID()),
		Description: "test stack config",
	}
	require.NoError(t, s.db.WithContext(ctx).Create(cfg).Error)
	return cfg
}

// CreateHelmComponentConfigConnection persists a ComponentConfigConnection with a HelmComponentConfig.
func (s *Seeder) CreateHelmComponentConfigConnection(ctx context.Context, t *testing.T, componentID, appConfigID string) *app.ComponentConfigConnection {
	ccc := &app.ComponentConfigConnection{
		ComponentID: componentID,
		AppConfigID: appConfigID,
		HelmComponentConfig: &app.HelmComponentConfig{
			ChartName: "my-chart",
			HelmConfig: &app.HelmConfig{
				ChartName: "my-chart",
				Namespace: "default",
				Values:    map[string]*string{},
			},
			PublicGitVCSConfig: &app.PublicGitVCSConfig{
				Repo:      "https://github.com/fake/helm-component",
				Branch:    "main",
				Directory: "/charts",
			},
		},
	}
	require.NoError(t, s.db.WithContext(ctx).Create(ccc).Error)
	return ccc
}

// CreateTerraformComponentConfigConnection persists a ComponentConfigConnection with a TerraformModuleComponentConfig.
func (s *Seeder) CreateTerraformComponentConfigConnection(ctx context.Context, t *testing.T, componentID, appConfigID string) *app.ComponentConfigConnection {
	ccc := &app.ComponentConfigConnection{
		ComponentID: componentID,
		AppConfigID: appConfigID,
		TerraformModuleComponentConfig: &app.TerraformModuleComponentConfig{
			Version: "latest",
			PublicGitVCSConfig: &app.PublicGitVCSConfig{
				Repo:      "https://github.com/fake/terraform-component",
				Branch:    "main",
				Directory: "/",
			},
		},
	}
	require.NoError(t, s.db.WithContext(ctx).Create(ccc).Error)
	return ccc
}

// CreateDockerBuildComponentConfigConnection persists a ComponentConfigConnection with a DockerBuildComponentConfig.
func (s *Seeder) CreateDockerBuildComponentConfigConnection(ctx context.Context, t *testing.T, componentID, appConfigID string) *app.ComponentConfigConnection {
	ccc := &app.ComponentConfigConnection{
		ComponentID: componentID,
		AppConfigID: appConfigID,
		DockerBuildComponentConfig: &app.DockerBuildComponentConfig{
			Dockerfile: "Dockerfile",
			PublicGitVCSConfig: &app.PublicGitVCSConfig{
				Repo:      "https://github.com/fake/docker-component",
				Branch:    "main",
				Directory: "/",
			},
		},
	}
	require.NoError(t, s.db.WithContext(ctx).Create(ccc).Error)
	return ccc
}

// CreateKubernetesManifestComponentConfigConnection persists a ComponentConfigConnection with a KubernetesManifestComponentConfig.
func (s *Seeder) CreateKubernetesManifestComponentConfigConnection(ctx context.Context, t *testing.T, componentID, appConfigID string) *app.ComponentConfigConnection {
	ccc := &app.ComponentConfigConnection{
		ComponentID: componentID,
		AppConfigID: appConfigID,
		KubernetesManifestComponentConfig: &app.KubernetesManifestComponentConfig{
			Manifest:  "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n",
			Namespace: "default",
		},
	}
	require.NoError(t, s.db.WithContext(ctx).Create(ccc).Error)
	return ccc
}

// CreateExternalImageComponentConfigConnection persists a ComponentConfigConnection with an ExternalImageComponentConfig.
func (s *Seeder) CreateExternalImageComponentConfigConnection(ctx context.Context, t *testing.T, componentID, appConfigID string) *app.ComponentConfigConnection {
	ccc := &app.ComponentConfigConnection{
		ComponentID: componentID,
		AppConfigID: appConfigID,
		ExternalImageComponentConfig: &app.ExternalImageComponentConfig{
			ImageURL: "nginx",
			Tag:      "latest",
		},
	}
	require.NoError(t, s.db.WithContext(ctx).Create(ccc).Error)
	return ccc
}

// CreateJobComponentConfigConnection persists a ComponentConfigConnection with a JobComponentConfig.
func (s *Seeder) CreateJobComponentConfigConnection(ctx context.Context, t *testing.T, componentID, appConfigID string) *app.ComponentConfigConnection {
	ccc := &app.ComponentConfigConnection{
		ComponentID: componentID,
		AppConfigID: appConfigID,
		JobComponentConfig: &app.JobComponentConfig{
			ImageURL: "ubuntu",
			Tag:      "latest",
			Cmd:      pq.StringArray{"echo", "hello"},
		},
	}
	require.NoError(t, s.db.WithContext(ctx).Create(ccc).Error)
	return ccc
}

// CreateActionWorkflow persists an ActionWorkflow definition scoped to the given app.
func (s *Seeder) CreateActionWorkflow(ctx context.Context, t *testing.T, appID string) *app.ActionWorkflow {
	wf := &app.ActionWorkflow{
		AppID:  appID,
		Name:   fmt.Sprintf("action-%s", domains.NewActionWorkflowID()),
		Status: app.ActionWorkflowStatusActive,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(wf).Error)
	return wf
}

// CreateActionWorkflowConfig persists an ActionWorkflowConfig with one manual trigger and one step.
// Three sequential creates are required: ActionWorkflowConfig -> ActionWorkflowTriggerConfig -> ActionWorkflowStepConfig.
func (s *Seeder) CreateActionWorkflowConfig(ctx context.Context, t *testing.T, appID, appConfigID, actionWorkflowID string) *app.ActionWorkflowConfig {
	awc := &app.ActionWorkflowConfig{
		AppID:            appID,
		AppConfigID:      appConfigID,
		ActionWorkflowID: actionWorkflowID,
		Timeout:          5 * time.Minute,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(awc).Error)

	trigger := &app.ActionWorkflowTriggerConfig{
		AppID:                  appID,
		AppConfigID:            appConfigID,
		ActionWorkflowConfigID: awc.ID,
		Type:                   app.ActionWorkflowTriggerTypeManual,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(trigger).Error)

	step := &app.ActionWorkflowStepConfig{
		AppID:                  appID,
		AppConfigID:            appConfigID,
		ActionWorkflowConfigID: awc.ID,
		Name:                   "step-1",
		Command:                "echo hello",
		InlineContents:         "echo hello",
		Idx:                    0,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(step).Error)

	awc.Triggers = []app.ActionWorkflowTriggerConfig{*trigger}
	awc.Steps = []app.ActionWorkflowStepConfig{*step}
	return awc
}

// CreateAppConfig creates a fully-populated AppConfig mirroring what a real CLI sync produces:
// all required configs, all optional configs, one component of each type, and one action workflow.
func (s *Seeder) CreateAppConfig(ctx context.Context, t *testing.T, appID string) *app.AppConfig {
	// Parent shell
	cfg := s.CreateBareAppConfig(ctx, t, appID)

	// Required configs
	s.CreateAppSandboxConfig(ctx, t, appID, cfg.ID)
	s.CreateAppRunnerConfig(ctx, t, appID, cfg.ID)
	s.CreateAppInputConfig(ctx, t, appID, cfg.ID)

	// Optional configs
	s.CreateAppSecretsConfig(ctx, t, appID, cfg.ID)
	s.CreateAppPermissionsConfig(ctx, t, appID, cfg.ID)
	s.CreateAppPoliciesConfig(ctx, t, appID, cfg.ID)
	s.CreateAppBreakGlassConfig(ctx, t, appID, cfg.ID)
	s.CreateAppStackConfig(ctx, t, appID, cfg.ID)

	// One component of each type
	helmComp := s.CreateComponent(ctx, t, appID, app.ComponentTypeHelmChart)
	tfComp := s.CreateComponent(ctx, t, appID, app.ComponentTypeTerraformModule)
	dockerComp := s.CreateComponent(ctx, t, appID, app.ComponentTypeDockerBuild)
	k8sComp := s.CreateComponent(ctx, t, appID, app.ComponentTypeKubernetesManifest)
	extImageComp := s.CreateComponent(ctx, t, appID, app.ComponentTypeExternalImage)
	jobComp := s.CreateComponent(ctx, t, appID, app.ComponentTypeJob)

	helmCCC := s.CreateHelmComponentConfigConnection(ctx, t, helmComp.ID, cfg.ID)
	tfCCC := s.CreateTerraformComponentConfigConnection(ctx, t, tfComp.ID, cfg.ID)
	dockerCCC := s.CreateDockerBuildComponentConfigConnection(ctx, t, dockerComp.ID, cfg.ID)
	k8sCCC := s.CreateKubernetesManifestComponentConfigConnection(ctx, t, k8sComp.ID, cfg.ID)
	extImageCCC := s.CreateExternalImageComponentConfigConnection(ctx, t, extImageComp.ID, cfg.ID)
	jobCCC := s.CreateJobComponentConfigConnection(ctx, t, jobComp.ID, cfg.ID)

	// One action workflow
	wf := s.CreateActionWorkflow(ctx, t, appID)
	s.CreateActionWorkflowConfig(ctx, t, appID, cfg.ID, wf.ID)

	// Finalize: set status active and populate ComponentIDs (mirrors real sync finish)
	componentIDs := pq.StringArray{
		helmComp.ID, tfComp.ID, dockerComp.ID,
		k8sComp.ID, extImageComp.ID, jobComp.ID,
	}
	activeStatusV2 := app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusActive))
	cfg.Status = app.AppConfigStatusActive
	cfg.StatusV2 = activeStatusV2
	cfg.ComponentIDs = componentIDs
	require.NoError(t, s.db.WithContext(ctx).Model(cfg).Updates(map[string]any{
		"status":        app.AppConfigStatusActive,
		"status_v2":     activeStatusV2,
		"component_ids": componentIDs,
	}).Error)

	// Do this at the end or GORM makes a bunch of duplicates for some reason.
	cfg.ComponentConfigConnections = []app.ComponentConfigConnection{
		*helmCCC, *tfCCC, *dockerCCC, *k8sCCC, *extImageCCC, *jobCCC,
	}

	return cfg
}
