package validate

import (
	"testing"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestValidatePolicyType(t *testing.T) {
	tests := []struct {
		input    config.AppPolicyType
		expected bool
	}{
		{config.AppPolicyTypeKubernetesCluster, false},
		{config.AppPolicyTypeTerraformModule, false},
		{config.AppPolicyTypeHelmChart, false},
		{config.AppPolicyTypeKubernetesManifest, false},
		{config.AppPolicyTypeDockerBuild, false},
		{config.AppPolicyTypeContainerImage, false},
		{config.AppPolicyTypeJob, false},
		{config.AppPolicyType("invalid_policy_type"), true},
	}

	for _, test := range tests {
		t.Run(string(test.input), func(t *testing.T) {
			err := validatePolicyType(test.input)
			assert.Equal(t, (err != nil), test.expected, "Expected error for policy type %s: %v, got: %v", test.input, test.expected, err)
		})
	}
}

func TestValidatePolicyEngine(t *testing.T) {
	tests := []struct {
		input    config.AppPolicyEngine
		expected bool
	}{
		{config.AppPolicyEngineKyverno, false},
		{config.AppPolicyEngineOPA, false},
		{"", false}, // empty is allowed for backwards compatibility
		{config.AppPolicyEngine("invalid_engine"), true},
	}

	for _, test := range tests {
		t.Run(string(test.input), func(t *testing.T) {
			err := validatePolicyEngine(test.input)
			assert.Equal(t, (err != nil), test.expected, "Expected error for engine %s: %v, got: %v", test.input, test.expected, err)
		})
	}
}

func TestValidatePolicyTypeEngineCompatibility(t *testing.T) {
	tests := []struct {
		policyType config.AppPolicyType
		engine     config.AppPolicyEngine
		expected   bool
	}{
		// kubernetes_cluster only supports kyverno
		{config.AppPolicyTypeKubernetesCluster, config.AppPolicyEngineKyverno, false},
		{config.AppPolicyTypeKubernetesCluster, config.AppPolicyEngineOPA, true},
		// component types support both
		{config.AppPolicyTypeTerraformModule, config.AppPolicyEngineKyverno, false},
		{config.AppPolicyTypeTerraformModule, config.AppPolicyEngineOPA, false},
		{config.AppPolicyTypeHelmChart, config.AppPolicyEngineKyverno, false},
		{config.AppPolicyTypeHelmChart, config.AppPolicyEngineOPA, false},
		// empty engine skips check
		{config.AppPolicyTypeKubernetesCluster, "", false},
	}

	for _, test := range tests {
		t.Run(string(test.policyType)+"_"+string(test.engine), func(t *testing.T) {
			err := validatePolicyTypeEngineCompatibility(test.policyType, test.engine)
			assert.Equal(t, (err != nil), test.expected, "Expected error: %v, got: %v", test.expected, err)
		})
	}
}

func TestValidatePolicyComponents(t *testing.T) {
	tests := []struct {
		name       string
		components []string
		expected   bool
	}{
		{"empty", []string{}, false},
		{"single component", []string{"rds_cluster"}, false},
		{"multiple components", []string{"rds_cluster", "vpc"}, false},
		{"wildcard only", []string{"*"}, false},
		{"wildcard with others", []string{"*", "rds_cluster"}, true},
		{"empty component name", []string{"rds_cluster", ""}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validatePolicyComponents(test.components)
			assert.Equal(t, (err != nil), test.expected, "Expected error: %v, got: %v", test.expected, err)
		})
	}
}
