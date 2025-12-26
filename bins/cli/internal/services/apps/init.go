package apps

import (
	"context"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/generator"
	"github.com/pkg/errors"
)

type ConfigGenParams struct {
	Path            string
	EnableDefaults  bool
	EnableComments  bool
	Overwrite       bool
	SkipNonRequired bool
}

func NewGen(params ConfigGenParams) *generator.ConfigGen {
	return generator.NewConfigGen(
		params.EnableDefaults,
		params.EnableComments,
		false,
		params.Overwrite,
		params.SkipNonRequired,
	)
}

type InitParams struct {
	TerraformVersion string
	AppName          string
	StackType        string
	RunnerType       string
	ComponentTypes   []string
	Actions          []string
	PrebuiltTemplate string
}

func (s *Service) Init(ctx context.Context, genParams ConfigGenParams, params *InitParams) error {
	var c *generator.ConfigStructure

	// Check if prebuilt template is selected
	if params != nil && params.PrebuiltTemplate != "" {
		switch params.PrebuiltTemplate {
		case "aws-eks":
			c = BuildEKSSimpleConfigStructure(genParams.Path)
		case "aws-ecs":
			// TODO: Implement BuildECSSimpleConfigStructure
			return errors.New("aws-ecs template not yet implemented")
		default:
			return errors.Errorf("unknown prebuilt template: %s", params.PrebuiltTemplate)
		}
	} else if params != nil && (params.AppName != "" || params.StackType != "" || params.RunnerType != "" || len(params.ComponentTypes) > 0) {
		c = BuildConfigStructureFromParams(genParams.Path, params)
	} else {
		c = generator.DefaultAppConfigConfigStructure(genParams.Path)
	}

	gen := generator.NewConfigGen(
		genParams.EnableDefaults,
		genParams.EnableComments,
		false,
		genParams.Overwrite,
		genParams.SkipNonRequired,
	)

	err := gen.Gen(genParams.Path, c)
	if err != nil {
		return errors.Wrap(err, "failed to generate app config")
	}

	return nil
}

type SampleActionsParams struct {
	EnableSampleActions bool
	Actions             []string
}

func (s *Service) InitSampleActions(ctx context.Context, genParams ConfigGenParams, params SampleActionsParams) error {
	c := generator.NewConfigStructure(genParams.Path)
	if len(params.Actions) > 0 {
		// action name is not being used now
		for range params.Actions {
			c.AddActions(
				generator.ConfigFileDefinition{
					Name: "sample_action.toml",
					Schemas: []generator.ConfigFileSchema{
						{Instance: &config.ActionConfig{}},
					},
				},
			)
		}
	}

	gen := generator.NewConfigGen(
		genParams.EnableDefaults,
		genParams.EnableComments,
		false,
		genParams.Overwrite,
		genParams.SkipNonRequired,
	)

	err := gen.Gen(genParams.Path, &c)
	if err != nil {
		return errors.Wrap(err, "failed to generate action configs")
	}

	return nil
}

type SampleComponentParams struct {
	EnableSampleComponents bool
	ComponentTypes         []string
}

func (s *Service) InitSampleComponents(ctx context.Context, genParams ConfigGenParams, params SampleComponentParams) error {
	c := generator.NewConfigStructure(genParams.Path)
	if len(params.ComponentTypes) > 0 {
		for _, componentType := range params.ComponentTypes {
			switch componentType {
			case "terraform-module":
				c.AddComponent(
					generator.ConfigFileDefinition{
						Name: "example_terraform_module.toml",
						Schemas: []generator.ConfigFileSchema{
							{Instance: &config.Component{Type: config.TerraformModuleComponentType}},
							{Instance: &config.TerraformModuleComponentConfig{}},
						},
					},
				)
			case "helm-chart":
				c.AddComponent(
					generator.ConfigFileDefinition{
						Name: "example_helm_chart.toml",
						Schemas: []generator.ConfigFileSchema{
							{Instance: &config.Component{Type: config.HelmChartComponentType}},
							{Instance: &config.HelmChartComponentConfig{}},
						},
					},
				)
			case "kubernetes-manifest":
				c.AddComponent(
					generator.ConfigFileDefinition{
						Name: "example_kubernetes_manifest.toml",
						Schemas: []generator.ConfigFileSchema{
							{Instance: &config.Component{Type: config.KubernetesManifestComponentType}},
							{Instance: &config.KubernetesManifestComponentConfig{}},
						},
					},
				)
			}
		}
	}

	gen := generator.NewConfigGen(
		genParams.EnableDefaults,
		genParams.EnableComments,
		false,
		genParams.Overwrite,
		genParams.SkipNonRequired,
	)

	err := gen.Gen(genParams.Path, &c)
	if err != nil {
		return errors.Wrap(err, "failed to generate component configs")
	}

	return nil
}

