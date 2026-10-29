package config

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/oci/signature"
)

type nestedFieldContract struct {
	typ    reflect.Type
	fields string
}

var nestedFieldContracts = []nestedFieldContract{
	{reflect.TypeFor[Component](), "Source:ingress Type:constrained-type Name:identity SourceFile:provenance VarName:var_name Labels:labels Dependencies:dependencies OperationRoles:nested Toggleable:toggleable DefaultEnabled:default_enabled KubernetesContext:kubernetes_context HelmChart:nested TerraformModule:nested DockerBuild:nested Job:nested ExternalImage:nested KubernetesManifest:nested Pulumi:nested References:derived Checksum:derived"},
	{reflect.TypeFor[HelmChartComponentConfig](), "ChartName:chart_name ValuesMap:values ValuesFiles:nested PublicRepo:nested ConnectedRepo:nested HelmRepo:nested Namespace:namespace StorageDriver:storage_driver TakeOwnership:take_ownership SkipCRDs:skip_crds DriftSchedule:drift_schedule Health:nested BuildTimeout:build_timeout DeployTimeout:deploy_timeout MaxAutoRetries:max_auto_retries SkipNoops:skip_noops AutoApproveOnPoliciesPassing:auto_approve_on_policies_passing Values:legacy-rejected"},
	{reflect.TypeFor[TerraformModuleComponentConfig](), "TerraformVersion:terraform_version EnvVarMap:env_vars VarsMap:vars VariablesFiles:nested PublicRepo:nested ConnectedRepo:nested DriftSchedule:drift_schedule BuildTimeout:build_timeout DeployTimeout:deploy_timeout MaxAutoRetries:max_auto_retries SkipNoops:skip_noops AutoApproveOnPoliciesPassing:auto_approve_on_policies_passing Teardown:nested Variables:legacy-rejected EnvVars:legacy-rejected"},
	{reflect.TypeFor[DockerBuildComponentConfig](), "Dockerfile:dockerfile EnvVarMap:env_vars PublicRepo:nested ConnectedRepo:nested BuildTimeout:build_timeout DeployTimeout:deploy_timeout MaxAutoRetries:max_auto_retries SkipNoops:skip_noops AutoApproveOnPoliciesPassing:auto_approve_on_policies_passing"},
	{reflect.TypeFor[ExternalImageComponentConfig](), "AWSECRImageConfig:nested GCPGARImageConfig:nested AzureACRImageConfig:nested PublicImageConfig:nested Verification:nested BuildTimeout:build_timeout DeployTimeout:deploy_timeout"},
	{reflect.TypeFor[KubernetesManifestComponentConfig](), "Manifest:manifest Kustomize:nested PublicRepo:nested ConnectedRepo:nested Namespace:namespace DriftSchedule:drift_schedule Health:nested BuildTimeout:build_timeout DeployTimeout:deploy_timeout MaxAutoRetries:max_auto_retries SkipNoops:skip_noops AutoApproveOnPoliciesPassing:auto_approve_on_policies_passing"},
	{reflect.TypeFor[JobComponentConfig](), "ImageURL:image_url Tag:tag Cmd:cmd EnvVarMap:env_vars Args:args BuildTimeout:build_timeout DeployTimeout:deploy_timeout EnvVars:legacy-rejected"},
	{reflect.TypeFor[PulumiComponentConfig](), "Runtime:runtime PulumiVersion:pulumi_version ConfigMap:config EnvVarMap:env_vars PublicRepo:nested ConnectedRepo:nested DriftSchedule:drift_schedule BuildTimeout:build_timeout DeployTimeout:deploy_timeout MaxAutoRetries:max_auto_retries SkipNoops:skip_noops AutoApproveOnPoliciesPassing:auto_approve_on_policies_passing"},
	{reflect.TypeFor[HelmValuesFile](), "Source:legacy-alias Contents:contents Path:path"},
	{reflect.TypeFor[TerraformVariablesFile](), "Contents:contents"},
	{reflect.TypeFor[HelmRepoConfig](), "RepoURL:repo_url Chart:chart Version:version"},
	{reflect.TypeFor[PublicRepoConfig](), "Repo:repo Directory:directory Branch:branch"},
	{reflect.TypeFor[ConnectedRepoConfig](), "Repo:repo Directory:directory Branch:branch"},
	{reflect.TypeFor[KustomizeConfig](), "Path:path Patches:patches EnableHelm:enable_helm LoadRestrictor:load_restrictor"},
	{reflect.TypeFor[ComponentHealthConfig](), "Enabled:enabled StabilizationWindow:stabilization_window BlockDeploy:block_deploy Probes:nested RequiredChecks:required_checks"},
	{reflect.TypeFor[ComponentHealthProbeConfig](), "Type:type Name:name URL:url Command:command"},
	{reflect.TypeFor[ComponentTeardownConfig](), "Retain:retain"},
	{reflect.TypeFor[AWSECRConfig](), "IAMRoleARN:iam_role_arn AWSRegion:region ImageURL:image_url Tag:tag UpdatePolicy:update_policy"},
	{reflect.TypeFor[GCPGARConfig](), "GCPProjectID:gcp_project_id GCPRegion:region ImageURL:image_url Tag:tag ServiceAccountEmail:service_account_email WorkloadIdentityProvider:workload_identity_provider UpdatePolicy:update_policy"},
	{reflect.TypeFor[AzureACRConfig](), "ImageURL:image_url Tag:tag RegistryURL:registry_url TenantID:tenant_id ClientID:client_id ClientSecretName:client_secret_name ClientCertificateName:client_certificate_name UpdatePolicy:update_policy"},
	{reflect.TypeFor[PublicImageConfig](), "ImageURL:image_url Tag:tag UpdatePolicy:update_policy"},
	{reflect.TypeFor[signature.Verification](), "RequireSignature:require_signature Authorities:nested"},
	{reflect.TypeFor[signature.Authority](), "Type:constrained-type Issuer:issuer Subject:subject SubjectRegexp:subject_regexp PublicKey:public_key"},
	{reflect.TypeFor[AppSandboxConfig](), "Source:ingress Type:type TerraformVersion:terraform_version Runtime:runtime PulumiVersion:pulumi_version PulumiConfig:pulumi_config ConnectedRepo:nested PublicRepo:nested DriftSchedule:drift_schedule EnvVarMap:env_vars VarsMap:vars VariablesFiles:nested OperationRoles:nested MaxAutoRetries:max_auto_retries SkipNoops:skip_noops AutoApproveOnPoliciesPassing:auto_approve_on_policies_passing References:derived"},
	{reflect.TypeFor[AppRunnerConfig](), "Source:ingress RunnerType:runner_type EnvVarMap:env_vars HelmDriver:helm_driver InitScriptURL:init_script_url PhoneHomeScriptURL:phone_home_script_url InstanceType:instance_type ImageAutoUpdate:image_auto_update ImageID:image_id RunnerAPIURL:runner_api_url PublicAPIURL:public_api_url EnvVars:legacy-rejected"},
	{reflect.TypeFor[AppInputConfig](), "Inputs:nested Groups:nested Source:ingress Sources:ingress"},
	{reflect.TypeFor[AppInput](), "Name:identity DisplayName:display_name Description:description Group:group Default:default Required:required Sensitive:sensitive Type:type Internal:legacy-ignored UserConfigurable:user_configurable"},
	{reflect.TypeFor[AppInputGroup](), "Name:identity Description:description DisplayName:display_name"},
	{reflect.TypeFor[SecretsConfig](), "Secrets:nested"},
	{reflect.TypeFor[AppSecret](), "Name:identity DisplayName:display_name Description:description Required:required AutoGenerate:auto_generate Format:format Default:default KubernetesSync:kubernetes_sync KubernetesSecretNamespace:kubernetes_secret_namespace KubernetesSecretName:kubernetes_secret_name KubernetesSyncTargets:nested"},
	{reflect.TypeFor[KubernetesSyncTarget](), "Namespaces:namespaces Name:name Key:key"},
	{reflect.TypeFor[PermissionsConfig](), "ProvisionRole:nested DeprovisionRole:nested MaintenanceRole:nested CustomRoles:nested Roles:ingress-normalized NamedPolicies:nested"},
	{reflect.TypeFor[AppAWSIAMRole](), "Type:type CloudPlatform:cloud_platform Name:name Description:description DisplayName:display_name Policies:nested PermissionsBoundary:permissions_boundary EnabledInStack:enabled_in_stack NamedPolicies:nested"},
	{reflect.TypeFor[AppAWSIAMPolicy](), "ManagedPolicyName:managed_policy_name Name:name Contents:contents GCPPermissions:gcp_permissions GCPPredefinedRole:gcp_predefined_role AzureActions:azure_actions AzureBuiltInRoles:azure_built_in_roles"},
	{reflect.TypeFor[NamedIAMPolicy](), "Name:identity Description:description Contents:contents SourceFile:provenance"},
	{reflect.TypeFor[NamedPolicyRef](), "Name:named_policies"},
	{reflect.TypeFor[PoliciesConfig](), "Policies:nested SourceFile:provenance"},
	{reflect.TypeFor[AppPolicy](), "Type:type Engine:constrained-engine Name:identity Contents:contents Components:components ApplyOnTeardown:apply_on_teardown SourceFile:provenance SourceLine:provenance"},
	{reflect.TypeFor[BreakGlass](), "Roles:nested"},
	{reflect.TypeFor[OperationRolesConfig](), "Type:singleton RuleMatrix:nested"},
	{reflect.TypeFor[OperationRoleRule](), "Principal:identity Operation:identity RoleName:role_name"},
	{reflect.TypeFor[EntityOperationRole](), "Operation:operation RoleName:role"},
	{reflect.TypeFor[ActionConfig](), "Name:identity Timeout:timeout Labels:labels Triggers:nested Steps:nested References:derived Dependencies:dependencies BreakGlassRole:break_glass_role Role:role EnableKubeConfig:enable_kube_config IsHealthcheck:is_healthcheck Image:image KubernetesContext:kubernetes_context"},
	{reflect.TypeFor[ActionStepConfig](), "Name:identity EnvVarMap:env_vars PublicRepo:nested ConnectedRepo:nested Command:command InlineContents:inline_contents References:derived"},
	{reflect.TypeFor[ActionTriggerConfig](), "Type:type Index:index CronSchedule:cron_schedule ComponentName:component_name"},
	{reflect.TypeFor[StackConfig](), "Type:type Name:name Description:description VPCNestedTemplateURL:vpc_nested_template_url RunnerNestedTemplateURL:runner_nested_template_url DeploymentScope:deployment_scope UIDefinition:ui_definition UIDefinitionContents:ui_definition_contents CustomNestedStacks:nested"},
	{reflect.TypeFor[CustomNestedStack](), "Name:identity TemplateURL:template_url Index:index Parameters:parameters Contents:contents ContentsHash:derived Status:derived TemplateSourceURL:derived"},
	{reflect.TypeFor[AppBranchConfig](), "Name:name ConnectedRepo:nested PublicRepo:nested InstallGroups:nested Preview:nested Run:nested PostDeployRunbooks:post_deploy_runbooks IgnoreChangesRegex:ignore_changes_regex SendStatusesOnIgnore:send_statuses_on_ignore"},
	{reflect.TypeFor[AppBranchInstallGroupConfig](), "Name:identity Order:order LabelSelector:label_selector Default:constrained-default AutoApproveOnPoliciesPassing:auto_approve_on_policies_passing DoNotDeploy:do_not_deploy"},
	{reflect.TypeFor[AppBranchPreviewConfig](), "Mode:mode InstallID:install_id InstallName:install_name LabelSelector:label_selector SetStatuses:set_statuses Comment:comment IgnoreDrafts:ignore_drafts React:react"},
	{reflect.TypeFor[AppBranchRunConfig](), "Mode:mode TagPrefix:tag_prefix GithubLabel:github_label"},
	{reflect.TypeFor[KubernetesContextsConfig](), "Contexts:nested"},
	{reflect.TypeFor[KubernetesContext](), "Name:identity Component:component"},
}

