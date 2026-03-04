package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

// Component Fakers

// GetTerraformComponent returns a fake terraform module component with the given name.
func GetTerraformComponent(name string) *config.Component {
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

// GetHelmComponent returns a fake helm chart component with the given name.
func GetHelmComponent(name string) *config.Component {
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

// GetDockerBuildComponent returns a fake docker build component with the given name.
func GetDockerBuildComponent(name string) *config.Component {
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

// GetKubernetesManifestComponent returns a fake kubernetes manifest component with the given name.
func GetKubernetesManifestComponent(name string) *config.Component {
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

// GetJobComponent returns a fake job component with the given name.
func GetJobComponent(name string) *config.Component {
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

// GetExternalImageComponent returns a fake external image component with the given name.
func GetExternalImageComponent(name string) *config.Component {
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

// Component faker providers for struct tags

func fakeTerraformComponent(v reflect.Value) (interface{}, error) {
	return GetTerraformComponent(generics.GetFakeObj[string]()), nil
}

func fakeHelmComponent(v reflect.Value) (interface{}, error) {
	return GetHelmComponent(generics.GetFakeObj[string]()), nil
}

func fakeDockerComponent(v reflect.Value) (interface{}, error) {
	return GetDockerBuildComponent(generics.GetFakeObj[string]()), nil
}

func fakeKubernetesManifestComponent(v reflect.Value) (interface{}, error) {
	return GetKubernetesManifestComponent(generics.GetFakeObj[string]()), nil
}

func fakeJobComponent(v reflect.Value) (interface{}, error) {
	return GetJobComponent(generics.GetFakeObj[string]()), nil
}

func fakeExternalImageComponent(v reflect.Value) (interface{}, error) {
	return GetExternalImageComponent(generics.GetFakeObj[string]()), nil
}