func (s *Service) InitConfigFile(ctx context.Context, path string, configType string, genParams ConfigGenParams) error {
	gen := NewGen(genParams)
	// Create a custom ConfigStructure with only the specified config file
	configStructure := generator.DefaultAppConfigConfigStructure(path)

	// Filter to only include the requested config type
	filteredConfigs := []generator.ConfigFileDefinition{}
	for _, config := range configStructure.Configs {
		if config.Name == configType {
			filteredConfigs = append(filteredConfigs, config)
			break
		}
	}

	if len(filteredConfigs) == 0 {
		return errors.Errorf("unknown config type: %s", configType)
	}

	// Create a new structure with only the requested config
	customStructure := &generator.ConfigStructure{
		Name:              path,
		Configs:           filteredConfigs,
		ConfigDirectories: []generator.ConfigDirectoryDefinition{},
	}

	err := gen.Gen(genParams.Path, customStructure)
	if err != nil {
		return errors.Wrapf(err, "failed to generate %s config", configType)
	}

	return nil
}

type SandboxParams struct {
	TerraformVersion string
	PublicRepo       string
	PublicRepoDir    string
	PublicRepoBranch string
	ConnectedRepo    string
	ConnectedRepoDir string
	ConnectedBranch  string
	DriftSchedule    string
	EnvVars          map[string]string
	Vars             map[string]string
	VarFiles         []string
}

func (s *Service) InitSandboxConfig(ctx context.Context, genParams ConfigGenParams, params SandboxParams) error {
	gen := NewGen(genParams)

	// Build the sandbox config instance
	sandboxConfig := &config.AppSandboxConfig{
		TerraformVersion: params.TerraformVersion,
		EnvVarMap:        params.EnvVars,
		VarsMap:          params.Vars,
	}

	// Set public repo if provided
	if params.PublicRepo != "" {
		sandboxConfig.PublicRepo = &config.PublicRepoConfig{
			Repo:      params.PublicRepo,
			Directory: params.PublicRepoDir,
			Branch:    params.PublicRepoBranch,
		}
	}

	// Set connected repo if provided
	if params.ConnectedRepo != "" {
		sandboxConfig.ConnectedRepo = &config.ConnectedRepoConfig{
			Repo:      params.ConnectedRepo,
			Directory: params.ConnectedRepoDir,
			Branch:    params.ConnectedBranch,
		}
	}

	// Set drift schedule if provided
	if params.DriftSchedule != "" {
		sandboxConfig.DriftSchedule = &params.DriftSchedule
	}

	// Set var files if provided
	if len(params.VarFiles) > 0 {
		sandboxConfig.VariablesFiles = make([]config.TerraformVariablesFile, len(params.VarFiles))
		for i, vf := range params.VarFiles {
			sandboxConfig.VariablesFiles[i] = config.TerraformVariablesFile{
				Contents: vf,
			}
		}
	}

	customStructure := &generator.ConfigStructure{
		Name: genParams.Path,
		Configs: []generator.ConfigFileDefinition{
			{
				Name: "sandbox.toml",
				Schemas: []generator.ConfigFileSchema{
					{
						Instance: sandboxConfig,
					},
				},
			},
		},
		ConfigDirectories: []generator.ConfigDirectoryDefinition{},
	}

	err := gen.Gen(genParams.Path, customStructure)
	if err != nil {
		return errors.Wrap(err, "failed to generate sandbox.toml config")
	}

	return nil
}