type nestedDiffFixture struct {
	input string
	owner string
}

func nestedDiffFixtures() []nestedDiffFixture {
	fixtures := []nestedDiffFixture{
		{"Components.0", "components/component.helm"},
		{"Components.0.HelmChart", "components/component.helm"},
		{"Components.1.TerraformModule", "components/component.terraform"},
		{"Components.2.DockerBuild", "components/component.docker"},
		{"Components.3.ExternalImage", "components/component.image"},
		{"Components.4.KubernetesManifest", "components/component.manifest"},
		{"Components.5.Job", "components/component.job"},
		{"Components.6.Pulumi", "components/component.pulumi"},
		{"Components.0.HelmChart.ValuesFiles.0", "components/component.helm/values_files"},
		{"Components.0.HelmChart.HelmRepo", "components/component.helm/helm_repo"},
		{"Components.1.TerraformModule.VariablesFiles.0", "components/component.terraform/var_file/var_file[0]"},
		{"Components.1.TerraformModule.Teardown", "components/component.terraform/teardown"},
		{"Components.4.KubernetesManifest.Kustomize", "components/component.manifest/kustomize"},
		{"Components.3.ExternalImage.AWSECRImageConfig", "components/component.image/aws_ecr"},
		{"Components.3.ExternalImage.GCPGARImageConfig", "components/component.image/gcp_gar"},
		{"Components.3.ExternalImage.AzureACRImageConfig", "components/component.image/azure_acr"},
		{"Components.3.ExternalImage.PublicImageConfig", "components/component.image/public_image"},
		{"Components.3.ExternalImage.Verification", "components/component.image/verification"},
		{"Components.3.ExternalImage.Verification.Authorities.0", "components/component.image/verification/authority.0"},
		{"Sandbox", "sandbox"}, {"Runner", "runner"},
		{"Sandbox.VariablesFiles.0", "sandbox/var_file/var_file[0]"},
		{"Sandbox.OperationRoles.0", "sandbox/operation_roles/operation_role.0"},
		{"Components.0.OperationRoles.0", "components/component.helm/operation_roles/operation_role.0"},
		{"Inputs", "inputs"}, {"Inputs.Inputs.0", "inputs/input.region"}, {"Inputs.Groups.0", "inputs/group.compute"},
		{"Secrets", "secrets"}, {"Secrets.Secrets.0", "secrets/secret.api_key"},
		{"Secrets.Secrets.0.KubernetesSyncTargets.0", "secrets/secret.api_key/kubernetes_sync_targets/target.0"},
		{"Permissions", "permissions"}, {"Permissions.NamedPolicies.0", "permissions/named_policy.shared"},
		{"Policies", "policies"}, {"Policies.Policies.0", "policies/policy.guard"},
		{"BreakGlass", "break_glass"}, {"OperationRoles", "operation_roles"},
		{"OperationRoles.RuleMatrix.0", "operation_roles/rule.nuon::component:*.deploy"},
		{"Actions.0", "actions/action.inspect"}, {"Actions.0.Steps.0", "actions/action.inspect/step.inspect"},
		{"Actions.0.Triggers.0", "actions/action.inspect/trigger.0"},
		{"Stack", "stack"}, {"Stack.CustomNestedStacks.0", "stack/custom_stack.storage"},
		{"Branch", "branch"}, {"Branches.0", "branches/branch.stable"},
		{"KubernetesContexts", "kubernetes_contexts"}, {"KubernetesContexts.Contexts.0", "kubernetes_contexts/kubernetes_context.compute"},
	}
	for _, parent := range []nestedDiffFixture{
		{"Components.0.HelmChart", "components/component.helm"},
		{"Components.1.TerraformModule", "components/component.terraform"},
		{"Components.2.DockerBuild", "components/component.docker"},
		{"Components.4.KubernetesManifest", "components/component.manifest"},
		{"Components.6.Pulumi", "components/component.pulumi"},
		{"Sandbox", "sandbox"}, {"Actions.0.Steps.0", "actions/action.inspect/step.inspect"},
		{"Branch", "branch"}, {"Branches.0", "branches/branch.stable"},
	} {
		fixtures = append(fixtures, nestedDiffFixture{parent.input + ".PublicRepo", parent.owner + "/public_repo"}, nestedDiffFixture{parent.input + ".ConnectedRepo", parent.owner + "/connected_repo"})
	}
	for _, parent := range []nestedDiffFixture{{"Components.0.HelmChart", "components/component.helm"}, {"Components.4.KubernetesManifest", "components/component.manifest"}} {
		fixtures = append(fixtures, nestedDiffFixture{parent.input + ".Health", parent.owner + "/health"}, nestedDiffFixture{parent.input + ".Health.Probes.0", parent.owner + "/health/probe.0"})
	}
	for _, parent := range []nestedDiffFixture{{"Branch", "branch"}, {"Branches.0", "branches/branch.stable"}} {
		fixtures = append(fixtures, nestedDiffFixture{parent.input + ".InstallGroups.0", parent.owner + "/install_groups/install_group.default"}, nestedDiffFixture{parent.input + ".Preview", parent.owner + "/preview"}, nestedDiffFixture{parent.input + ".Run", parent.owner + "/run"})
	}
	for _, parent := range []nestedDiffFixture{
		{"Permissions.ProvisionRole", "permissions/provision_role"}, {"Permissions.DeprovisionRole", "permissions/deprovision_role"},
		{"Permissions.MaintenanceRole", "permissions/maintenance_role"}, {"Permissions.CustomRoles.0", "permissions/custom_role.custom"},
		{"BreakGlass.Roles.0", "break_glass/role.emergency"},
	} {
		fixtures = append(fixtures, parent, nestedDiffFixture{parent.input + ".Policies.0", parent.owner + "/policies/policy.0"}, nestedDiffFixture{parent.input + ".NamedPolicies.0", parent.owner})
	}
	return fixtures
}

