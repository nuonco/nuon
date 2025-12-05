package plan

import (
	"bytes"
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	azurecredentials "github.com/powertoolsdev/mono/pkg/azure/credentials"
	"github.com/powertoolsdev/mono/pkg/config"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
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
		return nil, errors.Wrap(err, "unable to get install sandbox run")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	l.Info("configuring environment variables to execute terraform run as")
	envVars := p.getSandboxRunEnvVars(appCfg)

	l.Info("configuring terraform variables to execute terraform run as")
	vars, err := p.getSandboxRunTerraformVars(appCfg, req.RootDomain)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get vars")
	}
	for k, v := range appCfg.SandboxConfig.Variables {
		vars[k] = v
	}

	state, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install state")
	}
	stateData, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state")
	}

	l.Info("rendering environment variables")
	if err := render.RenderMap(&envVars, stateData); err != nil {
		l.Error("error rendering environment vars",
			zap.Any("env-vars", envVars),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	if err := render.RenderStruct(&appCfg.SandboxConfig, stateData); err != nil {
		l.Error("error rendering config",
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}

	l.Info("rendering terraform variables")
	if err := render.RenderMap(&vars, stateData); err != nil {
		l.Error("error rendering terraform variables",
			zap.Any("vars", vars),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render variables")
	}

	l.Info("outputs vars", zap.Any("vars", vars))

	if err := render.RenderStruct(&appCfg.PoliciesConfig, stateData); err != nil {
		l.Error("error rendering policies",
			zap.Any("policies", appCfg.PoliciesConfig),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render policies")
	}

	l.Info("getting policies")
	policies, err := p.getPolicies(&appCfg.PoliciesConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get policies")
	}

	l.Info("fetching sandbox git source")
	gitSource, err := activities.AwaitGetSandboxRunGitSourceByAppConfigID(ctx, appCfg.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get sandbox run git source")
	}

	l.Info("getting auth")
	awsAuth, azureAuth, err := p.getAuth(stack.InstallStackOutputs, run)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get sandbox run auth")
	}

	runAuth := &awscredentials.Config{}
	if awsAuth != nil {
		runAuth = awsAuth
	}
	plan := &plantypes.SandboxRunPlan{
		AppID:       install.AppID,
		AppConfigID: install.AppConfigID,
		InstallID:   install.ID,

		Vars:      vars,
		EnvVars:   envVars,
		VarsFiles: appCfg.SandboxConfig.VariablesFiles,
		GitSource: gitSource,
		State:     state,
		Policies:  policies,

		LocalArchive: nil,

		TerraformBackend: &plantypes.TerraformBackend{
			WorkspaceID: install.InstallSandbox.TerraformWorkspace.ID,
		},

		AzureAuth: azureAuth,
		AWSAuth:   awsAuth,
		// TODO(ja): provide both auth types to hooks
		Hooks: &plantypes.TerraformDeployHooks{
			Enabled: true,
			EnvVars: envVars,
			RunAuth: *runAuth,
		},
	}

	org, err := activities.AwaitGetOrgByInstallID(ctx, install.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org")
	}
	if org.SandboxMode {
		pdcJSONByts := new(bytes.Buffer)
		if err := json.Compact(pdcJSONByts, []byte(FakeTerraformPlanDisplayContents)); err != nil {
			return nil, errors.Wrap(err, "unable to get json")
		}

		stJSONByts := new(bytes.Buffer)
		if err := json.Compact(stJSONByts, []byte(FakeTerraformStateJSON)); err != nil {
			return nil, errors.Wrap(err, "unable to get json")
		}

		plan.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
			Terraform: &plantypes.TerraformSandboxMode{
				WorkspaceID:         install.InstallSandbox.TerraformWorkspace.ID,
				StateJSON:           stJSONByts.Bytes(),
				PlanContents:        FakeTerraformPlanContents,
				PlanDisplayContents: pdcJSONByts.String(),
			},
			Outputs: p.getSandboxModeOutputs(*install, *stack),
		}
	}

	return plan, nil
}

func (p *Planner) getPolicies(cfg *app.AppPoliciesConfig) (map[string]string, error) {
	obj := make(map[string]string, 0)

	for idx, policy := range cfg.Policies {
		if policy.Type != config.AppPolicyType(config.AppPolicyTypeKubernetesClusterKyverno) {
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
	envVars := make(map[string]string, 0)
	for k, v := range generics.ToStringMap(appCfg.SandboxConfig.EnvVars) {
		envVars[k] = v
	}

	if appCfg.RunnerConfig.Type != app.AppRunnerTypeAWS {
		return envVars
	}

	envVars["AWS_REGION"] = "{{.nuon.install_stack.outputs.region}}"

	return envVars
}

func (p *Planner) getSandboxRunTerraformVars(appCfg *app.AppConfig, rootDomain string) (map[string]any, error) {
	vars := make(map[string]any, 0)

	for k, v := range generics.ToStringMap(appCfg.SandboxConfig.Variables) {
		vars[k] = v
	}

	if appCfg.RunnerConfig.Type != app.AppRunnerTypeAWS {
		return vars, nil
	}

	builtin := map[string]any{
		"vpc_id":                   "{{.nuon.install_stack.outputs.vpc_id}}",
		"nuon_id":                  "{{.nuon.install.id}}",
		"region":                   "{{.nuon.install_stack.outputs.region}}",
		"public_root_domain":       fmt.Sprintf("{{.nuon.install.id}}.%s", rootDomain),
		"internal_root_domain":     fmt.Sprintf("{{.nuon.install.id}}.internal.%s", rootDomain),
		"provision_iam_role_arn":   "{{.nuon.install_stack.outputs.provision_iam_role_arn}}",
		"deprovision_iam_role_arn": "{{.nuon.install_stack.outputs.deprovision_iam_role_arn}}",
		"maintenance_iam_role_arn": "{{.nuon.install_stack.outputs.maintenance_iam_role_arn}}",
		"tags": map[string]string{
			"NUON_INSTALL_ID": "{{.nuon.install.id}}",
		},
	}

	for k, v := range builtin {
		vars[k] = v
	}

	return vars, nil
}

func (p *Planner) getAuth(outputs app.InstallStackOutputs, run *app.InstallSandboxRun) (*awscredentials.Config, *azurecredentials.Config, error) {
	switch {
	case outputs.AWSStackOutputs != nil:
		awsOutputs := outputs.AWSStackOutputs
		roleARN := awsOutputs.ProvisionIAMRoleARN
		switch run.RunType {
		case app.SandboxRunTypeReprovision:
			roleARN = outputs.AWSStackOutputs.ProvisionIAMRoleARN
		case app.SandboxRunTypeDeprovision:
			roleARN = outputs.AWSStackOutputs.DeprovisionIAMRoleARN
		}

		return &awscredentials.Config{
			Region: outputs.AWSStackOutputs.Region,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				SessionName: fmt.Sprintf("sandbox-run-%s", run.ID),
				RoleARN:     roleARN,
			},
		}, nil, nil
	case outputs.AzureStackOutputs != nil:
		azureOutputs := outputs.AzureStackOutputs
		return nil, &azurecredentials.Config{
			ServicePrincipal: &azurecredentials.ServicePrincipalCredentials{
				SubscriptionID:       azureOutputs.SubscriptionID,
				SubscriptionTenantID: azureOutputs.SubscriptionTenantID,
			},
			UseDefault: true,
		}, nil
	}

	return nil, nil, errors.New("unable to get auth data from stack outputs")
}

// TODO(ja): flesh out sandbox mode for azure
func (p *Planner) getSandboxModeOutputs(install app.Install, stack app.InstallStack) map[string]any {
	switch {
	case stack.InstallStackOutputs.AzureStackOutputs != nil:
		return map[string]any{
			"namespaces": []string{
				"default",
				install.ID,
			},
			"location": stack.InstallStackOutputs.AzureStackOutputs.ResourceGroupLocation,
		}
	default:
		return map[string]any{
			"namespaces": []string{
				"default",
				install.ID,
			},
			"region": stack.InstallStackOutputs.AWSStackOutputs.Region,
			"karpenter": map[string]any{
				"instance_profile": map[string]any{
					"id":   "karpenter-instance-profile-id",
					"arn":  "arn:aws:iam::123456789012:instance-profile/karpenter-profile",
					"name": "karpenter-instance-profile",
				},
				"discovery_key":   "karpenter-discovery-key",
				"discovery_value": "karpenter-discovery-value",
			},
			"account": map[string]any{
				"id":     "123456789012",
				"region": stack.InstallStackOutputs.AWSStackOutputs.Region,
			},
			"cluster": map[string]any{
				"arn":                        "arn:aws:eks:us-west-2:123456789012:cluster/nuon-cluster",
				"certificate_authority_data": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREF2TVMwd0t3WURWUVFERXlRME4yVTEKWkRNeE5DMDROelk...",
				"endpoint":                   "https://A1B2C3D4E5F6.gr7.us-west-2.eks.amazonaws.com",
				"name":                       "nuon-cluster",
				"platform_version":           "eks.9",
				"status":                     "ACTIVE",
				"oidc_issuer_url":            "https://oidc.eks.us-west-2.amazonaws.com/id/A1B2C3D4E5F6",
				"oidc_provider_arn":          "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-west-2.amazonaws.com/id/A1B2C3D4E5F6",
				"oidc_provider":              "123456789012:oidc-provider/oidc.eks.us-west-2.amazonaws.com/id/A1B2C3D4E5F6",
				"cluster_security_group_id":  "sg-0abc123def456",
				"node_security_group_id":     "sg-0xyz789uvw456",
			},
			"vpc": map[string]any{
				"id":                         "vpc-0abc123def456",
				"arn":                        "arn:aws:ec2:us-west-2:123456789012:vpc/vpc-0abc123def456",
				"cidr":                       "10.0.0.0/16",
				"azs":                        []string{"us-west-2a", "us-west-2b", "us-west-2c"},
				"private_subnet_cidr_blocks": []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"},
				"private_subnet_ids":         []string{"subnet-0abc123def456", "subnet-0ghi789jkl012", "subnet-0mno345pqr678"},
				"public_subnet_cidr_blocks":  []string{"10.0.4.0/24", "10.0.5.0/24", "10.0.6.0/24"},
				"public_subnet_ids":          []string{"subnet-0stu901vwx234", "subnet-0yza567bcd890", "subnet-0efg123hij456"},
				"runner_subnet_id":           "subnet-0klm789nop012",
				"runner_subnet_cidr":         "10.0.7.0/24",
				"default_security_group_id":  "sg-0qrs345tuv678",
			},
			"ecr": map[string]any{
				"repository_url":  "123456789012.dkr.ecr.us-west-2.amazonaws.com/nuon-app",
				"repository_arn":  "arn:aws:ecr:us-west-2:123456789012:repository/nuon-app",
				"repository_name": "nuon-app",
				"registry_id":     "123456789012",
				"registry_url":    "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			},
			"nuon_dns": map[string]any{
				"enabled": true,
				"public_domain": map[string]any{
					"zone_id":     "Z1A2B3C4D5E6F7",
					"name":        "example.com",
					"nameservers": []string{"ns-1234.awsdns-12.org", "ns-567.awsdns-34.com", "ns-890.awsdns-56.net", "ns-1234.awsdns-78.co.uk"},
				},
				"internal_domain": map[string]any{
					"zone_id":     "Z8G9H0I1J2K3L4",
					"name":        "internal.example.com",
					"nameservers": []string{"ns-5678.awsdns-90.org", "ns-123.awsdns-12.com", "ns-456.awsdns-34.net", "ns-789.awsdns-56.co.uk"},
				},
				"alb_ingress_controller": map[string]any{
					"enabled":  true,
					"id":       "alb-ingress-controller",
					"chart":    "aws-load-balancer-controller",
					"revision": "1.4.7",
				},
				"external_dns": map[string]any{
					"enabled":  true,
					"id":       "external-dns",
					"chart":    "external-dns",
					"revision": "1.12.1",
				},
				"cert_manager": map[string]any{
					"enabled":  true,
					"id":       "cert-manager",
					"chart":    "cert-manager",
					"revision": "1.11.0",
				},
				"ingress_nginx": map[string]any{
					"enabled":  true,
					"id":       "ingress-nginx",
					"chart":    "ingress-nginx",
					"revision": "4.7.1",
				},
			},
		}
	}
}
