package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

func fakePermissionsConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalPermissionsConfig(), nil
}

func GetMinimalPermissionsConfig() *config.PermissionsConfig {
	return &config.PermissionsConfig{
		Roles: []*config.AppAWSIAMRole{},
	}
}

func fakePoliciesConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalPoliciesConfig(), nil
}

func GetMinimalPoliciesConfig() *config.PoliciesConfig {
	return &config.PoliciesConfig{
		Policies: []config.AppPolicy{},
	}
}

func fakeSecretsConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalSecretsConfig(), nil
}

func GetMinimalSecretsConfig() *config.SecretsConfig {
	return &config.SecretsConfig{
		Secrets: []*config.AppSecret{},
	}
}

func fakeBreakGlassConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalBreakGlassConfig(), nil
}

func GetMinimalBreakGlassConfig() *config.BreakGlass {
	return &config.BreakGlass{
		Roles: []*config.AppAWSIAMRole{},
	}
}

func fakeStackConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalStackConfig(), nil
}

func GetMinimalStackConfig() *config.StackConfig {
	return &config.StackConfig{
		Type:        "aws-cloudformation",
		Name:        generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
	}
}
