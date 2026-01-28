package activities

import (
	"context"
	"encoding/json"

	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
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

	var input any
	if err := json.Unmarshal(req.InputJSON, &input); err != nil {
		l.Error("unable to parse input JSON", zap.Error(err))
		return nil, errors.Wrap(err, "unable to parse input JSON")
	}

	l.Debug("input JSON parsed successfully")

	var violations []PolicyViolation

	denyViolations, err := a.evaluateRule(ctx, l, req.Contents, input, "data.nuon.deny", "deny")
	if err != nil {
		return nil, errors.Wrap(err, "unable to evaluate deny rules")
	}
	violations = append(violations, denyViolations...)

	warnViolations, err := a.evaluateRule(ctx, l, req.Contents, input, "data.nuon.warn", "warn")
	if err != nil {
		return nil, errors.Wrap(err, "unable to evaluate warn rules")
	}
	violations = append(violations, warnViolations...)

	for i := range violations {
		violations[i].PolicyID = req.PolicyID
	}

	l.Info("policy evaluation complete",
		zap.Int("deny_count", len(denyViolations)),
		zap.Int("warn_count", len(warnViolations)),
	)

	return &EvaluateSinglePolicyResult{
		Violations: violations,
	}, nil
}

func (a *Activities) evaluateRule(
	ctx context.Context,
	l *zap.Logger,
	contents string,
	input any,
	queryStr string,
	severity string,
) ([]PolicyViolation, error) {
	l.Debug("preparing OPA query", zap.String("query", queryStr))

	query, err := rego.New(
		rego.Query(queryStr),
		rego.Module("policy.rego", contents),
	).PrepareForEval(ctx)
	if err != nil {
		l.Error("unable to prepare OPA query", zap.String("query", queryStr), zap.Error(err))
		return nil, errors.Wrapf(err, "unable to prepare OPA query for %s", queryStr)
	}

	l.Debug("OPA query prepared successfully", zap.String("query", queryStr))

	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		l.Error("unable to evaluate OPA policy", zap.String("query", queryStr), zap.Error(err))
		return nil, errors.Wrapf(err, "unable to evaluate OPA policy for %s", queryStr)
	}

	l.Debug("OPA policy evaluated", zap.String("query", queryStr), zap.Int("result_count", len(results)))

	var violations []PolicyViolation
	for _, result := range results {
		for _, expr := range result.Expressions {
			ruleResults, ok := expr.Value.([]interface{})
			if !ok {
				l.Warn("expression value is not a slice, skipping", zap.String("query", queryStr))
				continue
			}
			for _, item := range ruleResults {
				violation := PolicyViolation{
					Severity: severity,
				}

				switch v := item.(type) {
				case string:
					violation.Message = v
				case map[string]interface{}:
					if msg, ok := v["msg"].(string); ok {
						violation.Message = msg
					}
				}

				if violation.Message != "" {
					violations = append(violations, violation)
				}
			}
		}
	}

	return violations, nil
}
