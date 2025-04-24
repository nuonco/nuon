package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createSandboxRunPlan(ctx workflow.Context, req *CreateSandboxRunPlanRequest) (*plantypes.SandboxRunPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	install, err := activities.AwaitGetByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}

	run, err := activities.AwaitGetSandboxRunByRunID(ctx, req.RunID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install run")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	l.Info("configuring environment variables to execute terraform run as")
	envVars := p.getSandboxRunEnvVars(appCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get env vars")
	}

	l.Info("configuring terraform variables to execute terraform run as")
	vars, err := p.getSandboxRunTerraformVars(appCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get vars")
	}

	for k, v := range appCfg.SandboxConfig.Variables {
		vars[k] = v
	}

	state, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})
	stateData, err := state.AsMap()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state")
	}

	l.Info("rendering environment variables")
	if err := render.RenderMap(&envVars, stateData); err != nil {
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	l.Info("rendering terraform variables")
	if err := render.RenderMap(&vars, stateData); err != nil {
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	l.Info("rendering policies")
	if err := render.RenderStruct(&appCfg.PoliciesConfig, stateData); err != nil {
		return nil, errors.Wrap(err, "unable to render policies")
	}

	policies, err := p.getPolicies(&appCfg.PoliciesConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get policies")
	}

	l.Info("fetching sandbox git source")
	gitSource, err := activities.AwaitGetSandboxRunGitSourceByAppConfigID(ctx, appCfg.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get sandbox run git source")
	}

	roleARN := stack.InstallStackOutputs.AWSStackOutputs.ProvisionIAMRoleARN
	if run.RunType == app.SandboxRunTypeReprovision {
		roleARN = stack.InstallStackOutputs.AWSStackOutputs.MaintenanceIAMRoleARN
	}
	if run.RunType == app.SandboxRunTypeDeprovision {
		roleARN = stack.InstallStackOutputs.AWSStackOutputs.MaintenanceIAMRoleARN
	}

	return &plantypes.SandboxRunPlan{
		AppID:       install.AppID,
		AppConfigID: install.AppConfigID,
		InstallID:   install.ID,

		Vars:      vars,
		EnvVars:   envVars,
		GitSource: gitSource,
		State:     state,
		Policies:  policies,

		LocalArchive: nil,

		TerraformBackend: &plantypes.TerraformBackend{
			WorkspaceID: install.InstallSandbox.TerraformWorkspace.ID,
		},

		AzureAuth: nil,
		AWSAuth: &awscredentials.Config{
			Region: stack.InstallStackOutputs.AWSStackOutputs.Region,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				SessionName: fmt.Sprintf("sandbox-run-%s", req.RunID),
				RoleARN:     roleARN,
			},
		},
		Hooks: &plantypes.TerraformDeployHooks{
			Enabled: true,
			EnvVars: envVars,
			RunAuth: awscredentials.Config{
				Region: stack.InstallStackOutputs.AWSStackOutputs.Region,
				AssumeRole: &awscredentials.AssumeRoleConfig{
					SessionName: fmt.Sprintf("sandbox-run-%s", req.RunID),
					RoleARN:     roleARN,
				},
			},
		},
	}, nil
}

func (p *Planner) getPolicies(cfg *app.AppPoliciesConfig) (map[string]string, error) {
	obj := make(map[string]string, 0)

	for idx, policy := range cfg.Policies {
		if policy.Type != app.AppPolicyType(app.AppPolicyTypeKubernetesClusterKyverno) {
			continue
		}

		var parseObj map[string]any
		if err := yaml.Unmarshal([]byte(policy.Contents), &parseObj); err != nil {
			return nil, errors.Wrap(err, "unable to parse yaml")
		}

		obj[fmt.Sprintf("%d.yaml", idx)] = policy.Contents
	}

	return obj, nil
}

func (p *Planner) getSandboxRunEnvVars(appCfg *app.AppConfig) map[string]string {
	if appCfg.RunnerConfig.Type != app.AppRunnerTypeAWS {
		return map[string]string{}
	}

	return map[string]string{
		"AWS_REGION": "{{.nuon.install_stack.outputs.region}}",
	}
}

func (p *Planner) getSandboxRunTerraformVars(appCfg *app.AppConfig) (map[string]any, error) {
	if appCfg.RunnerConfig.Type != app.AppRunnerTypeAWS {
		return map[string]any{}, nil
	}

	vars := map[string]any{
		"vpc_id":                   "{{.nuon.install_stack.outputs.vpc_id}}",
		"nuon_id":                  "{{.nuon.install.id}}",
		"region":                   "{{.nuon.install_stack.outputs.region}}",
		"public_root_domain":       "{{.nuon.install.id}}.nuon.run",
		"internal_root_domain":     "{{.nuon.install.id}}.internal.nuon.run",
		"provision_iam_role_arn":   "{{.nuon.install_stack.outputs.provision_iam_role_arn}}",
		"deprovision_iam_role_arn": "{{.nuon.install_stack.outputs.deprovision_iam_role_arn}}",
		"maintenance_iam_role_arn": "{{.nuon.install_stack.outputs.maintenance_iam_role_arn}}",
		"tags": map[string]string{
			"NUON_INSTALL_ID": "{{.nuon.install.id}}",
		},
	}

	return vars, nil
}