type StackParams struct {
	Type                    string
	Name                    string
	Description             string
	VPCNestedTemplateURL    string
	RunnerNestedTemplateURL string
}

func (s *Service) InitStackConfig(ctx context.Context, genParams ConfigGenParams, params StackParams) error {
	gen := NewGen(genParams)

	// Build the stack config instance
	stackConfig := &config.StackConfig{
		Type:                    params.Type,
		Name:                    params.Name,
		Description:             params.Description,
		VPCNestedTemplateURL:    params.VPCNestedTemplateURL,
		RunnerNestedTemplateURL: params.RunnerNestedTemplateURL,
	}

	customStructure := &generator.ConfigStructure{
		Name: genParams.Path,
		Configs: []generator.ConfigFileDefinition{
			{
				Name: "stack.toml",
				Schemas: []generator.ConfigFileSchema{
					{
						Instance: stackConfig,
					},
				},
			},
		},
		ConfigDirectories: []generator.ConfigDirectoryDefinition{},
	}

	err := gen.Gen(genParams.Path, customStructure)
	if err != nil {
		return errors.Wrap(err, "failed to generate stack.toml config")
	}

	return nil
}

type RunnerParams struct {
	RunnerType    string
	EnvVars       map[string]string
	HelmDriver    string
	InitScriptURL string
}

func (s *Service) InitRunnerConfig(ctx context.Context, genParams ConfigGenParams, params RunnerParams) error {
	gen := NewGen(genParams)

	// Build the runner config instance
	runnerConfig := &config.AppRunnerConfig{
		RunnerType:    params.RunnerType,
		EnvVarMap:     params.EnvVars,
		HelmDriver:    params.HelmDriver,
		InitScriptURL: params.InitScriptURL,
	}

	customStructure := &generator.ConfigStructure{
		Name: genParams.Path,
		Configs: []generator.ConfigFileDefinition{
			{
				Name: "runner.toml",
				Schemas: []generator.ConfigFileSchema{
					{
						Instance: runnerConfig,
					},
				},
			},
		},
		ConfigDirectories: []generator.ConfigDirectoryDefinition{},
	}

	err := gen.Gen(genParams.Path, customStructure)
	if err != nil {
		return errors.Wrap(err, "failed to generate runner.toml config")
	}

	return nil
}

// TerraformModuleComponentParams holds parameters for Terraform module component configuration
type TerraformModuleComponentParams struct {
	Name             string
	VarName          string
	Dependencies     []string
	TerraformVersion string
	EnvVars          map[string]string
	Vars             map[string]string
	VarFiles         []string
	PublicRepo       string
	PublicRepoDir    string
	PublicRepoBranch string
	ConnectedRepo    string
	ConnectedRepoDir string
	ConnectedBranch  string
	DriftSchedule    string
}