func nestedPublicKey(seed string) string {
	hash := sha256.Sum256([]byte(seed))
	key := ed25519.NewKeyFromSeed(hash[:]).Public()
	encoded, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		panic(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: encoded}))
}

func nestedBaseline() *AppConfig {
	repo := func() *PublicRepoConfig {
		return &PublicRepoConfig{Repo: "https://github.com/acme/app", Directory: "src", Branch: "main"}
	}
	health := func() *ComponentHealthConfig {
		return &ComponentHealthConfig{Probes: []ComponentHealthProbeConfig{{Type: "http", Name: "ready", URL: "https://example.com/ready"}}}
	}
	role := func(name, kind string) *AppAWSIAMRole {
		return &AppAWSIAMRole{Name: name, Type: kind, Description: "Read access", Policies: []AppAWSIAMPolicy{{Name: "read", ManagedPolicyName: "ReadOnlyAccess"}}, NamedPolicies: []NamedPolicyRef{{Name: "shared"}}}
	}
	branch := func() *AppBranchConfig {
		return &AppBranchConfig{Name: "stable", PublicRepo: repo(), InstallGroups: []AppBranchInstallGroupConfig{{Name: "default", Default: true}}, Preview: &AppBranchPreviewConfig{Mode: "build-only"}, Run: &AppBranchRunConfig{Mode: "push"}}
	}
	return &AppConfig{
		Components: ComponentList{
			{Name: "helm", Type: HelmChartComponentType, OperationRoles: []EntityOperationRole{{Operation: OperationDeploy, RoleName: "maintenance"}}, HelmChart: &HelmChartComponentConfig{ChartName: "example", Namespace: "default", PublicRepo: repo(), ValuesFiles: []HelmValuesFile{{Path: "values.yaml", Contents: "replicas: 1"}}, Health: health()}},
			{Name: "terraform", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0", PublicRepo: repo(), VariablesFiles: []TerraformVariablesFile{{Contents: "region = \"us-east-1\""}}, Teardown: &ComponentTeardownConfig{Retain: []string{"aws_s3_bucket.data"}}}},
			{Name: "docker", Type: DockerBuildComponentType, DockerBuild: &DockerBuildComponentConfig{Dockerfile: "Dockerfile", PublicRepo: repo()}},
			{Name: "image", Type: ExternalImageComponentType, ExternalImage: &ExternalImageComponentConfig{PublicImageConfig: &PublicImageConfig{ImageURL: "docker.io/acme/app", Tag: "1.0.0"}, Verification: &signature.Verification{RequireSignature: true, Authorities: []signature.Authority{{Type: signature.AuthorityTypeKeyless, Issuer: "https://token.actions.githubusercontent.com", Subject: "https://github.com/acme/app"}}}}},
			{Name: "manifest", Type: KubernetesManifestComponentType, KubernetesManifest: &KubernetesManifestComponentConfig{Namespace: "default", Kustomize: &KustomizeConfig{Path: "overlays/prod"}, PublicRepo: repo(), Health: health()}},
			{Name: "job", Type: JobComponentType, Job: &JobComponentConfig{ImageURL: "docker.io/acme/job", Tag: "1.0.0", Cmd: []string{"echo"}, Args: []string{"ready"}}},
			{Name: "pulumi", Type: PulumiComponentType, Pulumi: &PulumiComponentConfig{Runtime: "go", PulumiVersion: "3.100.0", PublicRepo: repo()}},
		},
		Sandbox:        &AppSandboxConfig{Type: "terraform", TerraformVersion: "1.6.0", Runtime: "go", PublicRepo: repo(), VariablesFiles: []TerraformVariablesFile{{Contents: "region = \"us-east-1\""}}, OperationRoles: []EntityOperationRole{{Operation: OperationProvision, RoleName: "provision"}}},
		Runner:         &AppRunnerConfig{RunnerType: "aws", HelmDriver: "secret"},
		Inputs:         &AppInputConfig{Groups: []AppInputGroup{{Name: "compute", Description: "Compute"}, {Name: "network", Description: "Network"}}, Inputs: []AppInput{{Name: "region", Type: "string", DisplayName: "Region", Description: "Cloud region", Group: "compute"}}},
		Secrets:        &SecretsConfig{Secrets: []*AppSecret{{Name: "api_key", Description: "API key", KubernetesSyncTargets: []*KubernetesSyncTarget{{Namespaces: []string{"default"}, Name: "api-key", Key: "value"}}}}},
		Permissions:    &PermissionsConfig{ProvisionRole: role("provision", "provision"), DeprovisionRole: role("deprovision", "deprovision"), MaintenanceRole: role("maintenance", "maintenance"), CustomRoles: []*AppAWSIAMRole{role("custom", "custom")}, NamedPolicies: []NamedIAMPolicy{{Name: "shared", Contents: `{"Version":"2012-10-17","Statement":[]}`}, {Name: "alternate", Contents: `{"Version":"2012-10-17","Statement":[]}`}}},
		BreakGlass:     &BreakGlass{Roles: []*AppAWSIAMRole{role("emergency", "custom")}},
		Policies:       &PoliciesConfig{Policies: []AppPolicy{{Name: "guard", Type: AppPolicyTypeTerraformModule, Engine: AppPolicyEngineOPA, Contents: "package nuon\ndeny := []"}}},
		OperationRoles: &OperationRolesConfig{Type: OperationRuleConfigTypeMatrix, RuleMatrix: []*OperationRoleRule{{Principal: "nuon::component:*", Operation: OperationDeploy, RoleName: "maintenance"}}},
		Actions:        []*ActionConfig{{Name: "inspect", Timeout: "5m", Steps: []*ActionStepConfig{{Name: "inspect", Command: "echo ready"}}, Triggers: []*ActionTriggerConfig{{Type: "manual"}}}},
		Stack:          &StackConfig{Type: "azure-bicep", Name: "acme", Description: "Install stack", CustomNestedStacks: []CustomNestedStack{{Name: "storage", Index: 0, TemplateURL: "https://example.com/storage.json", Contents: `{"resources":[]}`}}},
		Branch:         branch(), Branches: []*AppBranchConfig{branch()},
		KubernetesContexts: &KubernetesContextsConfig{Contexts: []*KubernetesContext{{Name: "compute", Component: "terraform"}, {Name: "alternate", Component: "pulumi"}}},
	}
}

