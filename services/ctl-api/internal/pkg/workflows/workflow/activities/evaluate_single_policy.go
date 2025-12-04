package activities

import (
	"context"
	"encoding/json"

	"github.com/open-policy-agent/opa/rego"
	"github.com/pkg/errors"
)

type EvaluateSinglePolicyRequest struct {
	PolicyID string `json:"policy_id" validate:"required"`
	Contents string `json:"contents" validate:"required"`
	InputJSON []byte `json:"input_json" validate:"required"`
}

type EvaluateSinglePolicyResult struct {
	Violations []PolicyViolation `json:"violations" temporaljson:"violations,omitempty"`
}

// @temporal-gen activity
// @max-retries 1
func (a *Activities) EvaluateSinglePolicy(ctx context.Context, req *EvaluateSinglePolicyRequest) (*EvaluateSinglePolicyResult, error) {
	var input interface{}
	if err := json.Unmarshal(req.InputJSON, &input); err != nil {
		return nil, errors.Wrap(err, "unable to parse input JSON")
	}

	query, err := rego.New(
		rego.Query("data.policy.deny"),
		rego.Module("policy.rego", req.Contents),
	).PrepareForEval(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to prepare OPA query")
	}

	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return nil, errors.Wrap(err, "unable to evaluate OPA policy")
	}

	var violations []PolicyViolation
	for _, result := range results {
		for _, expr := range result.Expressions {
			denyResults, ok := expr.Value.([]interface{})
			if !ok {
				continue
			}
			for _, deny := range denyResults {
				violation := PolicyViolation{
					PolicyID: req.PolicyID,
				}

				switch v := deny.(type) {
				case string:
					violation.Message = v
				case map[string]interface{}:
					if msg, ok := v["message"].(string); ok {
						violation.Message = msg
					}
				}

				violations = append(violations, violation)
			}
		}
	}

	return &EvaluateSinglePolicyResult{
		Violations: violations,
	}, nil
}
