package activities

import (
	"context"
	"encoding/json"

	"github.com/open-policy-agent/opa/rego"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
)

type EvaluateSinglePolicyRequest struct {
	PolicyID  string `json:"policy_id" validate:"required"`
	Contents  string `json:"contents" validate:"required"`
	InputJSON []byte `json:"input_json" validate:"required"`
}

type EvaluateSinglePolicyResult struct {
	Violations []PolicyViolation `json:"violations" temporaljson:"violations,omitempty"`
}

// @temporal-gen activity
// @max-retries 1
// @schedule-to-close-timeout 2m
// @start-to-close-timeout 1m30s
func (a *Activities) EvaluateSinglePolicy(ctx context.Context, req *EvaluateSinglePolicyRequest) (*EvaluateSinglePolicyResult, error) {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(zap.String("policy_id", req.PolicyID))

	l.Info("evaluating policy")

	var input interface{}
	if err := json.Unmarshal(req.InputJSON, &input); err != nil {
		l.Error("unable to parse input JSON", zap.Error(err))
		return nil, errors.Wrap(err, "unable to parse input JSON")
	}

	l.Debug("input JSON parsed successfully")

	query, err := rego.New(
		rego.Query("data.policy.violation"),
		rego.Module("policy.rego", req.Contents),
	).PrepareForEval(ctx)
	if err != nil {
		l.Error("unable to prepare OPA query", zap.Error(err))
		return nil, errors.Wrap(err, "unable to prepare OPA query")
	}

	l.Debug("OPA query prepared successfully")

	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		l.Error("unable to evaluate OPA policy", zap.Error(err))
		return nil, errors.Wrap(err, "unable to evaluate OPA policy")
	}

	l.Debug("OPA policy evaluated", zap.Int("result_count", len(results)))

	var violations []PolicyViolation
	for _, result := range results {
		for _, expr := range result.Expressions {
			denyResults, ok := expr.Value.([]interface{})
			if !ok {
				l.Debug("expression value is not a slice, skipping")
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
					} else if msg, ok := v["msg"].(string); ok {
						violation.Message = msg
					}
				}

				violations = append(violations, violation)
			}
		}
	}

	l.Info("policy evaluation complete", zap.Int("violations_count", len(violations)))

	return &EvaluateSinglePolicyResult{
		Violations: violations,
	}, nil
}
