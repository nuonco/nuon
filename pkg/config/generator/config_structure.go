package generator

import (
	"fmt"
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/nuonco/nuon/pkg/config"
)

type ConfigFileSchema struct {
	SkipNonRequired bool
	Instance        any
}

func (c *ConfigFileSchema) Schema() *jsonschema.Schema {
	if c.Instance == nil {
		return nil
	}

	var schema *jsonschema.Schema
	r := NewDefaultReflector()

	refValue := reflect.ValueOf(c.Instance)

	if refValue.Kind() == reflect.Pointer {
		refValue = refValue.Elem()
	}

	if refValue.Kind() == reflect.Array || refValue.Kind() == reflect.Slice {
		if refValue.Len() == 0 {
			sliceType := refValue.Type().Elem()
			sliceElm := reflect.New(sliceType)
			schema = r.Reflect(sliceElm.Interface())
		} else {
			value := refValue.Index(0)
			schema = r.Reflect(value.Interface())
		}
	} else {
		schema = r.Reflect(refValue.Interface())
	}
	return schema
}

type ConfigFileDefinition struct {
	Header      string
	Name        string
	Schemas     []ConfigFileSchema
	TomlEncoded string
}

type ConfigDirectoryDefinition struct {
	Name string
	// configFiles
	Configs []ConfigFileDefinition
}

type ConfigStructure struct {
	Name string
	// config files
	Configs []ConfigFileDefinition
	// directory containing config files
	ConfigDirectories []ConfigDirectoryDefinition
	// non-config files written verbatim (e.g. README.md)
	RawFiles []RawFileDefinition
}

// RawFileDefinition is a plain file written to the config root as-is, without
// TOML encoding or a schema directive.
type RawFileDefinition struct {
	Name     string
	Contents string
}

func NewConfigStructure(name string) ConfigStructure {
	return ConfigStructure{
		Name:              name,
		Configs:           []ConfigFileDefinition{},
		ConfigDirectories: []ConfigDirectoryDefinition{},
	}
}

func (c *ConfigStructure) AddDirectoryFile(dirName string, cfd ConfigFileDefinition) error {
	// Find the directory
	for i := range c.ConfigDirectories {
		if c.ConfigDirectories[i].Name == dirName {
			// Check if file with same name already exists
			for _, existingConfig := range c.ConfigDirectories[i].Configs {
				if existingConfig.Name == cfd.Name {
					return fmt.Errorf("config file '%s' already exists in directory '%s'", cfd.Name, dirName)
				}
			}
			c.ConfigDirectories[i].Configs = append(c.ConfigDirectories[i].Configs, cfd)
			return nil
		}
	}

	// If directory doesn't exist, create it
	c.ConfigDirectories = append(c.ConfigDirectories, ConfigDirectoryDefinition{
		Name:    dirName,
		Configs: []ConfigFileDefinition{cfd},
	})
	return nil
}

func (c *ConfigStructure) AddFile(cfd ConfigFileDefinition, overwrite bool) error {
	for i := range c.Configs {
		if c.Configs[i].Name == cfd.Name {
			if !overwrite {
				return fmt.Errorf("config file '%s' already exists", cfd.Name)
			}
			c.Configs[i] = cfd
			return nil
		}
	}
	c.Configs = append(c.Configs, cfd)

	return nil
}

// updates the config in the structure
func (c *ConfigStructure) UpdateConfig(cfd ConfigFileDefinition) error {
	return c.AddFile(cfd, true)
}

func (c *ConfigStructure) AddComponent(cfd ConfigFileDefinition) error {
	for _, schema := range cfd.Schemas {
		var comp *config.Component
		switch v := schema.Instance.(type) {
		case *config.Component:
			comp = v
		case config.Component:
			comp = &v
		default:
			continue
		}

		if comp != nil {
			// map component type to schema header based on config/schema/types.go
			switch comp.Type {
			case config.TerraformModuleComponentType:
				cfd.Header = "terraform"
			case config.HelmChartComponentType:
				cfd.Header = "helm"
			case config.DockerBuildComponentType:
				cfd.Header = "docker-build"
			case config.ContainerImageComponentType, config.ExternalImageComponentType:
				cfd.Header = "container-image"
			case config.KubernetesManifestComponentType:
				cfd.Header = "kubernetes-manifest"
			case config.PulumiComponentType:
				cfd.Header = "pulumi"
			case config.JobComponentType:
				cfd.Header = "job"
			}
			break
		}
	}
	return c.AddDirectoryFile("components", cfd)
}