func (s *Service) InitTerraformModuleComponentConfig(ctx context.Context, genParams ConfigGenParams, params TerraformModuleComponentParams) error {
	gen := NewGen(genParams)

	// Build the terraform module component config
	tfModuleConfig := &config.TerraformModuleComponentConfig{
		TerraformVersion: params.TerraformVersion,
		EnvVarMap:        params.EnvVars,
		VarsMap:          params.Vars,
	}

	// Set public repo if provided
	if params.PublicRepo != "" {
		tfModuleConfig.PublicRepo = &config.PublicRepoConfig{
			Repo:      params.PublicRepo,
			Directory: params.PublicRepoDir,
			Branch:    params.PublicRepoBranch,
		}
	}

	// Set connected repo if provided
	if params.ConnectedRepo != "" {
		tfModuleConfig.ConnectedRepo = &config.ConnectedRepoConfig{
			Repo:      params.ConnectedRepo,
			Directory: params.ConnectedRepoDir,
			Branch:    params.ConnectedBranch,
		}
	}

	// Set drift schedule if provided
	if params.DriftSchedule != "" {
		tfModuleConfig.DriftSchedule = &params.DriftSchedule
	}

	// Set var files if provided
	if len(params.VarFiles) > 0 {
		tfModuleConfig.VariablesFiles = make([]config.TerraformVariablesFile, len(params.VarFiles))
		for i, vf := range params.VarFiles {
			tfModuleConfig.VariablesFiles[i] = config.TerraformVariablesFile{
				Contents: vf,
			}
		}
	}

	// Build the component wrapper
	component := &config.Component{
		Type:            config.TerraformModuleComponentType,
		Name:            params.Name,
		VarName:         params.VarName,
		Dependencies:    params.Dependencies,
		TerraformModule: tfModuleConfig,
	}

	customStructure := &generator.ConfigStructure{
		Name: genParams.Path,
		ConfigDirectories: []generator.ConfigDirectoryDefinition{
			{
				Name: "components",
				Configs: []generator.ConfigFileDefinition{
					{
						Name: params.Name + ".toml",
						Schemas: []generator.ConfigFileSchema{
							{
								Instance: component,
							},
						},
					},
				},
			},
		},
	}

	err := gen.Gen(genParams.Path, customStructure)
	if err != nil {
		return errors.Wrap(err, "failed to generate terraform module component config")
	}

	return nil
}

// HelmChartComponentParams holds parameters for Helm chart component configuration
type HelmChartComponentParams struct {
	Name             string
	VarName          string
	Dependencies     []string
	ChartName        string
	Values           map[string]string
	ValuesFiles      []string
	PublicRepo       string
	PublicRepoDir    string
	PublicRepoBranch string
	ConnectedRepo    string
	ConnectedRepoDir string
	ConnectedBranch  string
	HelmRepoURL      string
	HelmChart        string
	HelmVersion      string
	Namespace        string
	StorageDriver    string
	TakeOwnership    bool
	DriftSchedule    string
}

func (s *Service) InitHelmChartComponentConfig(ctx context.Context, genParams ConfigGenParams, params HelmChartComponentParams) error {
	gen := NewGen(genParams)

	// Build the helm chart component config
	helmConfig := &config.HelmChartComponentConfig{
		ChartName:     params.ChartName,
		ValuesMap:     params.Values,
		Namespace:     params.Namespace,
		StorageDriver: params.StorageDriver,
		TakeOwnership: params.TakeOwnership,
	}

	// Set public repo if provided
	if params.PublicRepo != "" {
		helmConfig.PublicRepo = &config.PublicRepoConfig{
			Repo:      params.PublicRepo,
			Directory: params.PublicRepoDir,
			Branch:    params.PublicRepoBranch,
		}
	}

	// Set connected repo if provided
	if params.ConnectedRepo != "" {
		helmConfig.ConnectedRepo = &config.ConnectedRepoConfig{
			Repo:      params.ConnectedRepo,
			Directory: params.ConnectedRepoDir,
			Branch:    params.ConnectedBranch,
		}
	}

	// Set helm repo if provided
	if params.HelmRepoURL != "" {
		helmConfig.HelmRepo = &config.HelmRepoConfig{
			RepoURL: params.HelmRepoURL,
			Chart:   params.HelmChart,
			Version: params.HelmVersion,
		}
	}

	// Set drift schedule if provided
	if params.DriftSchedule != "" {
		helmConfig.DriftSchedule = &params.DriftSchedule
	}

	// Set values files if provided
	if len(params.ValuesFiles) > 0 {
		helmConfig.ValuesFiles = make([]config.HelmValuesFile, len(params.ValuesFiles))
		for i, vf := range params.ValuesFiles {
			helmConfig.ValuesFiles[i] = config.HelmValuesFile{
				Contents: vf,
			}
		}
	}

	// Build the component wrapper
	component := &config.Component{
		Type:         config.HelmChartComponentType,
		Name:         params.Name,
		VarName:      params.VarName,
		Dependencies: params.Dependencies,
		HelmChart:    helmConfig,
	}

	customStructure := &generator.ConfigStructure{
		Name: genParams.Path,
		ConfigDirectories: []generator.ConfigDirectoryDefinition{
			{
				Name: "components",
				Configs: []generator.ConfigFileDefinition{
					{
						Name: params.Name + ".toml",
						Schemas: []generator.ConfigFileSchema{
							{
								Instance: component,
							},
						},
					},
				},
			},
		},
	}

	err := gen.Gen(genParams.Path, customStructure)
	if err != nil {
		return errors.Wrap(err, "failed to generate helm chart component config")
	}

	return nil
}

