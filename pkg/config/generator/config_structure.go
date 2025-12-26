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
		Name: "inputs.toml",
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
		Name: "sandbox.toml",
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
		Name: "stack.toml",
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
		Name: "runner.toml",
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
		Name: "policies.toml",
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
		Name: "break_glass.toml",
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
		Name: "secrets.toml",
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
		Name: "installer.toml",
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
		Name: "metadata.toml",
		Schemas: []ConfigFileSchema{
			{
				Instance: cfg,
			},
		},
	})
}

func DefaultAppConfigConfigStructure(name string) *ConfigStructure {
	return &ConfigStructure{
		Name: name,
		// Root-level config files
		Configs: []ConfigFileDefinition{
			{
				Name: "inputs.toml",
				Schemas: []ConfigFileSchema{
					{
						SkipNonRequired: false,
						Instance:        &config.AppInputConfig{},
					},
				},
			},
			{
				Name: "installer.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance: &config.InstallerConfig{},
					},
				},
			},
			{
				Name: "sandbox.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance: &config.AppSandboxConfig{},
					},
				},
			},
			{
				Name: "runner.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance: &config.AppRunnerConfig{},
					},
				},
			},
			{
				Name: "stack.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance: &config.StackConfig{},
					},
				},
			},
			{
				Name: "secrets.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance:        &config.SecretsConfig{},
						SkipNonRequired: false,
					},
				},
			},
			{
				Name: "break_glass.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance: &config.BreakGlass{},
					},
				},
			},
			{
				Name: "policies.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance:        &config.PoliciesConfig{},
						SkipNonRequired: false,
					},
				},
			},
		},
		// Subdirectories with their config files
		ConfigDirectories: []ConfigDirectoryDefinition{
			{
				Name: "components",
				Configs: []ConfigFileDefinition{
					{
						Name: "example_helm_chart.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.Component{},
							},
							{
								Instance: &config.HelmChartComponentConfig{},
							},
						},
					},
					{
						Name: "example_terraform_module.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.Component{},
							},
							{
								Instance: &config.TerraformModuleComponentConfig{},
							},
						},
					},
					{
						Name: "example_kubernetes_manifest.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.Component{},
							},
							{
								Instance: &config.KubernetesManifestComponentConfig{},
							},
						},
					},
				},
			},
			{
				Name: "permissions",
				Configs: []ConfigFileDefinition{
					{
						Name: "provision.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.PermissionsConfig{},
							},
						},
					},
					{
						Name: "maintenance.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.PermissionsConfig{},
							},
						},
					},
					{
						Name: "deprovision.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.PermissionsConfig{},
							},
						},
					},
				},
			},
			{
				Name: "actions",
				Configs: []ConfigFileDefinition{
					{
						Name: "example_action.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.ActionConfig{},
							},
						},
					},
				},
			},
			{
				Name: "installs",
				Configs: []ConfigFileDefinition{
					{
						Name: "example_install.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.Install{},
							},
						},
					},
				},
			},
		},
	}
}
