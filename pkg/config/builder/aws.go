package builder

import "github.com/nuonco/nuon/pkg/config"

type awsBuilder struct {
	handlers map[string]AttributeHandler
}

func newAWSBuilder() *awsBuilder {
	b := &awsBuilder{}
	b.handlers = map[string]AttributeHandler{
		AttributeTerraform:     b.applyTerraform,
		AttributeHelmCharts:    b.applyHelmCharts,
		AttributeKubernetes:    b.applyKubernetes,
		AttributeLambda:        b.applyLambda,
		AttributeDockerImage:   b.applyDockerImage,
		AttributeCustomScripts: b.applyCustomScripts,
	}
	return b
}

func (b *awsBuilder) Build(appAttributes []string) (*config.AppConfig, error) {
	cfg := &config.AppConfig{
		Version: "v1",
		Sandbox: &config.AppSandboxConfig{
			TerraformVersion: "1.11.3",
			PublicRepo: &config.PublicRepoConfig{
				Repo:      "nuonco/sandboxes",
				Directory: "aws-eks",
				Branch:    "main",
			},
		},
		Runner: &config.AppRunnerConfig{
			RunnerType:    "aws",
			HelmDriver:    "configmap",
			InitScriptURL: "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init-mng-v2.sh",
		},
		Permissions: &config.PermissionsConfig{
			ProvisionRole: &config.AppAWSIAMRole{
				Type:        "provision",
				Name:        "{{.nuon.install.id}}-provision",
				Description: "provision the sandbox and components; trigger actions.",
				DisplayName: "provision role",
				Policies: []config.AppAWSIAMPolicy{
					{ManagedPolicyName: "AdministratorAccess"},
				},
			},
			MaintenanceRole: &config.AppAWSIAMRole{
				Type:        "maintenance",
				Name:        "{{.nuon.install.id}}-maintenance",
				Description: "operate and remediate the app's components and use actions.",
				DisplayName: "maintenance role",
				Policies: []config.AppAWSIAMPolicy{
					{ManagedPolicyName: "AdministratorAccess"},
				},
			},
			DeprovisionRole: &config.AppAWSIAMRole{
				Type:        "deprovision",
				Name:        "{{.nuon.install.id}}-deprovision",
				Description: "deprovision sandbox and components.",
				DisplayName: "deprovision role",
				Policies: []config.AppAWSIAMPolicy{
					{ManagedPolicyName: "AdministratorAccess"},
				},
			},
		},
		Stack: &config.StackConfig{
			Type:                    "aws-cloudformation",
			Name:                    "nuon-onboarding-{{.nuon.install.id}}",
			Description:             "Nuon onboarding stack",
			VPCNestedTemplateURL:    defaultAWSVPCTemplateURL,
			RunnerNestedTemplateURL: defaultAWSRunnerTemplateURL,
		},
	}

	for _, attr := range appAttributes {
		if handler, ok := b.handlers[attr]; ok {
			handler(cfg)
		}
	}

	return cfg, nil
}

func (b *awsBuilder) applyTerraform(cfg *config.AppConfig) {
	cfg.Components = append(cfg.Components, &config.Component{
		Name: "sample-terraform-module",
		Type: config.TerraformModuleComponentType,
		TerraformModule: &config.TerraformModuleComponentConfig{
			TerraformVersion: "1.11.3",
			PublicRepo: &config.PublicRepoConfig{
				Repo:      sampleRepo,
				Directory: "samples/terraform-module",
				Branch:    sampleBranch,
			},
		},
	})
}

func (b *awsBuilder) applyHelmCharts(cfg *config.AppConfig) {
	cfg.Components = append(cfg.Components, &config.Component{
		Name: "sample-helm-chart",
		Type: config.HelmChartComponentType,
		HelmChart: &config.HelmChartComponentConfig{
			ChartName: "sample-chart",
			Namespace: "{{.nuon.install.id}}",
			PublicRepo: &config.PublicRepoConfig{
				Repo:      sampleRepo,
				Directory: "samples/helm-chart",
				Branch:    sampleBranch,
			},
		},
	})
}

func (b *awsBuilder) applyKubernetes(cfg *config.AppConfig) {
	cfg.Components = append(cfg.Components, &config.Component{
		Name: "sample-kubernetes-manifest",
		Type: config.KubernetesManifestComponentType,
		KubernetesManifest: &config.KubernetesManifestComponentConfig{
			Namespace: "{{.nuon.install.id}}",
			PublicRepo: &config.PublicRepoConfig{
				Repo:      sampleRepo,
				Directory: "samples/kubernetes-manifest",
				Branch:    sampleBranch,
			},
			Kustomize: &config.KustomizeConfig{
				Path: ".",
			},
		},
	})
}

func (b *awsBuilder) applyLambda(cfg *config.AppConfig) {
	cfg.Components = append(cfg.Components, &config.Component{
		Name: "sample-lambda",
		Type: config.TerraformModuleComponentType,
		TerraformModule: &config.TerraformModuleComponentConfig{
			TerraformVersion: "1.11.3",
			PublicRepo: &config.PublicRepoConfig{
				Repo:      sampleRepo,
				Directory: "samples/lambda",
				Branch:    sampleBranch,
			},
		},
	})
}

func (b *awsBuilder) applyDockerImage(cfg *config.AppConfig) {
	cfg.Components = append(cfg.Components, &config.Component{
		Name: "sample-docker-build",
		Type: config.DockerBuildComponentType,
		DockerBuild: &config.DockerBuildComponentConfig{
			Dockerfile: "Dockerfile",
			PublicRepo: &config.PublicRepoConfig{
				Repo:      sampleRepo,
				Directory: "samples/docker-build",
				Branch:    sampleBranch,
			},
		},
	})
}

func (b *awsBuilder) applyCustomScripts(cfg *config.AppConfig) {
	cfg.Components = append(cfg.Components, &config.Component{
		Name: "sample-custom-script",
		Type: config.TerraformModuleComponentType,
		TerraformModule: &config.TerraformModuleComponentConfig{
			TerraformVersion: "1.11.3",
			PublicRepo: &config.PublicRepoConfig{
				Repo:      sampleRepo,
				Directory: "samples/custom-script",
				Branch:    sampleBranch,
			},
		},
	})
}
