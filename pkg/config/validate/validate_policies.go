package validate

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/open-policy-agent/opa/v1/ast"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

func ValidatePolicies(a *config.AppConfig) error {
	return ValidatePoliciesWithLogger(a, zap.NewNop())
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
			if _, err := ast.ParseModule("policy.rego", policy.Contents); err != nil {
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

		policyLogger.Debug("policy validation passed")
	}

	l.Info("all policies validated successfully")
	return nil
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

// componentScopedPolicyType reports whether a policy type is evaluated per
// component and therefore requires a non-empty components list. Sandbox and
// kubernetes_cluster policies are not component-scoped and ignore components.
func componentScopedPolicyType(policyType config.AppPolicyType) bool {
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
	if len(components) == 0 {
		// Component-scoped policies with an empty components list never run,
		// which silently disables them. Require an explicit target (["*"] for
		// all components of the type, or specific component names).
		if componentScopedPolicyType(policyType) {
			return fmt.Errorf("policy %q (type %s) requires a non-empty components list; use [\"*\"] to apply to all components of this type or list specific component names", policyName, policyType)
		}
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
