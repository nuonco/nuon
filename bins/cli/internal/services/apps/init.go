package apps

import (
	"context"
	"fmt"

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
	var err error

	// Check if prebuilt template is selected
	if params != nil && params.PrebuiltTemplate != "" {
		switch params.PrebuiltTemplate {
		case "aws-eks":
			c, err = BuildEKSSimpleConfigStructure(ctx, genParams.Path)
			if err != nil {
				return errors.Wrap(err, "unable to create config structure")
			}
		case "aws-ecs-breakglass":
			c, err = BuildECSSimpleConfigStructure(ctx, genParams.Path)
			if err != nil {
				return errors.Wrap(err, "unable to create config structure")
			}
		case "aws-eks-auto":
			c, err = BuildECSSimpleConfigStructure(ctx, genParams.Path)
			if err != nil {
				return errors.Wrap(err, "unable to create config structure")
			}
		case "clickhouse-aws-eks":
			c, err = BuildECSSimpleConfigStructure(ctx, genParams.Path)
			if err != nil {
				return errors.Wrap(err, "unable to create config structure")
			}
		case "cockroachdb-aws-eks":
			c, err = BuildECSSimpleConfigStructure(ctx, genParams.Path)
			if err != nil {
				return errors.Wrap(err, "unable to create config structure")
			}
		case "grafana-aws-eks":
			c, err = BuildECSSimpleConfigStructure(ctx, genParams.Path)
			if err != nil {
				return errors.Wrap(err, "unable to create config structure")
			}
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

	err = gen.Gen(genParams.Path, c)
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

type ActionParams struct {
	Name             string
	Timeout          string
	TriggerType      string
	CronSchedule     string
	ComponentName    string
	StepName         string
	StepCommand      string
	InlineContents   string
	EnvVars          map[string]string
	PublicRepo       string
	PublicRepoDir    string
	PublicRepoBranch string
	ConnectedRepo    string
	ConnectedRepoDir string
	ConnectedBranch  string
	BreakGlassRole   string
	Dependencies     []string
}

func (s *Service) InitActionConfig(ctx context.Context, genParams ConfigGenParams, params ActionParams) error {
	gen := NewGen(genParams)

	trigger := &config.ActionTriggerConfig{
		Type: params.TriggerType,
	}

	if params.CronSchedule != "" {
		trigger.CronSchedule = params.CronSchedule
	}

	if params.ComponentName != "" {
		trigger.ComponentName = params.ComponentName
	}

	step := &config.ActionStepConfig{
		Name:      params.StepName,
		Command:   params.StepCommand,
		EnvVarMap: params.EnvVars,
	}

	if params.InlineContents != "" {
		step.InlineContents = params.InlineContents
	}

	if params.PublicRepo != "" {
		step.PublicRepo = &config.PublicRepoConfig{
			Repo:      params.PublicRepo,
			Directory: params.PublicRepoDir,
			Branch:    params.PublicRepoBranch,
		}
	}

	if params.ConnectedRepo != "" {
		step.ConnectedRepo = &config.ConnectedRepoConfig{
			Repo:      params.ConnectedRepo,
			Directory: params.ConnectedRepoDir,
			Branch:    params.ConnectedBranch,
		}
	}

	actionConfig := &config.ActionConfig{
		Name:         params.Name,
		Timeout:      params.Timeout,
		Triggers:     []*config.ActionTriggerConfig{trigger},
		Steps:        []*config.ActionStepConfig{step},
		Dependencies: params.Dependencies,
	}

	if params.BreakGlassRole != "" {
		actionConfig.BreakGlassRole = params.BreakGlassRole
	}

	customStructure := &generator.ConfigStructure{
		Name: genParams.Path,
		ConfigDirectories: []generator.ConfigDirectoryDefinition{
			{
				Name: "actions",
				Configs: []generator.ConfigFileDefinition{
					{
						Name: params.Name + ".toml",
						Schemas: []generator.ConfigFileSchema{
							{
								Instance: actionConfig,
							},
						},
					},
				},
			},
		},
	}

	err := gen.Gen(genParams.Path, customStructure)
	if err != nil {
		return errors.Wrap(err, "failed to generate action config")
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

func BuildNamedConfigStructure(ctx context.Context, configName, folderName string) (*generator.ConfigStructure, error) {
	configStructure, err := ReadAndConvertConfig(ctx, ConfigReaderParams{
		Folder: folderName,
	}, configName)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to build %s template app config structue", folderName))
	}
	return configStructure, nil
}

func BuildEKSSimpleConfigStructure(ctx context.Context, configName string) (*generator.ConfigStructure, error) {
	return BuildNamedConfigStructure(ctx, configName, "eks-simple")
}

func BuildECSSimpleConfigStructure(ctx context.Context, configName string) (*generator.ConfigStructure, error) {
	return BuildNamedConfigStructure(ctx, configName, "ecs-simple")
}

func BuildECSBreakglassConfigStructure(ctx context.Context, configName string) (*generator.ConfigStructure, error) {
	return BuildNamedConfigStructure(ctx, configName, "ecs-breakglass")
}

func BuildEKSAutoConfigStructure(ctx context.Context, configName string) (*generator.ConfigStructure, error) {
	return BuildNamedConfigStructure(ctx, configName, "eks-simple-auto")
}

func BuildGrafanaAWSEKSConfigStructure(ctx context.Context, configName string) (*generator.ConfigStructure, error) {
	return BuildNamedConfigStructure(ctx, configName, "grafana")
}

func BuildClickhouseAWSEKSConfigStructure(ctx context.Context, configName string) (*generator.ConfigStructure, error) {
	return BuildNamedConfigStructure(ctx, configName, "clickhouse")
}

func BuildCockroachdbAWSEKSConfigStructure(ctx context.Context, configName string) (*generator.ConfigStructure, error) {
	return BuildNamedConfigStructure(ctx, configName, "cockroachdb")
}