func (c *ConfigStructure) AddActions(cfd ConfigFileDefinition) error {
	return c.AddDirectoryFile("actions", cfd)
}

func (c *ConfigStructure) AddPermission(cfd ConfigFileDefinition) error {
	return c.AddDirectoryFile("permissions", cfd)
}

// UpdateInputs updates the inputs.toml configuration
func (c *ConfigStructure) UpdateInputs(cfg *config.AppInputConfig) error {
	return c.UpdateConfig(ConfigFileDefinition{
		Header: "inputs",
		Name:   "inputs.toml",
		Schemas: []ConfigFileSchema{
			{
				SkipNonRequired: false,
				Instance:        cfg,
			},
		},
	})
}

// UpdateSandbox updates the sandbox.toml configuration
func (c *ConfigStructure) UpdateSandbox(cfg *config.AppSandboxConfig) error {
	return c.UpdateConfig(ConfigFileDefinition{
		Header: "sandbox",
		Name:   "sandbox.toml",
		Schemas: []ConfigFileSchema{
			{
				Instance: cfg,
			},
		},
	})
}

// UpdateStack updates the stack.toml configuration
func (c *ConfigStructure) UpdateStack(cfg *config.StackConfig) error {
	return c.UpdateConfig(ConfigFileDefinition{
		Header: "stack",
		Name:   "stack.toml",
		Schemas: []ConfigFileSchema{
			{
				Instance: cfg,
			},
		},
	})
}

// UpdateRunner updates the runner.toml configuration
func (c *ConfigStructure) UpdateRunner(cfg *config.AppRunnerConfig) error {
	return c.UpdateConfig(ConfigFileDefinition{
		Header: "runner",
		Name:   "runner.toml",
		Schemas: []ConfigFileSchema{
			{
				Instance: cfg,
			},
		},
	})
}

// UpdatePolicies updates the policies.toml configuration
func (c *ConfigStructure) UpdatePolicies(cfg *config.PoliciesConfig) error {
	return c.UpdateConfig(ConfigFileDefinition{
		Header: "policies",
		Name:   "policies.toml",
		Schemas: []ConfigFileSchema{
			{
				SkipNonRequired: false,
				Instance:        cfg,
			},
		},
	})
}

// UpdateBreakGlass updates the break_glass.toml configuration
func (c *ConfigStructure) UpdateBreakGlass(cfg *config.BreakGlass) error {
	return c.UpdateConfig(ConfigFileDefinition{
		Header: "break-glass",
		Name:   "break_glass.toml",
		Schemas: []ConfigFileSchema{
			{
				Instance: cfg,
			},
		},
	})
}

// UpdateSecrets updates the secrets.toml configuration
func (c *ConfigStructure) UpdateSecrets(cfg *config.SecretsConfig) error {
	return c.UpdateConfig(ConfigFileDefinition{
		Header: "secrets",
		Name:   "secrets.toml",
		Schemas: []ConfigFileSchema{
			{
				SkipNonRequired: false,
				Instance:        cfg,
			},
		},
	})
}

// UpdateInstaller updates the installer.toml configuration
func (c *ConfigStructure) UpdateInstaller(cfg *config.InstallerConfig) error {
	return c.UpdateConfig(ConfigFileDefinition{
		Header: "installer",
		Name:   "installer.toml",
		Schemas: []ConfigFileSchema{
			{
				Instance: cfg,
			},
		},
	})
}

// UpdateMetadata updates the metadata.toml configuration
func (c *ConfigStructure) UpdateMetadata(cfg *config.MetadataConfig) error {
	return c.UpdateConfig(ConfigFileDefinition{
		Header: "metadata",
		Name:   "metadata.toml",
		Schemas: []ConfigFileSchema{
			{
				Instance: cfg,
			},
		},
	})
}

