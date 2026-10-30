package validate

import (
	"fmt"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/nuonco/nuon/pkg/config"
	policyopa "github.com/nuonco/nuon/pkg/policies/opa"
)

func ValidatePolicies(a *config.AppConfig) error {
	return ValidatePoliciesWithLogger(a, zap.NewNop())
}

func ValidateOPAPolicy(contents string) error {
	_, err := policyopa.Parse(contents)
	return err
}

func ValidatePoliciesWithLogger(a *config.AppConfig, l *zap.Logger) error {
	if a.Policies == nil || len(a.Policies.Policies) < 1 {
		l.Debug("no policies to validate")
		return nil
	}

	l.Info("validating policies", zap.Int("count", len(a.Policies.Policies)))

	for idx, policy := range a.Policies.Policies {
		policyLogger := l.With(
			zap.Int("policy_idx", idx),
			zap.String("policy_type", string(policy.Type)),
			zap.String("engine", string(policy.Engine)),
		)

		if policy.Engine == config.AppPolicyEngineOPA {
			if err := ValidateOPAPolicy(policy.Contents); err != nil {
				policyLogger.Error("invalid OPA rego policy", zap.Error(err))
				return config.ErrConfig{
					Description: fmt.Sprintf("policy %d (%s) was invalid rego", idx, policy.Type),
					Err:         err,
				}
			}
			policyLogger.Debug("OPA rego policy parsed successfully")
		} else {
			var obj map[string]any
			if err := yaml.Unmarshal([]byte(policy.Contents), &obj); err != nil {
				policyLogger.Error("invalid YAML policy", zap.Error(err))
				return config.ErrConfig{
					Description: fmt.Sprintf("policy %d (%s) was invalid yaml", idx, policy.Type),
					Err:         err,
				}
			}
			policyLogger.Debug("YAML policy parsed successfully")
		}

		if err := validatePolicyType(policy.Type); err != nil {
			policyLogger.Error("invalid policy type", zap.Error(err))
			return err
		}

		if err := validatePolicyEngine(policy.Engine); err != nil {
			policyLogger.Error("invalid policy engine", zap.Error(err))
			return err
		}

		if err := validatePolicyTypeEngineCompatibility(policy.Type, policy.Engine); err != nil {
			policyLogger.Error("policy type and engine incompatible", zap.Error(err))
			return err
		}

		policyName := policy.Name
		if policyName == "" {
			policyName = fmt.Sprintf("#%d", idx)
		}
		if err := ValidatePolicyComponents(policyName, policy.Type, policy.Components); err != nil {
			policyLogger.Error("invalid policy components", zap.Error(err), zap.String("policy_name", policyName), zap.Strings("components", policy.Components))
			return err
		}
		if err := ValidatePolicyApplyOnTeardown(policyName, policy.Type, policy.ApplyOnTeardown); err != nil {
			policyLogger.Error("invalid policy apply_on_teardown", zap.Error(err), zap.String("policy_name", policyName))
			return err
		}

		policyLogger.Debug("policy validation passed")
	}

	l.Info("all policies validated successfully")
	return nil
}

func ValidatePolicyNames(a *config.AppConfig) error {
	if a == nil || a.Policies == nil {
		return nil
	}
	return ValidateUniquePolicyNames(a.Policies.Policies)
}

func ValidateUniquePolicyNames(policies []config.AppPolicy) error {
	seen := make(map[string]struct{}, len(policies))
	for _, policy := range policies {
		name := policyName(policy)
		if _, ok := seen[name]; ok {
			if name == "" {
				return fmt.Errorf("policies must have unique names")
			}
			return fmt.Errorf("duplicate policy name %q", name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func policyName(policy config.AppPolicy) string {
	if policy.Name != "" {
		return policy.Name
	}
	policy.SetNameFromContents()
	return policy.Name
}

func validatePolicyType(policyType config.AppPolicyType) error {
	switch policyType {
	case config.AppPolicyTypeKubernetesCluster,
		config.AppPolicyTypeTerraformModule,
		config.AppPolicyTypeHelmChart,
		config.AppPolicyTypeKubernetesManifest,
		config.AppPolicyTypeDockerBuild,
		config.AppPolicyTypeContainerImage,
		config.AppPolicyTypePulumi,
		config.AppPolicyTypeSandbox:
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
		config.AppPolicyTypeContainerImage,
		config.AppPolicyTypePulumi,
		config.AppPolicyTypeSandbox:
		// component-based and sandbox policy types only support OPA engine
		if engine != config.AppPolicyEngineOPA {
			return fmt.Errorf("policy type %s requires engine %s, got %s", policyType, config.AppPolicyEngineOPA, engine)
		}
	}

	return nil
}

// ComponentScopedPolicyType reports whether a policy type is evaluated per
// component. Sandbox and kubernetes_cluster policies are not component-scoped
// and ignore components.
func ComponentScopedPolicyType(policyType config.AppPolicyType) bool {
	switch policyType {
	case config.AppPolicyTypeTerraformModule,
		config.AppPolicyTypeHelmChart,
		config.AppPolicyTypeKubernetesManifest,
		config.AppPolicyTypeDockerBuild,
		config.AppPolicyTypeContainerImage,
		config.AppPolicyTypePulumi:
		return true
	default:
		return false
	}
}

// ValidatePolicyComponents validates the components list for a single policy.
// It is shared by config validation (nuon CLI sync) and the ctl-api create
// endpoint so both enforce the same rules.
func ValidatePolicyComponents(policyName string, policyType config.AppPolicyType, components []string) error {
	hasWildcard := false
	for _, c := range components {
		if c == "" {
			return fmt.Errorf("policy %q has an empty component name in its components list", policyName)
		}
		if c == "*" {
			hasWildcard = true
		}
	}

	if hasWildcard && len(components) > 1 {
		return fmt.Errorf("policy %q (type %s) combines wildcard \"*\" with other component names", policyName, policyType)
	}

	return nil
}

// teardownEvaluatedPolicyType reports whether a policy type is evaluated
// against teardown and deprovision plans.
func teardownEvaluatedPolicyType(policyType config.AppPolicyType) bool {
	switch policyType {
	case config.AppPolicyTypeTerraformModule,
		config.AppPolicyTypeHelmChart,
		config.AppPolicyTypeKubernetesManifest,
		config.AppPolicyTypePulumi,
		config.AppPolicyTypeSandbox:
		return true
	default:
		return false
	}
}

// ValidatePolicyApplyOnTeardown validates apply_on_teardown for a single policy.
// It is shared by config validation (nuon CLI sync) and the ctl-api create
// endpoint so both enforce the same rules.
func ValidatePolicyApplyOnTeardown(policyName string, policyType config.AppPolicyType, applyOnTeardown *bool) error {
	if applyOnTeardown != nil && !teardownEvaluatedPolicyType(policyType) {
		return fmt.Errorf("policy %q (type %s) does not support apply_on_teardown; it is not evaluated on teardown", policyName, policyType)
	}
	return nil
}