func nestedValue(root any, path string) reflect.Value {
	v := reflect.ValueOf(root)
	for _, part := range strings.Split(path, ".") {
		for v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if v.Kind() == reflect.Slice {
			i, _ := strconv.Atoi(part)
			v = v.Index(i)
		} else {
			v = v.FieldByName(part)
		}
	}
	for v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	return v
}

func nestedFixtureConfig(f nestedDiffFixture, field string) *AppConfig {
	c := nestedBaseline()
	if strings.Contains(f.input, ".ConnectedRepo") {
		p := strings.TrimSuffix(f.input, ".ConnectedRepo")
		v := nestedValue(c, p)
		v.FieldByName("PublicRepo").SetZero()
		v.FieldByName("ConnectedRepo").Set(reflect.ValueOf(&ConnectedRepoConfig{Repo: "acme/app", Directory: "src", Branch: "main"}))
	}
	if f.input == "Actions.0.Steps.0.PublicRepo" {
		c.Actions[0].Steps[0].PublicRepo = &PublicRepoConfig{Repo: "https://github.com/acme/app", Directory: "scripts", Branch: "main"}
	}
	if f.input == "Components.0.HelmChart.HelmRepo" {
		c.Components[0].HelmChart.PublicRepo = nil
		c.Components[0].HelmChart.HelmRepo = &HelmRepoConfig{RepoURL: "https://example.com/charts", Chart: "example", Version: "1.0.0"}
	}
	e := c.Components[3].ExternalImage
	switch f.input {
	case "Components.3.ExternalImage.AWSECRImageConfig":
		e.PublicImageConfig = nil
		e.AWSECRImageConfig = &AWSECRConfig{IAMRoleARN: "arn:aws:iam::123456789012:role/pull", AWSRegion: "us-east-1", ImageURL: "123456789012.dkr.ecr.us-east-1.amazonaws.com/app", Tag: "1.0.0"}
	case "Components.3.ExternalImage.GCPGARImageConfig":
		e.PublicImageConfig = nil
		e.GCPGARImageConfig = &GCPGARConfig{GCPProjectID: "acme-prod", GCPRegion: "us-east1", ImageURL: "us-east1-docker.pkg.dev/acme-prod/app/api", Tag: "1.0.0"}
	case "Components.3.ExternalImage.AzureACRImageConfig":
		e.PublicImageConfig = nil
		e.AzureACRImageConfig = &AzureACRConfig{ImageURL: "acme.azurecr.io/app", RegistryURL: "acme.azurecr.io", Tag: "1.0.0"}
		if field == "TenantID" || field == "ClientID" || field == "ClientSecretName" || field == "ClientCertificateName" {
			e.AzureACRImageConfig.TenantID, e.AzureACRImageConfig.ClientID, e.AzureACRImageConfig.ClientSecretName = "acme-tenant", "acme-client", "api_key"
			if field == "ClientCertificateName" {
				e.AzureACRImageConfig.ClientSecretName, e.AzureACRImageConfig.ClientCertificateName = "", "api_key"
			}
		}
	}
	if f.input == "Components.4.KubernetesManifest" && field == "Manifest" {
		c.Components[4].KubernetesManifest.Kustomize, c.Components[4].KubernetesManifest.PublicRepo = nil, nil
		c.Components[4].KubernetesManifest.Manifest = "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: before"
	}
	if strings.HasSuffix(f.input, ".Health.Probes.0") && field == "Command" {
		nestedValue(c, f.input).Set(reflect.ValueOf(ComponentHealthProbeConfig{Type: "exec", Command: []string{"echo", "ready"}}))
	}
	if strings.HasSuffix(f.input, ".Verification.Authorities.0") {
		a := &e.Verification.Authorities[0]
		if field == "SubjectRegexp" {
			a.Subject, a.SubjectRegexp = "", "^https://github.com/acme/.*$"
		}
		if field == "PublicKey" {
			*a = signature.Authority{Type: signature.AuthorityTypePublicKey, PublicKey: nestedPublicKey("before")}
		}
	}
	if strings.HasSuffix(f.input, ".Preview") && field == "LabelSelector" {
		nestedValue(c, f.input).FieldByName("LabelSelector").Set(reflect.ValueOf(map[string]string{"env": "before"}))
	}
	if strings.HasSuffix(f.input, ".Run") {
		v := nestedValue(c, f.input)
		if field == "TagPrefix" {
			v.FieldByName("Mode").SetString("on_tag")
			v.FieldByName("TagPrefix").SetString("release")
		}
		if field == "GithubLabel" {
			v.FieldByName("Mode").SetString("on_github_label")
			v.FieldByName("GithubLabel").SetString("release")
			b := nestedValue(c, strings.TrimSuffix(f.input, ".Run"))
			b.FieldByName("PublicRepo").SetZero()
			b.FieldByName("ConnectedRepo").Set(reflect.ValueOf(&ConnectedRepoConfig{Repo: "acme/app", Branch: "main", Directory: "src"}))
		}
	}
	if strings.HasSuffix(f.input, ".Steps.0") && field == "InlineContents" {
		c.Actions[0].Steps[0].Command = ""
	}
	if strings.HasSuffix(f.input, ".Policies.0") && strings.Contains(f.input, "Role") {
		p := nestedValue(c, f.input).Addr().Interface().(*AppAWSIAMPolicy)
		role := nestedValue(c, strings.TrimSuffix(f.input, ".Policies.0")).Addr().Interface().(*AppAWSIAMRole)
		switch field {
		case "Contents":
			p.ManagedPolicyName = ""
			p.Contents = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:ListBucket","Resource":"*"}]}`
		case "GCPPermissions":
			p.ManagedPolicyName, role.CloudPlatform = "", "gcp"
			p.GCPPermissions = []string{"compute.instances.get"}
		case "GCPPredefinedRole":
			p.ManagedPolicyName, role.CloudPlatform, p.GCPPredefinedRole = "", "gcp", "roles/editor"
		case "AzureActions":
			p.ManagedPolicyName, role.CloudPlatform = "", "azure"
			p.AzureActions = []string{"Microsoft.Compute/virtualMachines/read"}
		case "AzureBuiltInRoles":
			p.ManagedPolicyName, role.CloudPlatform = "", "azure"
			p.AzureBuiltInRoles = []string{"Contributor"}
		}
	}
	if f.input == "Inputs.Groups.0" && field == "Name" {
		c.Inputs.Inputs[0].Group = "network"
	}
	if f.input == "Actions.0.Triggers.0" {
		if field == "CronSchedule" {
			c.Actions[0].Triggers[0].Type = "cron"
			c.Actions[0].Triggers[0].CronSchedule = "0 * * * *"
		}
		if field == "ComponentName" {
			c.Actions[0].Triggers[0].Type = "post-deploy-component"
			c.Actions[0].Triggers[0].ComponentName = "terraform"
		}
	}
	if f.input == "Stack" && field == "Type" {
		c.Stack.VPCNestedTemplateURL = "https://acme.s3.us-east-1.amazonaws.com/vpc.json"
		c.Stack.RunnerNestedTemplateURL = "https://acme.s3.us-east-1.amazonaws.com/runner.json"
	}
	if field == "Type" && nestedValue(c, f.input).Type() == reflect.TypeFor[AppAWSIAMRole]() {
		nestedValue(c, f.input).FieldByName("Type").SetString("maintenance")
	}
	if field == "Name" && nestedValue(c, f.input).Type() == reflect.TypeFor[AppAWSIAMRole]() {
		roleName := "custom"
		if strings.Contains(f.input, "CustomRoles") {
			roleName = "maintenance"
		}
		c.OperationRoles.RuleMatrix[0].RoleName = roleName
		c.Components[0].OperationRoles[0].RoleName = roleName
		c.Sandbox.OperationRoles[0].RoleName = roleName
	}
	if f.input == "Permissions.NamedPolicies.0" && field == "Name" {
		for _, role := range c.Permissions.allRoles() {
			role.NamedPolicies = []NamedPolicyRef{{Name: "alternate"}}
		}
		c.BreakGlass.Roles[0].NamedPolicies = []NamedPolicyRef{{Name: "alternate"}}
	}
	if f.input == "Sandbox" && (field == "Runtime" || field == "PulumiVersion" || field == "PulumiConfig") {
		c.Sandbox.Type = "pulumi"
	}
	if field == "PostDeployRunbooks" {
		c.Runbooks = []*RunbookConfig{{Name: "audit", Steps: []*RunbookStepConfig{{Name: "check", Type: RunbookStepTypeAction, Command: "echo ready"}}}}
	}
	return c
}

