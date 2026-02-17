package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

// Permissions Config Fakers

func fakePermissionsConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalPermissionsConfig(), nil
}

// BuildMinimalPermissionsConfig returns a minimal valid PermissionsConfig.
// Returns an empty config (no roles) which is valid.
func BuildMinimalPermissionsConfig() *config.PermissionsConfig {
	return &config.PermissionsConfig{
		Roles: []*config.AppAWSIAMRole{},
	}
}

// Policies Config Fakers

func fakePoliciesConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalPoliciesConfig(), nil
}

// BuildMinimalPoliciesConfig returns a minimal valid PoliciesConfig.
// Returns an empty config (no policies) which is valid.
func BuildMinimalPoliciesConfig() *config.PoliciesConfig {
	return &config.PoliciesConfig{
		Policies: []config.AppPolicy{},
	}
}

// Secrets Config Fakers

func fakeSecretsConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalSecretsConfig(), nil
}

// BuildMinimalSecretsConfig returns a minimal valid SecretsConfig.
// Returns an empty config (no secrets) which is valid.
func BuildMinimalSecretsConfig() *config.SecretsConfig {
	return &config.SecretsConfig{
		Secrets: []*config.AppSecret{},
	}
}

// BreakGlass Config Fakers

func fakeBreakGlassConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalBreakGlassConfig(), nil
}

// BuildMinimalBreakGlassConfig returns a minimal valid BreakGlass config.
// Returns an empty config (no roles) which is valid.
func BuildMinimalBreakGlassConfig() *config.BreakGlass {
	return &config.BreakGlass{
		Roles: []*config.AppAWSIAMRole{},
	}
}

// Stack Config Fakers

func fakeStackConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalStackConfig(), nil
}

// BuildMinimalStackConfig returns a minimal valid StackConfig.
func BuildMinimalStackConfig() *config.StackConfig {
	return &config.StackConfig{
		Type:        "aws-cloudformation",
		Name:        generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
	}
}
