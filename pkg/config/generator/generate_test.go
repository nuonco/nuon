package generator

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/config"
)

func NewTestingConfigStructure(name string) *ConfigStructure {
	cs := NewConfigStructure(name)

	cs.UpdateInputs(&config.AppInputConfig{
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
	})

	cs.UpdateInstaller(&config.InstallerConfig{
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
	})

	cs.UpdateSandbox(&config.AppSandboxConfig{
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
	})

	cs.UpdateRunner(&config.AppRunnerConfig{})
	cs.UpdateStack(&config.StackConfig{})
	cs.UpdateSecrets(&config.SecretsConfig{})
	cs.UpdateBreakGlass(&config.BreakGlass{})
	cs.UpdatePolicies(&config.PoliciesConfig{})

	cs.AddComponent(ConfigFileDefinition{
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
	})

	cs.AddComponent(ConfigFileDefinition{
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
	})

	cs.AddComponent(ConfigFileDefinition{
		Name: "example_kubernetes_manifest.toml",
		Schemas: []ConfigFileSchema{
			{
				Instance: &config.Component{},
			},
			{
				Instance: &config.KubernetesManifestComponentConfig{},
			},
		},
	})

	cs.AddPermission(ConfigFileDefinition{
		Name: "provision.toml",
		Schemas: []ConfigFileSchema{
			{
				Instance: &config.AppAWSIAMRole{},
			},
			{
				Instance: []config.AppAWSIAMRole{{}},
			},
		},
	})

	cs.AddPermission(ConfigFileDefinition{
		Name: "maintenance.toml",
		Schemas: []ConfigFileSchema{
			{
				Instance: &config.AppAWSIAMRole{},
			},
			{
				Instance: []config.AppAWSIAMRole{},
			},
		},
	})

	cs.AddPermission(ConfigFileDefinition{
		Name: "deprovision.toml",
		Schemas: []ConfigFileSchema{
			{
				Instance: &config.AppAWSIAMRole{},
			},
		},
	})

	cs.AddActions(ConfigFileDefinition{
		Name: "example_action.toml",
		Schemas: []ConfigFileSchema{
			{
				Instance: &config.ActionConfig{},
			},
		},
	})

	cs.AddDirectoryFile("installs", ConfigFileDefinition{
		Name: "example_install.toml",
		Schemas: []ConfigFileSchema{
			{
				Instance: &config.Install{},
			},
		},
	})

	return &cs
}

func TestGenerate(t *testing.T) {
	defer func() {
		err := os.RemoveAll("./test-app-init")
		if err != nil {
			t.Errorf("Failed to clean generated config %v", err)
		}
	}()

	generator := NewConfigGen(
		true,
		true,
		false,
		true,
		false,
		"",
	)
	err := generator.Gen("./test-app-init/", NewTestingConfigStructure("test-app-init"))
	assert.NoError(t, err, "generator existed with error")
}

func TestGenerateWithInstanceValues(t *testing.T) {
	generator := NewConfigGen(true, true, false, true, false, "")

	defer func() {
		err := os.RemoveAll("./test-config-init")
		if err != nil {
			t.Errorf("Failed to clean generated config %v", err)
		}
	}()

	err := generator.Gen("./test-config-init/", NewTestingConfigStructure("test-config-init"))
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	sandboxContent, err := os.ReadFile("./test-config-init/sandbox.toml")
	if err != nil {
		t.Fatalf("Failed to read generated sandbox.toml: %v", err)
	}

	sandboxStr := string(sandboxContent)

	if !strings.Contains(sandboxStr, `terraform_version = "1.11.3"`) {
		t.Errorf("Generated sandbox.toml does not contain instance value for terraform_version")
	}

	if !strings.Contains(sandboxStr, `repo = "nuonco/aws-eks-sandbox"`) {
		t.Errorf("Generated sandbox.toml does not contain instance value for public_repo.repo")
	}

	if !strings.Contains(sandboxStr, `directory = "."`) {
		t.Errorf("Generated sandbox.toml does not contain instance value for public_repo.directory")
	}

	if !strings.Contains(sandboxStr, `branch = "main"`) {
		t.Errorf("Generated sandbox.toml does not contain instance value for public_repo.branch")
	}

	installerContent, err := os.ReadFile("./test-config-init/installer.toml")
	if err != nil {
		t.Fatalf("Failed to read generated installer.toml: %v", err)
	}

	installerStr := string(installerContent)

	if !strings.Contains(installerStr, `name = "installer"`) {
		t.Errorf("Generated installer.toml does not contain instance value for name")
	}

	if !strings.Contains(installerStr, `demo_url = "https://nuon.co"`) {
		t.Errorf("Generated installer.toml does not contain instance value for demo_url")
	}

	t.Logf("Successfully verified instance values in generated TOML files")
}