func DefaultAppConfigConfigStructure(name string) *ConfigStructure {
	kubernetesSync := true

	return &ConfigStructure{
		Name: name,
		// Root-level config files
		Configs: []ConfigFileDefinition{
			{
				Name: "metadata.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance: &config.MetadataConfig{
							Version:     "v1",
							DisplayName: "My App",
							Description: "A Nuon-deployed application.",
							Readme:      "./README.md",
						},
					},
				},
			},
			{
				Name: "sandbox.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance: &config.AppSandboxConfig{
							TerraformVersion: "1.11.3",
							PublicRepo: &config.PublicRepoConfig{
								Repo:      "nuonco/aws-eks-sandbox",
								Directory: ".",
								Branch:    "main",
							},
							VarsMap: map[string]string{
								"cluster_name":    "n-{{.nuon.install.id}}",
								"cluster_version": "1.34",
							},
						},
					},
				},
			},
			{
				Name: "runner.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance: &config.AppRunnerConfig{
							RunnerType:    "aws",
							HelmDriver:    "configmap",
							InitScriptURL: "https://raw.githubusercontent.com/nuonco/runner/refs/tags/aws-v0.1.0/scripts/aws/init-mng.sh",
							EnvVarMap: map[string]string{
								"HELM_MAX_HISTORY":   "10",
								"RUNNER_AUTH_METHOD": "iid",
							},
						},
					},
				},
			},
			{
				Name: "stack.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance: &config.StackConfig{
							Type:                    "aws-cloudformation",
							Name:                    "my-app-{{.nuon.install.id}}",
							Description:             "CloudFormation stack for My App: install {{.nuon.install.id}}",
							VPCNestedTemplateURL:    "https://nuon-artifacts.s3.us-west-2.amazonaws.com/aws-cloudformation-templates/v0.4.2/vpc/eks/default/stack.yaml",
							RunnerNestedTemplateURL: "https://nuon-artifacts.s3.us-west-2.amazonaws.com/aws-cloudformation-templates/v0.4.2/runner/asg/stack.yaml",
						},
					},
				},
			},
			{
				Name: "secrets.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance: &config.SecretsConfig{
							Secrets: []*config.AppSecret{
								{
									Name:                      "db_password",
									DisplayName:               "Database Password",
									Description:               "Password for the application database.",
									AutoGenerate:              true,
									KubernetesSync:            &kubernetesSync,
									KubernetesSecretNamespace: "my-app",
									KubernetesSecretName:      "my-app-secrets",
								},
							},
						},
					},
				},
			},
			{
				Name: "break_glass.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance: &config.BreakGlass{
							Roles: []*config.AppAWSIAMRole{
								{
									Name:        "{{.nuon.install.id}}-app-break-glass",
									DisplayName: "Break Glass Admin",
									Description: "grants admin access for emergencies",
									Policies: []config.AppAWSIAMPolicy{
										{ManagedPolicyName: "AdministratorAccess"},
									},
								},
							},
						},
					},
				},
			},
		},
		// Subdirectories with their config files
		ConfigDirectories: []ConfigDirectoryDefinition{
			{
				Name: "input_groups",
				Configs: []ConfigFileDefinition{
					{
						Header: "input-group",
						Name:   "example.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.AppInputGroup{
									Name:        "example",
									DisplayName: "Example",
									Description: "Example input group.",
								},
							},
						},
					},
				},
			},
			{
				Name: "inputs/example",
				Configs: []ConfigFileDefinition{
					{
						Header: "input",
						Name:   "example_input.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.AppInput{
									Name:             "instance_type",
									DisplayName:      "Node Instance Size",
									Description:      "EC2 instance type for worker nodes.",
									Group:            "example",
									Default:          "t3a.medium",
									Type:             "string",
									UserConfigurable: true,
								},
							},
						},
					},
				},
			},
			{
				Name: "components/example_helm_chart",
				Configs: []ConfigFileDefinition{
					{
						Name: "nuon.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.Component{
									Type: config.HelmChartComponentType,
									Name: "my_app_chart",
								},
							},
							{
								Instance: &config.HelmChartComponentConfig{
									ChartName: "my-app",
									Namespace: "my-app",
									ConnectedRepo: &config.ConnectedRepoConfig{
										Repo:      "your-org/your-repo",
										Directory: "components/chart",
										Branch:    "main",
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "components/example_terraform_module",
				Configs: []ConfigFileDefinition{
					{
						Name: "nuon.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.Component{
									Type: config.TerraformModuleComponentType,
									Name: "my_app_infra",
								},
							},
							{
								Instance: &config.TerraformModuleComponentConfig{
									TerraformVersion: "1.9.0",
									ConnectedRepo: &config.ConnectedRepoConfig{
										Repo:      "your-org/your-repo",
										Directory: "components/terraform",
										Branch:    "main",
									},
									VarsMap: map[string]string{
										"install_id": "{{.nuon.install.id}}",
										"region":     "{{.nuon.install_stack.outputs.region}}",
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "components/example_kubernetes_manifest",
				Configs: []ConfigFileDefinition{
					{
						Name: "nuon.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.Component{
									Type: config.KubernetesManifestComponentType,
									Name: "my_app_manifests",
								},
							},
							{
								Instance: &config.KubernetesManifestComponentConfig{
									Namespace: "my-app",
									ConnectedRepo: &config.ConnectedRepoConfig{
										Repo:      "your-org/your-repo",
										Directory: "components/manifests",
										Branch:    "main",
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "permissions",
				Configs: []ConfigFileDefinition{
					{
						Header: "permission",
						Name:   "provision.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.AppAWSIAMRole{
									Type:        "provision",
									Name:        "{{.nuon.install.id}}-provision",
									DisplayName: "provision role",
									Description: "Provisions the sandbox and components.",
									Policies: []config.AppAWSIAMPolicy{
										{ManagedPolicyName: "AdministratorAccess"},
									},
								},
							},
						},
					},
					{
						Header: "permission",
						Name:   "maintenance.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.AppAWSIAMRole{
									Type:        "maintenance",
									Name:        "{{.nuon.install.id}}-maintenance",
									DisplayName: "maintenance role",
									Description: "Deploys and maintains components.",
									Policies: []config.AppAWSIAMPolicy{
										{ManagedPolicyName: "AdministratorAccess"},
									},
								},
							},
						},
					},
					{
						Header: "permission",
						Name:   "deprovision.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.AppAWSIAMRole{
									Type:        "deprovision",
									Name:        "{{.nuon.install.id}}-deprovision",
									DisplayName: "deprovision role",
									Description: "Tears down the sandbox and components.",
									Policies: []config.AppAWSIAMPolicy{
										{ManagedPolicyName: "AdministratorAccess"},
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "actions/example_action",
				Configs: []ConfigFileDefinition{
					{
						Name: "nuon.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.ActionConfig{
									Name:    "healthcheck",
									Timeout: "30s",
									Triggers: []*config.ActionTriggerConfig{
										{Type: "cron", CronSchedule: "*/5 * * * *"},
										{Type: "manual"},
									},
									Steps: []*config.ActionStepConfig{
										{Name: "run-healthcheck", Command: "echo healthcheck"},
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "policies",
				Configs: []ConfigFileDefinition{
					{
						Header: "policy",
						Name:   "example_policy.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.AppPolicy{
									Type:   config.AppPolicyTypeTerraformModule,
									Engine: config.AppPolicyEngineOPA,
									Name:   "example-policy",
								},
							},
						},
					},
				},
			},
		},
		RawFiles: []RawFileDefinition{
			{
				Name:     "README.md",
				Contents: defaultReadme,
			},
		},
	}
}

const defaultReadme = `# My App

A Nuon-deployed application.

This directory contains the Nuon app configuration. Edit the files to describe
your app, then sync it with:

    nuon apps sync -c metadata.toml

See https://docs.nuon.co for more information.
`
