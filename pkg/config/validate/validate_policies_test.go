package validate

import (
	"testing"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestValidatePolicyType(t *testing.T) {
	tests := map[string]struct {
		input    config.AppPolicyType
		expected bool
	}{
		"kubernetes_cluster":  {config.AppPolicyTypeKubernetesCluster, false},
		"terraform_module":    {config.AppPolicyTypeTerraformModule, false},
		"helm_chart":          {config.AppPolicyTypeHelmChart, false},
		"kubernetes_manifest": {config.AppPolicyTypeKubernetesManifest, false},
		"docker_build":        {config.AppPolicyTypeDockerBuild, false},
		"container_image":     {config.AppPolicyTypeContainerImage, false},
		"sandbox":             {config.AppPolicyTypeSandbox, false},
		"invalid_policy_type": {config.AppPolicyType("invalid_policy_type"), true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validatePolicyType(test.input)
			assert.Equal(t, (err != nil), test.expected, "Expected error for policy type %s: %v, got: %v", test.input, test.expected, err)
		})
	}
}

func TestValidatePolicyEngine(t *testing.T) {
	tests := map[string]struct {
		input    config.AppPolicyEngine
		expected bool
	}{
		"kyverno":        {config.AppPolicyEngineKyverno, false},
		"opa":            {config.AppPolicyEngineOPA, false},
		"empty_allowed":  {"", false}, // empty is allowed for backwards compatibility
		"invalid_engine": {config.AppPolicyEngine("invalid_engine"), true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validatePolicyEngine(test.input)
			assert.Equal(t, (err != nil), test.expected, "Expected error for engine %s: %v, got: %v", test.input, test.expected, err)
		})
	}
}

func TestValidatePolicyTypeEngineCompatibility(t *testing.T) {
	tests := map[string]struct {
		policyType config.AppPolicyType
		engine     config.AppPolicyEngine
		expected   bool
	}{
		// kubernetes_cluster only supports kyverno
		"kubernetes_cluster_kyverno":   {config.AppPolicyTypeKubernetesCluster, config.AppPolicyEngineKyverno, false},
		"kubernetes_cluster_opa_error": {config.AppPolicyTypeKubernetesCluster, config.AppPolicyEngineOPA, true},
		// component-based types only support OPA
		"terraform_module_kyverno_error":    {config.AppPolicyTypeTerraformModule, config.AppPolicyEngineKyverno, true},
		"terraform_module_opa":              {config.AppPolicyTypeTerraformModule, config.AppPolicyEngineOPA, false},
		"helm_chart_kyverno_error":          {config.AppPolicyTypeHelmChart, config.AppPolicyEngineKyverno, true},
		"helm_chart_opa":                    {config.AppPolicyTypeHelmChart, config.AppPolicyEngineOPA, false},
		"kubernetes_manifest_kyverno_error": {config.AppPolicyTypeKubernetesManifest, config.AppPolicyEngineKyverno, true},
		"kubernetes_manifest_opa":           {config.AppPolicyTypeKubernetesManifest, config.AppPolicyEngineOPA, false},
		// empty engine skips check
		"kubernetes_cluster_empty_engine": {config.AppPolicyTypeKubernetesCluster, "", false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validatePolicyTypeEngineCompatibility(test.policyType, test.engine)
			assert.Equal(t, (err != nil), test.expected, "Expected error: %v, got: %v", test.expected, err)
		})
	}
}

func TestValidatePolicyComponents(t *testing.T) {
	tests := map[string]struct {
		components []string
		expected   bool
	}{
		"empty":                {[]string{}, false},
		"single_component":     {[]string{"rds_cluster"}, false},
		"multiple_components":  {[]string{"rds_cluster", "vpc"}, false},
		"wildcard_only":        {[]string{"*"}, false},
		"wildcard_with_others": {[]string{"*", "rds_cluster"}, true},
		"empty_component_name": {[]string{"rds_cluster", ""}, true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validatePolicyComponents(test.components)
			assert.Equal(t, (err != nil), test.expected, "Expected error: %v, got: %v", test.expected, err)
		})
	}
}
