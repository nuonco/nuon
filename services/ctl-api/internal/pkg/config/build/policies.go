package build

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/validate"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type PolicyInput struct {
	Type            config.AppPolicyType
	Engine          config.AppPolicyEngine
	Name            string
	Description     string
	Contents        string
	Components      []string
	ApplyOnTeardown *bool
}

// PolicyInputsFromConfig leaves Description empty: it has no config form.
func PolicyInputsFromConfig(policies *config.PoliciesConfig) []PolicyInput {
	if policies == nil {
		return nil
	}
	out := make([]PolicyInput, 0, len(policies.Policies))
	for _, policy := range policies.Policies {
		out = append(out, PolicyInput{
			Type:            policy.Type,
			Engine:          policy.Engine,
			Name:            policy.Name,
			Contents:        policy.Contents,
			Components:      policy.Components,
			ApplyOnTeardown: policy.ApplyOnTeardown,
		})
	}
	return out
}

func PoliciesConfig(policies []PolicyInput, appID, appConfigID string) (*app.AppPoliciesConfig, error) {
	if err := validate.ValidateUniquePolicyNames(appPolicies(policies)); err != nil {
		return nil, err
	}

	objs := make([]app.AppPolicyConfig, 0, len(policies))
	for idx, policy := range policies {
		if !generics.SliceContains(policy.Type, app.AllPolicyTypes) {
			return nil, fmt.Errorf("invalid policy type %q: must be one of (%s)", policy.Type, strings.Join(generics.ToStringSlice(app.AllPolicyTypes), ","))
		}

		policyName := policy.Name
		if policyName == "" {
			policyName = fmt.Sprintf("#%d", idx)
		}
		if policy.Engine == config.AppPolicyEngineOPA {
			if err := validate.ValidateOPAPolicyForType(policy.Contents, policy.Type); err != nil {
				return nil, fmt.Errorf("policy %q is invalid: %w", policyName, err)
			}
		}
		if err := validate.ValidatePolicyComponents(policyName, policy.Type, policy.Components); err != nil {
			return nil, err
		}
		if err := validate.ValidatePolicyApplyOnTeardown(policyName, policy.Type, policy.ApplyOnTeardown); err != nil {
			return nil, err
		}

		components := policy.Components
		if len(components) == 0 && validate.ComponentScopedPolicyType(policy.Type) {
			components = []string{"*"}
		}

		objs = append(objs, app.AppPolicyConfig{
			AppID:           appID,
			AppConfigID:     appConfigID,
			Type:            policy.Type,
			Engine:          policy.Engine,
			Name:            policy.Name,
			Description:     policy.Description,
			Contents:        policy.Contents,
			Components:      components,
			ApplyOnTeardown: policy.ApplyOnTeardown,
		})
	}

	return &app.AppPoliciesConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
		Policies:    objs,
	}, nil
}

func appPolicies(policies []PolicyInput) []config.AppPolicy {
	out := make([]config.AppPolicy, len(policies))
	for i, policy := range policies {
		out[i] = config.AppPolicy{
			Type:       policy.Type,
			Engine:     policy.Engine,
			Name:       policy.Name,
			Contents:   policy.Contents,
			Components: policy.Components,
		}
	}
	return out
}