// KubernetesManifestComponentParams holds parameters for Kubernetes manifest component configuration
type KubernetesManifestComponentParams struct {
	Name          string
	VarName       string
	Dependencies  []string
	Manifest      string
	Namespace     string
	DriftSchedule string
}

func (s *Service) InitKubernetesManifestComponentConfig(ctx context.Context, genParams ConfigGenParams, params KubernetesManifestComponentParams) error {
	gen := NewGen(genParams)

	// Build the kubernetes manifest component config
	k8sManifestConfig := &config.KubernetesManifestComponentConfig{
		Manifest:  params.Manifest,
		Namespace: params.Namespace,
	}

	// Set drift schedule if provided
	if params.DriftSchedule != "" {
		k8sManifestConfig.DriftSchedule = &params.DriftSchedule
	}

	// Build the component wrapper
	component := &config.Component{
		Type:               config.KubernetesManifestComponentType,
		Name:               params.Name,
		VarName:            params.VarName,
		Dependencies:       params.Dependencies,
		KubernetesManifest: k8sManifestConfig,
	}

	customStructure := &generator.ConfigStructure{
		Name: genParams.Path,
		ConfigDirectories: []generator.ConfigDirectoryDefinition{
			{
				Name: "components",
				Configs: []generator.ConfigFileDefinition{
					{
						Name: params.Name + ".toml",
						Schemas: []generator.ConfigFileSchema{
							{
								Instance: component,
							},
						},
					},
				},
			},
		},
	}

	err := gen.Gen(genParams.Path, customStructure)
	if err != nil {
		return errors.Wrap(err, "failed to generate kubernetes manifest component config")
	}

	return nil
}

// build config structure from raw params
func BuildConfigStructureFromParams(path string, params *InitParams) *generator.ConfigStructure {
	structure := &generator.ConfigStructure{
		Name:              path,
		Configs:           []generator.ConfigFileDefinition{},
		ConfigDirectories: []generator.ConfigDirectoryDefinition{},
	}

	// Add inputs config
	structure.UpdateInputs(&config.AppInputConfig{})

	// Add sandbox config
	structure.UpdateSandbox(&config.AppSandboxConfig{})

	// Add stack config
	if params.StackType != "" || params.AppName != "" {
		stackConfig := &config.StackConfig{}
		if params.StackType != "" {
			stackConfig.Type = params.StackType
		}
		if params.AppName != "" {
			stackConfig.Name = params.AppName
		}
		structure.UpdateStack(stackConfig)
	} else {
		structure.UpdateStack(&config.StackConfig{})
	}

	// Add runner config
	if params.RunnerType != "" {
		runnerConfig := &config.AppRunnerConfig{
			RunnerType: params.RunnerType,
		}
		structure.UpdateRunner(runnerConfig)
	} else {
		structure.UpdateRunner(&config.AppRunnerConfig{})
	}

	// Add secrets config
	structure.UpdateSecrets(&config.SecretsConfig{})

	// Add break glass config
	structure.UpdateBreakGlass(&config.BreakGlass{})

	// Add policies config
	structure.UpdatePolicies(&config.PoliciesConfig{})

	// Add component configs
	if len(params.ComponentTypes) > 0 {
		for _, componentType := range params.ComponentTypes {
			switch componentType {
			case "terraform-module":
				structure.AddComponent(
					generator.ConfigFileDefinition{
						Name: "example_terraform_module.toml",
						Schemas: []generator.ConfigFileSchema{
							{Instance: &config.Component{Type: config.TerraformModuleComponentType}},
							{Instance: &config.TerraformModuleComponentConfig{}},
						},
					},
				)
			case "helm-chart":
				structure.AddComponent(
					generator.ConfigFileDefinition{
						Name: "example_helm_chart.toml",
						Schemas: []generator.ConfigFileSchema{
							{Instance: &config.Component{Type: config.HelmChartComponentType}},
							{Instance: &config.HelmChartComponentConfig{}},
						},
					},
				)
			case "kubernetes-manifest":
				structure.AddComponent(
					generator.ConfigFileDefinition{
						Name: "example_kubernetes_manifest.toml",
						Schemas: []generator.ConfigFileSchema{
							{Instance: &config.Component{Type: config.KubernetesManifestComponentType}},
							{Instance: &config.KubernetesManifestComponentConfig{}},
						},
					},
				)
			}
		}
	}

	// Add action configs
	if len(params.Actions) > 0 {
		for _, actionName := range params.Actions {
			structure.AddActions(
				generator.ConfigFileDefinition{
					Name: actionName + ".toml",
					Schemas: []generator.ConfigFileSchema{
						{Instance: &config.ActionConfig{}},
					},
				},
			)
		}
	}

	return structure
}

