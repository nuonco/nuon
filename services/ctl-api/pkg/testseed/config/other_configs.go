package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

// Permissions Config Fakers

func fakePermissionsConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalPermissionsConfig(), nil
}

// GetMinimalPermissionsConfig returns a minimal valid PermissionsConfig.
// Returns an empty config (no roles) which is valid.
func GetMinimalPermissionsConfig() *config.PermissionsConfig {
	return &config.PermissionsConfig{
		Roles: []*config.AppAWSIAMRole{},
	}
}

// Policies Config Fakers

func fakePoliciesConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalPoliciesConfig(), nil
}

// GetMinimalPoliciesConfig returns a minimal valid PoliciesConfig.
// Returns an empty config (no policies) which is valid.
func GetMinimalPoliciesConfig() *config.PoliciesConfig {
	return &config.PoliciesConfig{
		Policies: []config.AppPolicy{},
	}
}

// Secrets Config Fakers

func fakeSecretsConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalSecretsConfig(), nil
}

// GetMinimalSecretsConfig returns a minimal valid SecretsConfig.
// Returns an empty config (no secrets) which is valid.
func GetMinimalSecretsConfig() *config.SecretsConfig {
	return &config.SecretsConfig{
		Secrets: []*config.AppSecret{},
	}
}

// BreakGlass Config Fakers

func fakeBreakGlassConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalBreakGlassConfig(), nil
}

// GetMinimalBreakGlassConfig returns a minimal valid BreakGlass config.
// Returns an empty config (no roles) which is valid.
func GetMinimalBreakGlassConfig() *config.BreakGlass {
	return &config.BreakGlass{
		Roles: []*config.AppAWSIAMRole{},
	}
}

// Stack Config Fakers

func fakeStackConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalStackConfig(), nil
}

// GetMinimalStackConfig returns a minimal valid StackConfig.
func GetMinimalStackConfig() *config.StackConfig {
	return &config.StackConfig{
		Type:        "aws-cloudformation",
		Name:        generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
	}
}
