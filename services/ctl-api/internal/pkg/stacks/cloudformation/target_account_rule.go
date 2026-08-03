package cloudformation

import (
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// injectTargetAccountRule adds a CloudFormation Rules section asserting that the
// stack is being applied in the install's target AWS account. CloudFormation
// evaluates Rules at create/update time before touching any resource, so a
// template applied in the wrong account fails immediately with the assertion
// description instead of partially provisioning.
//
// goformation's Template struct has no Rules field, and Template.JSON() runs the
// output through intrinsics post-processing — so the section is injected into the
// serialized JSON after that step, where a literal {"Ref": ...} survives as-is.
// A no-op when TargetAWSAccountID is unset, leaving those templates byte-identical.
func injectTargetAccountRule(tmplJSON []byte, inp *stacks.TemplateInput) ([]byte, error) {
	if inp.TargetAWSAccountID == "" {
		return tmplJSON, nil
	}

	var tmpl map[string]json.RawMessage
	if err := json.Unmarshal(tmplJSON, &tmpl); err != nil {
		return nil, fmt.Errorf("unable to unmarshal template for target account rule: %w", err)
	}

	rules := map[string]any{
		"NuonTargetAWSAccount": map[string]any{
			"Assertions": []any{
				map[string]any{
					"Assert": map[string]any{
						"Fn::Equals": []any{
							map[string]any{"Ref": "AWS::AccountId"},
							inp.TargetAWSAccountID,
						},
					},
					"AssertDescription": fmt.Sprintf(
						"This template is pinned to AWS account %s and cannot be deployed to a different account (Nuon install %s).",
						inp.TargetAWSAccountID, inp.Install.ID,
					),
				},
			},
		},
	}

	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal target account rule: %w", err)
	}
	tmpl["Rules"] = rulesJSON

	out, err := json.MarshalIndent(tmpl, "", " ")
	if err != nil {
		return nil, fmt.Errorf("unable to marshal template with target account rule: %w", err)
	}

	return out, nil
}
