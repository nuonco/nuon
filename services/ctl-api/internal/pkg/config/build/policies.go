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
	Type        config.AppPolicyType
	Engine      config.AppPolicyEngine
	Name        string
	Description string
	Contents    string
	Components  []string
}

// PolicyInputsFromConfig leaves Description empty: it has no config form.
func PolicyInputsFromConfig(policies *config.PoliciesConfig) []PolicyInput {
	if policies == nil {
		return nil
	}
	out := make([]PolicyInput, 0, len(policies.Policies))
	for _, policy := range policies.Policies {
		out = append(out, PolicyInput{
			Type:       policy.Type,
			Engine:     policy.Engine,
			Name:       policy.Name,
			Contents:   policy.Contents,
			Components: policy.Components,
		})
	}
	return out
}

func PoliciesConfig(policies []PolicyInput, appID, appConfigID string) (*app.AppPoliciesConfig, error) {
	objs := make([]app.AppPolicyConfig, 0, len(policies))
	for idx, policy := range policies {
		if !generics.SliceContains(policy.Type, app.AllPolicyTypes) {
			return nil, fmt.Errorf("invalid policy type %q: must be one of (%s)", policy.Type, strings.Join(generics.ToStringSlice(app.AllPolicyTypes), ","))
		}

		policyName := policy.Name
		if policyName == "" {
			policyName = fmt.Sprintf("#%d", idx)
		}
		if err := validate.ValidatePolicyComponents(policyName, policy.Type, policy.Components); err != nil {
			return nil, err
		}

		objs = append(objs, app.AppPolicyConfig{
			AppID:       appID,
			AppConfigID: appConfigID,
			Type:        policy.Type,
			Engine:      policy.Engine,
			Name:        policy.Name,
			Description: policy.Description,
			Contents:    policy.Contents,
			Components:  policy.Components,
		})
	}

	return &app.AppPoliciesConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
		Policies:    objs,
	}, nil
}
