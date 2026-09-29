package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

func fakeSandboxConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalSandboxConfig(), nil
}

func BuildMinimalSandboxConfig() *config.AppSandboxConfig {
	return &config.AppSandboxConfig{
		TerraformVersion: "latest",
		PublicRepo: &config.PublicRepoConfig{
			Repo:      "https://github.com/nuonco/nuon-terraform-starter",
			Directory: "/",
			Branch:    "main",
		},
		EnvVarMap: map[string]string{},
		VarsMap:   map[string]string{},
	}
}

func BuildMinimalSandboxConfigWithConnectedRepo() *config.AppSandboxConfig {
	return &config.AppSandboxConfig{
		TerraformVersion: "latest",
		ConnectedRepo: &config.ConnectedRepoConfig{
			Repo:      generics.GetFakeObj[string](),
			Directory: "/",
			Branch:    "main",
		},
		EnvVarMap: map[string]string{},
		VarsMap:   map[string]string{},
	}
}