func nestedEditValue(v reflect.Value, typ reflect.Type, field, word string, number int) {
	if v.Kind() == reflect.Pointer {
		v.Set(reflect.New(v.Type().Elem()))
		nestedEditValue(v.Elem(), typ, field, word, number)
		if field == "Enabled" || field == "ApplyOnTeardown" || field == "EnableKubeConfig" || field == "SetStatuses" || field == "Comment" || field == "IgnoreDrafts" || field == "React" {
			v.Elem().SetBool(false)
		}
		return
	}
	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(!v.Bool())
	case reflect.Int, reflect.Int64:
		v.SetInt(int64(number + 1))
	case reflect.Interface:
		v.Set(reflect.ValueOf(word))
	case reflect.Map:
		v.Set(reflect.ValueOf(map[string]string{"env": word}))
	case reflect.Slice:
		if v.Type() == reflect.TypeFor[[]refs.Ref]() {
			v.Set(reflect.ValueOf([]refs.Ref{{Type: refs.RefTypeInputs, Name: word}}))
			return
		}
		values := []string{word}
		switch field {
		case "Retain":
			values = []string{"aws_s3_bucket." + word}
		case "Dependencies":
			values = []string{"terraform"}
		case "Components":
			values = []string{"terraform"}
		case "PostDeployRunbooks":
			values = []string{"audit"}
		case "Command":
			values = []string{"echo", word}
		case "GCPPermissions":
			values = []string{"storage.objects.get"}
		case "AzureActions":
			values = []string{"Microsoft.Storage/storageAccounts/read"}
		case "AzureBuiltInRoles":
			values = []string{"Reader"}
		}
		v.Set(reflect.ValueOf(values))
	case reflect.String:
		value := word
		switch field {
		case "Type":
			switch typ {
			case reflect.TypeFor[AppSandboxConfig]():
				value = "pulumi"
			case reflect.TypeFor[AppInput]():
				value = "json"
			case reflect.TypeFor[AppAWSIAMRole]():
				value = "custom"
			case reflect.TypeFor[AppPolicy]():
				value = string(AppPolicyTypeSandbox)
			case reflect.TypeFor[ActionTriggerConfig]():
				value = "post-provision"
			case reflect.TypeFor[StackConfig]():
				value = "aws-cloudformation"
			case reflect.TypeFor[ComponentHealthProbeConfig]():
				value = "tcp"
			}
		case "Runtime":
			value = "python"
		case "TerraformVersion":
			value = "1.7." + strconv.Itoa(number)
		case "PulumiVersion":
			value = "3.101." + strconv.Itoa(number)
		case "BuildTimeout", "DeployTimeout", "Timeout", "StabilizationWindow":
			value = strconv.Itoa(number+10) + "m"
		case "DriftSchedule", "CronSchedule":
			value = strconv.Itoa(number) + " * * * *"
		case "Format":
			value = "base64"
		case "Group":
			value = "network"
		case "CloudPlatform":
			value = "aws"
		case "RunnerType":
			value = "gcp"
		case "HelmDriver", "StorageDriver":
			value = "configmap"
		case "ImageID":
			value = "ami-0abc" + strconv.Itoa(number)
		case "InstanceType":
			value = "t3a.large"
		case "KubernetesContext":
			value = "alternate"
		case "Component", "ComponentName":
			value = "pulumi"
		case "RoleName", "Role", "BreakGlassRole":
			value = "custom"
		case "Principal":
			value = "nuon::component:helm"
		case "Operation":
			value = "teardown"
		case "DeploymentScope":
			value = "subscription"
		case "Mode":
			if typ == reflect.TypeFor[AppBranchRunConfig]() {
				value = "manual_only"
			} else {
				value = "none"
			}
		case "LoadRestrictor":
			value = "rootOnly"
		case "AWSRegion":
			value = "us-west-2"
		case "GCPRegion":
			value = "us-west1"
		case "Tag", "Version":
			value = "2.0." + strconv.Itoa(number)
		case "UpdatePolicy":
			value = "^2." + strconv.Itoa(number)
		case "URL", "RunnerAPIURL", "PublicAPIURL", "RepoURL", "InitScriptURL", "PhoneHomeScriptURL", "VPCNestedTemplateURL", "RunnerNestedTemplateURL", "TemplateURL", "Issuer":
			value = "https://example.com/" + word
		case "IAMRoleARN", "PermissionsBoundary":
			value = "arn:aws:iam::123456789012:role/" + word
		case "ServiceAccountEmail":
			value = word + "@acme-prod.iam.gserviceaccount.com"
		case "Repo":
			value = "acme/" + word
			if typ == reflect.TypeFor[PublicRepoConfig]() {
				value = "https://github.com/" + value
			}
		case "RegistryURL":
			value = word + ".azurecr.io"
		case "ManagedPolicyName":
			value = "AmazonS3ReadOnlyAccess"
		case "PublicKey":
			value = nestedPublicKey(word)
		case "GCPPredefinedRole":
			value = "roles/viewer"
		case "Manifest":
			value = "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: " + word
		case "UIDefinitionContents":
			value = `{"$schema":"https://schema.management.azure.com/schemas/0.1.2-preview/CreateUIDefinition.MultiVm.json#","handler":"Microsoft.Azure.CreateUIDef","parameters":{"basics":[],"steps":[],"outputs":{"name":"` + word + `"}}}`
		case "Contents":
			switch typ {
			case reflect.TypeFor[AppAWSIAMPolicy](), reflect.TypeFor[NamedIAMPolicy]():
				value = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"arn:aws:s3:::` + word + `/*"}]}`
			case reflect.TypeFor[CustomNestedStack]():
				value = `{"resources":[],"outputs":{"name":{"type":"string","value":"` + word + `"}}}`
			case reflect.TypeFor[TerraformVariablesFile]():
				value = `region = "` + word + `"`
			case reflect.TypeFor[HelmValuesFile]():
				value = "name: " + word
			case reflect.TypeFor[AppPolicy]():
				value = "package nuon\ndeny := [\"blocked\"]"
			}
		}
		if typ == reflect.TypeFor[NamedPolicyRef]() {
			value = "alternate"
		}
		v.SetString(value)
	default:
		panic(fmt.Sprintf("no authored edit for %v.%s", typ, field))
	}
}

func nestedDiffAt(t require.TestingT, root *diff.Diff, path string) *diff.Diff {
	for _, key := range strings.Split(path, "/") {
		var next *diff.Diff
		for _, child := range root.Children {
			if child.Key == key {
				next = child
				break
			}
		}
		require.NotNil(t, next, "missing child %s in %s", key, path)
		root = next
	}
	return root
}

func TestAppConfigDiff_NestedClassificationInventory(t *testing.T) {
	contracts := map[reflect.Type]map[string]string{}
	counts := map[string]int{}
	for _, contract := range nestedFieldContracts {
		fields := map[string]string{}
		for _, entry := range strings.Fields(contract.fields) {
			name, category, _ := strings.Cut(entry, ":")
			_, duplicate := fields[name]
			require.False(t, duplicate, "%v.%s", contract.typ, name)
			fields[name] = category
			switch category {
			case "nested", "derived", "provenance", "ingress", "ingress-normalized", "legacy-rejected", "legacy-alias", "legacy-ignored", "singleton", "constrained-default", "constrained-type", "constrained-engine":
				counts[category]++
			default:
				counts["semantic"]++
			}
		}
		require.Len(t, fields, contract.typ.NumField(), "%v needs an explicit decision for each field", contract.typ)
		for _, field := range reflect.VisibleFields(contract.typ) {
			require.Contains(t, fields, field.Name, "%v.%s is unclassified", contract.typ, field.Name)
		}
		contracts[contract.typ] = fields
	}
	covered := map[reflect.Type]bool{}
	oneFieldCases := 0
	for _, fixture := range nestedDiffFixtures() {
		v := nestedValue(nestedFixtureConfig(fixture, ""), fixture.input)
		require.Contains(t, contracts, v.Type(), fixture.input)
		covered[v.Type()] = true
		for _, category := range contracts[v.Type()] {
			switch category {
			case "nested", "ingress", "ingress-normalized", "legacy-rejected", "legacy-alias", "singleton", "constrained-default", "constrained-type", "constrained-engine":
			default:
				oneFieldCases++
			}
		}
	}
	for typ, fields := range contracts {
		require.True(t, covered[typ], "%v needs an AppConfig.Diff fixture", typ)
		for field, category := range fields {
			if category != "nested" {
				continue
			}
			f, _ := typ.FieldByName(field)
			child := f.Type
			for child.Kind() == reflect.Slice || child.Kind() == reflect.Pointer {
				child = child.Elem()
			}
			require.Contains(t, contracts, child, "%v.%s needs a nested inventory", typ, field)
		}
	}
	t.Logf("%d types; %d fixture paths; %d one-field regression cases; classifications: %v", len(contracts), len(nestedDiffFixtures()), oneFieldCases, counts)
}

func TestAppConfigDiff_NestedOneFieldRegressions(t *testing.T) {
	t.Parallel()
	contracts := map[reflect.Type]string{}
	for _, c := range nestedFieldContracts {
		contracts[c.typ] = c.fields
	}
	for _, fixture := range nestedDiffFixtures() {
		typ := nestedValue(nestedFixtureConfig(fixture, ""), fixture.input).Type()
		for _, entry := range strings.Fields(contracts[typ]) {
			field, expected, _ := strings.Cut(entry, ":")
			switch expected {
			case "nested", "ingress", "ingress-normalized", "legacy-rejected", "legacy-alias", "singleton", "constrained-default", "constrained-type", "constrained-engine":
				continue
			}
			t.Run(fixture.input+"/"+field, func(t *testing.T) {
				t.Parallel()
				old, next := nestedFixtureConfig(fixture, field), nestedFixtureConfig(fixture, field)
				before, after := nestedValue(old, fixture.input), nestedValue(next, fixture.input)
				nestedEditValue(after.FieldByName(field), typ, field, "edited", 7)
				if reflect.DeepEqual(before.FieldByName(field).Interface(), after.FieldByName(field).Interface()) {
					t.Fatalf("non-discriminating edit for %v.%s", typ, field)
				}
				nestedValidate(t, old)
				nestedValidate(t, next)
				for _, pair := range [][2]*AppConfig{{old, next}, {next, old}} {
					result := pair[1].Diff(pair[0])
					if expected == "derived" || expected == "provenance" || expected == "legacy-ignored" {
						require.False(t, result.DirectSummary().HasChanged, "%v.%s is not semantic", typ, field)
						continue
					}
					section, _, _ := strings.Cut(fixture.owner, "/")
					var changed []string
					for _, child := range result.Children {
						if child.DirectSummary().HasChanged {
							changed = append(changed, child.Key)
						}
					}
					require.Equal(t, []string{section}, changed, "%v.%s must be visible and section-local", typ, field)
					rekeys := field == "Name" && ((typ == reflect.TypeFor[AppBranchConfig]() && strings.HasPrefix(fixture.input, "Branches.")) || (typ == reflect.TypeFor[AppAWSIAMRole]() && (strings.Contains(fixture.input, "CustomRoles.") || strings.HasPrefix(fixture.input, "BreakGlass.Roles."))))
					if expected == "identity" || rekeys {
						separator := strings.LastIndex(fixture.owner, "/")
						parent, oldKey := fixture.owner[:separator+1], fixture.owner[separator+1:]
						newKey := ""
						if typ == reflect.TypeFor[OperationRoleRule]() {
							newKey = "rule." + after.FieldByName("Principal").String() + "." + after.FieldByName("Operation").String()
						} else {
							newKey = strings.TrimSuffix(oldKey, before.FieldByName("Name").String()) + after.FieldByName("Name").String()
						}
						for _, key := range []string{oldKey, newKey} {
							require.True(t, nestedDiffAt(t, result, parent+key).DirectSummary().HasChanged, "%v.%s must expose both membership edits", typ, field)
						}
						continue
					}
					owner := nestedDiffAt(t, result, fixture.owner)
					if typ == reflect.TypeFor[TerraformVariablesFile]() || typ == reflect.TypeFor[HelmValuesFile]() {
						require.True(t, owner.DirectSummary().HasChanged, "%s must expose %s", fixture.owner, field)
						continue
					}
					leaf := nestedDiffAt(t, owner, expected)
					require.True(t, leaf.DirectSummary().HasChanged, "%v.%s owns behavior at %s", typ, field, expected)
					for _, sibling := range owner.Children {
						if sibling.Key != expected {
							require.False(t, sibling.DirectSummary().HasChanged, "%v.%s leaked into %s", typ, field, sibling.Key)
						}
					}
				}
			})
		}
	}
}

func nestedValidate(t require.TestingT, c *AppConfig) {
	for _, component := range c.Components {
		require.NoError(t, component.Validate(), component.Name)
	}
	for _, policy := range c.Policies.Policies {
		if policy.Type == AppPolicyTypeKubernetesCluster {
			require.Equal(t, AppPolicyEngineKyverno, policy.Engine)
			var document map[string]any
			require.NoError(t, yaml.Unmarshal([]byte(policy.Contents), &document))
		} else {
			require.Equal(t, AppPolicyEngineOPA, policy.Engine)
			module, err := ast.ParseModule("policy.rego", policy.Contents)
			require.NoError(t, err)
			require.Equal(t, "data.nuon", module.Package.Path.String())
		}
	}
	require.NoError(t, c.Inputs.ValidateInputs())
	for _, secret := range c.Secrets.Secrets {
		require.NoError(t, secret.Validate())
	}
	require.NoError(t, c.Permissions.Validate())
	require.NoError(t, c.BreakGlass.Validate())
	require.NoError(t, c.OperationRoles.ValidateWithConfig(c.Components, c.Actions, c.Permissions, c.BreakGlass))
	require.NoError(t, c.Branch.Validate())
	for _, branch := range c.Branches {
		require.NoError(t, branch.Validate())
	}
	for _, runbook := range c.Runbooks {
		require.NoError(t, runbook.parse())
	}
	require.NoError(t, c.KubernetesContexts.Validate())
	require.NoError(t, c.Stack.parse())
	require.NoError(t, c.Runner.validateImageID())
}

func TestAppConfigDiff_NestedTypedDefaultProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		word := hegel.Draw(t, hegel.FromRegex(`[a-z]{1,8}`, true))
		number := strconv.FormatInt(hegel.Draw(t, hegel.Integers[int64](0, 1<<40)), 10)
		cases := []struct {
			kind string
			old  any
			next any
		}{
			{"number", json.Number(number), number},
			{"yaml", []string{word, "second"}, "[" + word + " second]"},
			{"yaml", map[string]string{"key": word}, "map[key:" + word + "]"},
		}
		for _, tc := range cases {
			nestedAssertTypedDefaultChange(t, tc.kind, tc.old, tc.next)
		}
	}, hegel.WithDerandomize(true), hegel.WithTestCases(32), hegel.WithDatabase(hegel.DatabaseDisabled()))
}

func TestAppConfigDiff_NestedTypedDefaultRegressions(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind string
		old  any
		next any
	}{
		{"bool_vs_string", "bool", true, "true"},
		{"nil_vs_string", "string", nil, "<nil>"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nestedAssertTypedDefaultChange(t, tc.kind, tc.old, tc.next)
		})
	}
}

func nestedAssertTypedDefaultChange(t require.TestingT, kind string, before, after any) {
	makeConfig := func(value any) *AppConfig {
		return &AppConfig{Inputs: &AppInputConfig{Inputs: []AppInput{{Name: "setting", Type: kind, Default: value}}}}
	}
	old, next := makeConfig(before), makeConfig(after)
	require.NoError(t, old.Inputs.ValidateInputs())
	require.NoError(t, next.Inputs.ValidateInputs())
	require.Equal(t, fmt.Sprint(before), fmt.Sprint(after))
	require.NoError(t, ValidateInputValueSyntax(kind, fmt.Sprint(before)))
	for _, pair := range [][2]*AppConfig{{old, next}, {next, old}} {
		leaf := nestedDiffAt(t, pair[1].Diff(pair[0]), "inputs/input.setting/default")
		require.True(t, leaf.DirectSummary().HasChanged, "%s authored defaults retain JSON type, not the builder's effective string", kind)
		require.NotEmpty(t, leaf.FormatChanged(""))
	}
}

func TestAppConfigDiff_NestedOmittedBooleans(t *testing.T) {
	// Rows and columns are nil, false, true. Inherited role settings retain authored intent.
	falseDefault := [3][3]bool{{false, false, true}, {false, false, true}, {true, true, false}}
	trueDefault := [3][3]bool{{false, true, false}, {true, false, true}, {false, true, false}}
	inherit := [3][3]bool{{false, true, true}, {true, false, true}, {true, true, false}}
	type boundary struct {
		fixture nestedDiffFixture
		field   string
		key     string
		changes [3][3]bool
	}
	cases := []boundary{
		{nestedDiffFixture{"Actions.0", "actions/action.inspect"}, "EnableKubeConfig", "enable_kube_config", trueDefault},
		{nestedDiffFixture{"Components.0", "components/component.helm"}, "DefaultEnabled", "default_enabled", falseDefault},
		{nestedDiffFixture{"Components.0", "components/component.helm"}, "Toggleable", "toggleable", falseDefault},
		{nestedDiffFixture{"Policies.Policies.0", "policies/policy.guard"}, "ApplyOnTeardown", "apply_on_teardown", trueDefault},
		{nestedDiffFixture{"Permissions.ProvisionRole", "permissions/provision_role"}, "EnabledInStack", "enabled_in_stack", inherit},
		{nestedDiffFixture{"BreakGlass.Roles.0", "break_glass/role.emergency"}, "EnabledInStack", "enabled_in_stack", inherit},
	}
	for _, fixture := range nestedDiffFixtures() {
		typ := nestedValue(nestedFixtureConfig(fixture, ""), fixture.input).Type()
		if typ == reflect.TypeFor[ComponentHealthConfig]() {
			cases = append(cases, boundary{fixture, "Enabled", "enabled", trueDefault}, boundary{fixture, "BlockDeploy", "block_deploy", falseDefault})
		}
		if typ == reflect.TypeFor[AppBranchPreviewConfig]() {
			for _, pair := range [][2]string{{"SetStatuses", "set_statuses"}, {"Comment", "comment"}, {"IgnoreDrafts", "ignore_drafts"}, {"React", "react"}} {
				cases = append(cases, boundary{fixture, pair[0], pair[1], trueDefault})
			}
		}
		if _, ok := typ.FieldByName("SkipNoops"); ok {
			for _, pair := range [][2]string{{"SkipNoops", "skip_noops"}, {"AutoApproveOnPoliciesPassing", "auto_approve_on_policies_passing"}} {
				cases = append(cases, boundary{fixture, pair[0], pair[1], falseDefault})
			}
		}
	}
	states := []struct {
		name  string
		value *bool
	}{{"nil", nil}, {"false", boolPtr(false)}, {"true", boolPtr(true)}}
	for _, tc := range cases {
		t.Run(tc.fixture.input+"/"+tc.field, func(t *testing.T) {
			for i, before := range states {
				for j, after := range states {
					t.Run(before.name+"_to_"+after.name, func(t *testing.T) {
						old, next := nestedFixtureConfig(tc.fixture, tc.field), nestedFixtureConfig(tc.fixture, tc.field)
						nestedValue(old, tc.fixture.input).FieldByName(tc.field).Set(reflect.ValueOf(before.value))
						nestedValue(next, tc.fixture.input).FieldByName(tc.field).Set(reflect.ValueOf(after.value))
						result := next.Diff(old)
						leaf := nestedDiffAt(t, result, tc.fixture.owner+"/"+tc.key)
						require.Equal(t, tc.changes[i][j], leaf.DirectSummary().HasChanged)
						if before.value == nil && after.value == nil {
							require.False(t, result.DirectSummary().HasChanged)
						}
						if tc.changes[i][j] {
							require.NotEmpty(t, leaf.FormatChanged(""))
						}
					})
				}
			}
		})
	}
}

func TestAppConfigDiff_NestedOrderedAndUnorderedProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		values := drawDistinctStrings(t, hegel.FromRegex(`[a-z]{1,8}`, true), 2, 5)
		permuted := drawPermutation(t, values)
		old, next := nestedBaseline(), nestedBaseline()
		old.Components[0].HelmChart.Health.Probes = []ComponentHealthProbeConfig{{Type: "exec", Command: values}}
		next.Components[0].HelmChart.Health.Probes = []ComponentHealthProbeConfig{{Type: "exec", Command: permuted}}
		leaf := nestedDiffAt(t, next.Diff(old), "components/component.helm/health/probe.0/command")
		require.Equal(t, !slices.Equal(values, permuted), leaf.DirectSummary().HasChanged)
		old, next = nestedBaseline(), nestedBaseline()
		old.Secrets.Secrets[0].KubernetesSyncTargets[0].Namespaces = values
		next.Secrets.Secrets[0].KubernetesSyncTargets[0].Namespaces = permuted
		leaf = nestedDiffAt(t, next.Diff(old), "secrets/secret.api_key/kubernetes_sync_targets/target.0/namespaces")
		require.False(t, leaf.DirectSummary().HasChanged)
	}, hegel.WithDerandomize(true), hegel.WithTestCases(32), hegel.WithDatabase(hegel.DatabaseDisabled()))
}

func TestAppConfigDiff_NestedDefaultInstallGroupSwitch(t *testing.T) {
	for _, branchFixture := range []nestedDiffFixture{{"Branch", "branch"}, {"Branches.0", "branches/branch.stable"}} {
		t.Run(branchFixture.input, func(t *testing.T) {
			old, next := nestedBaseline(), nestedBaseline()
			before := nestedValue(old, branchFixture.input).Addr().Interface().(*AppBranchConfig)
			after := nestedValue(next, branchFixture.input).Addr().Interface().(*AppBranchConfig)
			before.InstallGroups = []AppBranchInstallGroupConfig{{Name: "primary", Default: true}, {Name: "secondary", Default: false}}
			after.InstallGroups = []AppBranchInstallGroupConfig{{Name: "primary", Default: false}, {Name: "secondary", Default: true}}
			require.NoError(t, before.Validate())
			require.NoError(t, after.Validate())
			for _, pair := range [][2]*AppConfig{{old, next}, {next, old}} {
				result := pair[1].Diff(pair[0])
				for _, name := range []string{"primary", "secondary"} {
					leaf := nestedDiffAt(t, result, branchFixture.owner+"/install_groups/install_group."+name+"/default")
					require.True(t, leaf.DirectSummary().HasChanged)
				}
			}
		})
	}
}

func TestAppConfigDiff_NestedOmittedRetryBoundaries(t *testing.T) {
	zero, one, low, high, maximum := 0, 1, 3, 11, 20
	states := []struct {
		name      string
		value     *int
		effective int
	}{
		{"nil", nil, 0},
		{"zero", &zero, 0},
		{"one", &one, 1},
		{"three", &low, 3},
		{"eleven", &high, 11},
		{"maximum", &maximum, 20},
	}
	for _, fixture := range nestedDiffFixtures() {
		typ := nestedValue(nestedFixtureConfig(fixture, ""), fixture.input).Type()
		if _, ok := typ.FieldByName("MaxAutoRetries"); !ok {
			continue
		}
		t.Run(fixture.input, func(t *testing.T) {
			for _, before := range states {
				for _, after := range states {
					t.Run(before.name+"_to_"+after.name, func(t *testing.T) {
						old, next := nestedFixtureConfig(fixture, "MaxAutoRetries"), nestedFixtureConfig(fixture, "MaxAutoRetries")
						nestedValue(old, fixture.input).FieldByName("MaxAutoRetries").Set(reflect.ValueOf(before.value))
						nestedValue(next, fixture.input).FieldByName("MaxAutoRetries").Set(reflect.ValueOf(after.value))
						nestedValidate(t, old)
						nestedValidate(t, next)
						leaf := nestedDiffAt(t, next.Diff(old), fixture.owner+"/max_auto_retries")
						changed := before.effective != after.effective
						require.Equal(t, changed, leaf.DirectSummary().HasChanged)
						if changed {
							require.NotEmpty(t, leaf.FormatChanged(""))
						}
					})
				}
			}
		})
	}
}

func TestAppConfigDiff_NestedConstrainedTransitions(t *testing.T) {
	t.Run("helm_to_job", func(t *testing.T) {
		old, next := nestedBaseline(), nestedBaseline()
		next.Components[0].Type, next.Components[0].HelmChart = JobComponentType, nil
		next.Components[0].Job = &JobComponentConfig{ImageURL: "docker.io/acme/job", Tag: "1.0.0"}
		nestedValidate(t, old)
		nestedValidate(t, next)
		for _, pair := range [][2]*AppConfig{{old, next}, {next, old}} {
			require.True(t, nestedDiffAt(t, pair[1].Diff(pair[0]), "components/component.helm/type").DirectSummary().HasChanged)
		}
	})
	t.Run("keyless_to_public_key", func(t *testing.T) {
		old, next := nestedBaseline(), nestedBaseline()
		next.Components[3].ExternalImage.Verification.Authorities[0] = signature.Authority{Type: signature.AuthorityTypePublicKey, PublicKey: nestedPublicKey("public-key")}
		nestedValidate(t, old)
		nestedValidate(t, next)
		for _, pair := range [][2]*AppConfig{{old, next}, {next, old}} {
			owner := nestedDiffAt(t, pair[1].Diff(pair[0]), "components/component.image/verification/authority.0")
			for _, key := range []string{"type", "issuer", "subject", "public_key"} {
				require.True(t, nestedDiffAt(t, owner, key).DirectSummary().HasChanged)
			}
		}
	})
	t.Run("opa_to_kyverno", func(t *testing.T) {
		old, next := nestedBaseline(), nestedBaseline()
		next.Policies.Policies[0].Type, next.Policies.Policies[0].Engine = AppPolicyTypeKubernetesCluster, AppPolicyEngineKyverno
		next.Policies.Policies[0].Contents = "apiVersion: kyverno.io/v1\nkind: ClusterPolicy\nmetadata:\n  name: guard\nspec:\n  validationFailureAction: Enforce\n  rules: []"
		nestedValidate(t, old)
		nestedValidate(t, next)
		for _, pair := range [][2]*AppConfig{{old, next}, {next, old}} {
			owner := nestedDiffAt(t, pair[1].Diff(pair[0]), "policies/policy.guard")
			for _, key := range []string{"type", "engine", "contents"} {
				require.True(t, nestedDiffAt(t, owner, key).DirectSummary().HasChanged)
			}
		}
	})
}
