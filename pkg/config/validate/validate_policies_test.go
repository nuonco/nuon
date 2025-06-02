package validate

import (
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestValidatePolicyType(t *testing.T) {
	tests := []struct {
		input    config.AppPolicyType
		expected bool
	}{
		{config.AppPolicyTypeActionWorkflowRunnerJobKyverno, false},
		{config.AppPolicyTypeKubernetesClusterKyverno, false},
		{config.AppPolicyTypeHelmDeployRunnerJobKyverno, false},
		{config.AppPolicyTypeTerraformDeployRunnerJobKyverno, false},
		{config.AppPolicyType("invalid_policy_type"), true},
	}

	for _, test := range tests {
		t.Run(string(test.input), func(t *testing.T) {
			err := validatePolicyType(test.input)
			fmt.Println("Testing policy type:", test.input, "Expected error:", test.expected, "Got error:", err)
			assert.Equal(t, (err != nil), test.expected, "Expected error for policy type %s: %v, got: %v", test.input, test.expected, err)
		})
	}
}
