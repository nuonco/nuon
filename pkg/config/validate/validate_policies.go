package validate

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/config"
	"gopkg.in/yaml.v2"
)

func ValidatePolicies(a *config.AppConfig) error {
	if a.Policies == nil || len(a.Policies.Policies) < 1 {
		return nil
	}

	for idx, policy := range a.Policies.Policies {
		var obj map[string]any

		if err := yaml.Unmarshal([]byte(policy.Contents), &obj); err != nil {
			return config.ErrConfig{
				Description: fmt.Sprintf("policy %d (%s) was invalid", idx, policy.Type),
				Err:         err,
			}
		}

		if err := validatePolicyType(policy.Type); err != nil {
			return err
		}
	}

	return nil
}

func validatePolicyType(policyType config.AppPolicyType) error {
	switch config.AppPolicyType(policyType) {
	case config.AppPolicyTypeActionWorkflowRunnerJobKyverno,
		config.AppPolicyTypeKubernetesClusterKyverno,
		config.AppPolicyTypeHelmDeployRunnerJobKyverno,
		config.AppPolicyTypeTerraformDeployRunnerJobKyverno:
		return nil
	default:
		return fmt.Errorf("invalid policy type %s", policyType)
	}
}
