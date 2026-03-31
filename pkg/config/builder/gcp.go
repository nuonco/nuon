package builder

import "github.com/nuonco/nuon/pkg/config"

type gcpBuilder struct {
	handlers map[string]AttributeHandler
}

func newGCPBuilder() *gcpBuilder {
	b := &gcpBuilder{}
	b.handlers = map[string]AttributeHandler{
		AttributeTerraform:     b.applyTerraform,
		AttributeHelmCharts:    b.applyHelmCharts,
		AttributeKubernetes:    b.applyKubernetes,
		AttributeDockerImage:   b.applyDockerImage,
		AttributeCustomScripts: b.applyCustomScripts,
	}
	return b
}

func (b *gcpBuilder) Build(appAttributes []string) (*config.AppConfig, error) {
	cfg := &config.AppConfig{
		Version: "v1",
		Sandbox: &config.AppSandboxConfig{
			TerraformVersion: "1.11.3",
			PublicRepo: &config.PublicRepoConfig{
				Repo:      "nuonco/sandboxes",
				Directory: "gcp",
				Branch:    "main",
			},
		},
		Runner: &config.AppRunnerConfig{
			RunnerType: "gcp",
		},
		Stack: &config.StackConfig{
			Type:        "gcp-terraform",
			Name:        "nuon-onboarding-{{.nuon.install.id}}",
			Description: "Nuon onboarding stack",
		},
	}

	for _, attr := range appAttributes {
		if handler, ok := b.handlers[attr]; ok {
			handler(cfg)
		}
	}

	return cfg, nil
}

func (b *gcpBuilder) applyTerraform(cfg *config.AppConfig) {
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

func (b *gcpBuilder) applyHelmCharts(cfg *config.AppConfig) {
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

func (b *gcpBuilder) applyKubernetes(cfg *config.AppConfig) {
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

func (b *gcpBuilder) applyDockerImage(cfg *config.AppConfig) {
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

func (b *gcpBuilder) applyCustomScripts(cfg *config.AppConfig) {
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
