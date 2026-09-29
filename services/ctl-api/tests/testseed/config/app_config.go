package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

func fakeAppConfig(v reflect.Value) (interface{}, error) {
	return &config.AppConfig{
		Version:     "1",
		DisplayName: generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
		Sandbox:     BuildMinimalSandboxConfig(),
		Runner:      BuildMinimalRunnerConfig(),
		Components:  config.ComponentList{},
		Actions:     []*config.ActionConfig{},
	}, nil
}

func BuildMinimalAppConfig() *config.AppConfig {
	return &config.AppConfig{
		Version:     "1",
		DisplayName: generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
		Sandbox:     BuildMinimalSandboxConfig(),
		Runner:      BuildMinimalRunnerConfig(),
		Components:  config.ComponentList{},
		Actions:     []*config.ActionConfig{},
	}
}

func BuildFullAppConfig() *config.AppConfig {
	return &config.AppConfig{
		Version:     "1",
		DisplayName: generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
		Sandbox:     BuildMinimalSandboxConfig(),
		Runner:      BuildMinimalRunnerConfig(),
		Components: config.ComponentList{
			BuildTerraformComponent("terraform-component"),
			BuildHelmComponent("helm-component"),
			BuildDockerBuildComponent("docker-build-component"),
			BuildKubernetesManifestComponent("k8s-manifest-component"),
			BuildExternalImageComponent("external-image-component"),
		},
		Actions: []*config.ActionConfig{},
	}
}
