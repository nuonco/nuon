package validate

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
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

		if err := validatePolicyEngine(policy.Engine); err != nil {
			return err
		}

		if err := validatePolicyTypeEngineCompatibility(policy.Type, policy.Engine); err != nil {
			return err
		}

		if err := validatePolicyComponents(policy.Components); err != nil {
			return err
		}
	}

	return nil
}

func validatePolicyType(policyType config.AppPolicyType) error {
	switch policyType {
	case config.AppPolicyTypeKubernetesCluster,
		config.AppPolicyTypeTerraformModule,
		config.AppPolicyTypeHelmChart,
		config.AppPolicyTypeKubernetesManifest:
		return nil
	default:
		return fmt.Errorf("invalid policy type %s", policyType)
	}
}

func validatePolicyEngine(engine config.AppPolicyEngine) error {
	// Empty engine is allowed for backwards compatibility - will default based on type
	if engine == "" {
		return nil
	}

	switch engine {
	case config.AppPolicyEngineKyverno, config.AppPolicyEngineOPA:
		return nil
	default:
		return fmt.Errorf("invalid policy engine %s", engine)
	}
}

func validatePolicyTypeEngineCompatibility(policyType config.AppPolicyType, engine config.AppPolicyEngine) error {
	// If no engine specified, skip compatibility check (will use default)
	if engine == "" {
		return nil
	}

	switch policyType {
	case config.AppPolicyTypeKubernetesCluster:
		// kubernetes_cluster only supports kyverno
		if engine != config.AppPolicyEngineKyverno {
			return fmt.Errorf("policy type %s requires engine %s, got %s", policyType, config.AppPolicyEngineKyverno, engine)
		}
	case config.AppPolicyTypeTerraformModule,
		config.AppPolicyTypeHelmChart,
		config.AppPolicyTypeKubernetesManifest,
		config.AppPolicyTypeDockerBuild,
		config.AppPolicyTypeContainerImage:
		// component-based policy types only support OPA engine
		if engine != config.AppPolicyEngineOPA {
			return fmt.Errorf("policy type %s (component-based) requires engine %s, got %s", policyType, config.AppPolicyEngineOPA, engine)
		}
	}

	return nil
}

func validatePolicyComponents(components []string) error {
	if len(components) == 0 {
		return nil
	}

	// Check for invalid wildcard usage
	hasWildcard := false
	for _, c := range components {
		if c == "*" {
			hasWildcard = true
		}
		if c == "" {
			return fmt.Errorf("empty component name in components list")
		}
	}

	// If wildcard is present, it should be the only element
	if hasWildcard && len(components) > 1 {
		return fmt.Errorf("wildcard \"*\" cannot be combined with other component names")
	}

	return nil
}