// BuildEKSSimpleConfigStructure creates a config structure for aws-eks style app
func BuildEKSSimpleConfigStructure(path string) *generator.ConfigStructure {
	structure := &generator.ConfigStructure{
		Name:              path,
		Configs:           []generator.ConfigFileDefinition{},
		ConfigDirectories: []generator.ConfigDirectoryDefinition{},
	}

	sandboxConfig := &config.AppSandboxConfig{
		TerraformVersion: "1.11.3",
		PublicRepo: &config.PublicRepoConfig{
			Repo:      "nuonco/aws-eks-sandbox",
			Directory: "",
			Branch:    "main",
		},
		VarsMap: map[string]string{
			"cluster_identifier": "n-{{.nuon.install.id}}",
			"enable_dns":         "true",
			"public_domain":      "{{.nuon.install.id}}.{{.nuon.inputs.domain}}",
			"internal_domain":    "internal.{{.nuon.install.id}}.{{.nuon.inputs.domain}}",
		},
		VariablesFiles: []config.TerraformVariablesFile{
			{
				Contents: "./sandbox.tfvars",
			},
		},
	}
	structure.UpdateSandbox(sandboxConfig)

	stackConfig := &config.StackConfig{
		Type:                    "aws-cloudformation",
		Name:                    "nuon-demos-eks-simple-{{.nuon.install.id}}",
		Description:             "QuickLink to install runner for BYOC Nuon: Install {{.nuon.install.id}}",
		VPCNestedTemplateURL:    "https://nuon-artifacts.s3.us-west-2.amazonaws.com/aws-cloudformation-templates/v0.1.8/vpc/eks/default/stack.yaml",
		RunnerNestedTemplateURL: "https://nuon-artifacts.s3.us-west-2.amazonaws.com/aws-cloudformation-templates/v0.1.8/runner/asg/stack.yaml",
	}
	structure.UpdateStack(stackConfig)

	runnerConfig := &config.AppRunnerConfig{
		RunnerType:    "aws",
		HelmDriver:    "configmap",
		InitScriptURL: "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init-mng.sh",
		EnvVarMap:     map[string]string{},
	}
	structure.UpdateRunner(runnerConfig)

	inputsConfig := &config.AppInputConfig{
		Groups: []config.AppInputGroup{
			{
				Name:        "dns",
				Description: "DNS Configrations",
				DisplayName: "Configurations for the root domain for Route53",
			},
		},
		Inputs: []config.AppInput{
			{
				Name:        "domain",
				Description: "domain for the whoami endpoint e.g., nuon.run",
				Default:     "nuon.run",
				DisplayName: "Domain",
				Group:       "dns",
			},
			{
				Name:        "sub_domain",
				Description: "The sub domain for the Whoami service",
				Default:     "whoami",
				DisplayName: "Sub Domain",
				Group:       "dns",
			},
		},
	}
	structure.UpdateInputs(inputsConfig)

	policiesConfig := &config.PoliciesConfig{
		Policies: []config.AppPolicy{
			{
				Type:     config.AppPolicyTypeKubernetesClusterKyverno,
				Contents: "./disallow-ingress-nginx-custom-snippets.yml",
			},
		},
	}
	structure.UpdatePolicies(policiesConfig)

	provisionRole := &config.AppAWSIAMRole{
		Type:        string(config.PermissionsRoleTypeProvision),
		Name:        "{{.nuon.install.id}}-provision",
		Description: "provision the sandbox and components; trigger actions.",
		DisplayName: "provision role",
		Policies: []config.AppAWSIAMPolicy{
			{
				ManagedPolicyName: "AdministratorAccess",
			},
		},
		PermissionsBoundary: "./provision_boundary.json",
	}

	structure.AddPermission(generator.ConfigFileDefinition{
		Name: "provision.toml",
		Schemas: []generator.ConfigFileSchema{
			{Instance: provisionRole},
		},
	})

	maintenanceRole := &config.AppAWSIAMRole{
		Type:        string(config.PermissionsRoleTypeMaintenance),
		Name:        "{{.nuon.install.id}}-maintenance",
		Description: "operate and remediate the app's components and use actions.",
		DisplayName: "maintenance role",
		Policies: []config.AppAWSIAMPolicy{
			{
				ManagedPolicyName: "AdministratorAccess",
			},
			{
				Name: "limited-rds-secrets-manager-policy",
				Contents: `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "VisualEditor0",
      "Effect": "Allow",
      "Action": [
        "secretsmanager:CreateSecret",
        "secretsmanager:PutSecretValue",
        "secretsmanager:TagResource",
        "secretsmanager:UpdateSecret",
        "secretsmanager:GetSecretValue"
      ],
      "Resource": "arn:aws:secretsmanager:{{ .nuon.cloud_account.aws.region }}:*:secret:rds!*",
      "Condition": {
        "StringEquals": {
          "aws:RequestedRegion": "{{ .nuon.cloud_account.aws.region }}",
          "aws:ResourceTag/nuon_id": "{{ .nuon.install.id }}"
        }
      }
    }
  ]
}`,
			},
			{
				Name: "secrets-list-policy",
				Contents: `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "VisualEditor0",
      "Effect": "Allow",
      "Action": "secretsmanager:ListSecrets",
      "Resource": "*"
    }
  ]
}`,
			},
			{
				Name: "s3-bucket-policy",
				Contents: `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "VisualEditor0",
      "Effect": "Allow",
      "Action": "s3:PutBucketPolicy",
      "Resource": "*"
    }
  ]
}`,
			},
		},
		PermissionsBoundary: "./maintenance_boundary.json",
	}

	structure.AddPermission(generator.ConfigFileDefinition{
		Name: "maintenance.toml",
		Schemas: []generator.ConfigFileSchema{
			{Instance: maintenanceRole},
		},
	})

	deprovisionRole := &config.AppAWSIAMRole{
		Type:        string(config.PermissionsRoleTypeDeprovision),
		Name:        "{{.nuon.install.id}}-deprovision",
		Description: "deprovision sandbox and components. you must still delete the cf stack to delete the runner, ec2 vm, and vpc.",
		DisplayName: "deprovision role",
		Policies: []config.AppAWSIAMPolicy{
			{
				ManagedPolicyName: "AdministratorAccess",
			},
		},
		PermissionsBoundary: "./deprovision_boundary.json",
	}

	structure.AddPermission(generator.ConfigFileDefinition{
		Name: "deprovision.toml",
		Schemas: []generator.ConfigFileSchema{
			{Instance: deprovisionRole},
		},
	})

	return structure
}
