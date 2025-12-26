package generator

import (
	"os"
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/stretchr/testify/assert"
	"gotest.tools/assert"
)

func NewTestingConfigStructure(name string) *ConfigStructure {
	return &ConfigStructure{
		Name: name,
		// Root-level config files
		Configs: []ConfigFileDefinition{
			{
				Name: "inputs.toml",
				Schemas: []ConfigFileSchema{
					{
						SkipNonRequired: false,
						Instance: &config.AppInputConfig{
							Groups: []config.AppInputGroup{
								{
									Name:        "network",
									Description: "Configure the install's network settings.",
									DisplayName: "Network",
								},
							},
							Inputs: []config.AppInput{
								{
									Name:        "root_domain",
									Description: "Domain to host this install under",
									Default:     "example.nuon.run",
									Sensitive:   false,
									DisplayName: "Root Domain",
									Group:       "network",
									Internal:    true,
									Type:        "string",
								},
							},
						},
					},
				},
			},
			{
				Name: "installer.toml",
				Schemas: []ConfigFileSchema{
					{
						Instance: &config.InstallerConfig{
							Name:                "installer",
							Description:         "one click installer",
							DocumentationURL:    "docs-url",
							CommunityURL:        "community-url",
							HomepageURL:         "homepage-url",
							GithubURL:           "github-url",
							LogoURL:             "logo-url",
							DemoURL:             "https://nuon.co",
							PostInstallMarkdown: "Installation complete!",
							FooterMarkdown:      "Footer text",
							CopyrightMarkdown:   "Copyright 2024",
							OgImageURL:          "og-image-url",
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
							EnvVarMap: map[string]string{
								"cluster_name": "n-{{.nuon.install.id}}",
							},
							VarsMap: map[string]string{
								"cluster_name":         "{{ .nuon.install.id }}",
								"account_id":           "{{.nuon.install_stack.outputs.account_id}}",
								"enable_nuon_dns":      "true",
								"public_root_domain":   "{{ .nuon.inputs.inputs.root_domain }}",
								"internal_root_domain": "internal.{{ .nuon.inputs.inputs.root_domain }}",
							},
							VariablesFiles: []config.TerraformVariablesFile{
								{
									Contents: "./sandbox.tfvars",
								},
							},
						},
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
								Instance: &config.Component{
									Name: "whoami",
									Type: config.HelmChartComponentType,
								},
							},
							{
								Instance: &config.HelmChartComponentConfig{
									ChartName:     "whoami",
									Namespace:     "whoami",
									StorageDriver: "configmap",
									PublicRepo: &config.PublicRepoConfig{
										Repo:      "nuonco/demo",
										Directory: "eks-simple/src/components/whoami",
										Branch:    "main",
									},
									ValuesFiles: []config.HelmValuesFile{
										{
											Contents: "./whoami.yaml",
										},
									},
								},
							},
						},
					},
					{
						Name: "example_terraform_module.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.Component{
									Name: "s3_bucket",
									Type: config.TerraformModuleComponentType,
								},
							},
							{
								Instance: &config.TerraformModuleComponentConfig{
									TerraformVersion: "1.11.3",
									PublicRepo: &config.PublicRepoConfig{
										Repo:      "mrwong/s3_for_tests",
										Directory: ".",
										Branch:    "main",
									},
									VarsMap: map[string]string{
										"bucket_name_prefix": "{{.nuon.install.id}}",
										"region":             "{{ .nuon.install_stack.outputs.region }}",
									},
								},
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
								Instance: &config.AppAWSIAMRole{},
							},
							{
								Instance: []config.AppAWSIAMRole{{}},
							},
						},
					},
					{
						Name: "maintenance.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.AppAWSIAMRole{},
							},
							{
								Instance: []config.AppAWSIAMRole{},
							},
						},
					},
					{
						Name: "deprovision.toml",
						Schemas: []ConfigFileSchema{
							{
								Instance: &config.AppAWSIAMRole{},
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

// generates a  new config directory
func TestGenerate(t *testing.T) {
	// Test case 1: Basic generation
	generator := NewConfigGen(
		true,
		true,
		false,
		true,
		false,
	)
	err := generator.Gen("./test-config-init/", NewTestingConfigStructure("test-app-config"))
	assert.NoError(t, err, "generator existed with error")
}

// this is a ai generated tests, not to be trusted, only used for dev purposed
func TestGenerateWithInstanceValues(t *testing.T) {
	// This test verifies that instance values are being used in the generated TOML
	generator := NewConfigGen(true, true, false, true, false)

	// Generate the config files
	err := generator.Gen("./test-config-init/", NewTestingConfigStructure("test-app-config"))
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	// Test sandbox.toml - Check that instance values from seed configs are used
	sandboxContent, err := os.ReadFile("./test-config-init/sandbox.toml")
	if err != nil {
		t.Fatalf("Failed to read generated sandbox.toml: %v", err)
	}

	sandboxStr := string(sandboxContent)

	// Verify terraform_version from instance
	if !strings.Contains(sandboxStr, `terraform_version = "1.11.3"`) {
		t.Errorf("Generated sandbox.toml does not contain instance value for terraform_version")
	}

	// Verify public_repo values from instance (should be uncommented since they have values)
	if !strings.Contains(sandboxStr, `repo = "nuonco/aws-eks-sandbox"`) {
		t.Errorf("Generated sandbox.toml does not contain instance value for public_repo.repo")
	}

	if !strings.Contains(sandboxStr, `directory = "."`) {
		t.Errorf("Generated sandbox.toml does not contain instance value for public_repo.directory")
	}

	if !strings.Contains(sandboxStr, `branch = "main"`) {
		t.Errorf("Generated sandbox.toml does not contain instance value for public_repo.branch")
	}

	// Test installer.toml - Check that instance values are used
	installerContent, err := os.ReadFile("./test-config-init/installer.toml")
	if err != nil {
		t.Fatalf("Failed to read generated installer.toml: %v", err)
	}

	installerStr := string(installerContent)

	// Verify installer name from instance
	if !strings.Contains(installerStr, `name = "installer"`) {
		t.Errorf("Generated installer.toml does not contain instance value for name")
	}

	if !strings.Contains(installerStr, `demo_url = "https://nuon.co"`) {
		t.Errorf("Generated installer.toml does not contain instance value for demo_url")
	}

	t.Logf("Successfully verified instance values in generated TOML files")
}
