package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

func fakePermissionsConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalPermissionsConfig(), nil
}

func BuildMinimalPermissionsConfig() *config.PermissionsConfig {
	return &config.PermissionsConfig{
		Roles: []*config.AppAWSIAMRole{},
	}
}

func fakePoliciesConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalPoliciesConfig(), nil
}

func BuildMinimalPoliciesConfig() *config.PoliciesConfig {
	return &config.PoliciesConfig{
		Policies: []config.AppPolicy{},
	}
}

func fakeSecretsConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalSecretsConfig(), nil
}

func BuildMinimalSecretsConfig() *config.SecretsConfig {
	return &config.SecretsConfig{
		Secrets: []*config.AppSecret{},
	}
}

func fakeBreakGlassConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalBreakGlassConfig(), nil
}

func BuildMinimalBreakGlassConfig() *config.BreakGlass {
	return &config.BreakGlass{
		Roles: []*config.AppAWSIAMRole{},
	}
}

func fakeStackConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalStackConfig(), nil
}

func BuildMinimalStackConfig() *config.StackConfig {
	return &config.StackConfig{
		Type:        "aws-cloudformation",
		Name:        generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
	}
}
