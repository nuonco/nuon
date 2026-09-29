package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

func BuildTerraformComponent(name string) *config.Component {
	return &config.Component{
		Type: config.TerraformModuleComponentType,
		Name: name,
		TerraformModule: &config.TerraformModuleComponentConfig{
			TerraformVersion: "latest",
			PublicRepo: &config.PublicRepoConfig{
				Repo:      "https://github.com/nuonco/nuon-terraform-starter",
				Directory: "/",
				Branch:    "main",
			},
			EnvVarMap: map[string]string{},
			VarsMap:   map[string]string{},
		},
	}
}

func BuildHelmComponent(name string) *config.Component {
	return &config.Component{
		Type: config.HelmChartComponentType,
		Name: name,
		HelmChart: &config.HelmChartComponentConfig{
			ChartName: generics.GetFakeObj[string](),
			Namespace: "default",
			PublicRepo: &config.PublicRepoConfig{
				Repo:      "https://github.com/nuonco/helm-starter",
				Directory: "/charts",
				Branch:    "main",
			},
			ValuesMap: map[string]string{},
		},
	}
}

func BuildDockerBuildComponent(name string) *config.Component {
	return &config.Component{
		Type: config.DockerBuildComponentType,
		Name: name,
		DockerBuild: &config.DockerBuildComponentConfig{
			Dockerfile: "Dockerfile",
			PublicRepo: &config.PublicRepoConfig{
				Repo:      "https://github.com/nuonco/docker-starter",
				Directory: "/",
				Branch:    "main",
			},
			EnvVarMap: map[string]string{},
		},
	}
}

func BuildKubernetesManifestComponent(name string) *config.Component {
	return &config.Component{
		Type: config.KubernetesManifestComponentType,
		Name: name,
		KubernetesManifest: &config.KubernetesManifestComponentConfig{
			PublicRepo: &config.PublicRepoConfig{
				Repo:      "https://github.com/nuonco/k8s-manifests",
				Directory: "/manifests",
				Branch:    "main",
			},
		},
	}
}

func BuildJobComponent(name string) *config.Component {
	return &config.Component{
		Type: config.JobComponentType,
		Name: name,
		Job: &config.JobComponentConfig{
			ImageURL:  "ubuntu",
			Tag:       "latest",
			Cmd:       []string{"echo", "hello"},
			EnvVarMap: map[string]string{},
			Args:      []string{},
		},
	}
}

func BuildExternalImageComponent(name string) *config.Component {
	return &config.Component{
		Type: config.ExternalImageComponentType,
		Name: name,
		ExternalImage: &config.ExternalImageComponentConfig{
			PublicImageConfig: &config.PublicImageConfig{
				ImageURL: "nginx",
				Tag:      "latest",
			},
		},
	}
}

func fakeTerraformComponent(v reflect.Value) (interface{}, error) {
	return BuildTerraformComponent(generics.GetFakeObj[string]()), nil
}

func fakeHelmComponent(v reflect.Value) (interface{}, error) {
	return BuildHelmComponent(generics.GetFakeObj[string]()), nil
}

func fakeDockerComponent(v reflect.Value) (interface{}, error) {
	return BuildDockerBuildComponent(generics.GetFakeObj[string]()), nil
}

func fakeKubernetesManifestComponent(v reflect.Value) (interface{}, error) {
	return BuildKubernetesManifestComponent(generics.GetFakeObj[string]()), nil
}

func fakeJobComponent(v reflect.Value) (interface{}, error) {
	return BuildJobComponent(generics.GetFakeObj[string]()), nil
}

func fakeExternalImageComponent(v reflect.Value) (interface{}, error) {
	return BuildExternalImageComponent(generics.GetFakeObj[string]()), nil
}
