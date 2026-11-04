package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	policyopa "github.com/nuonco/nuon/pkg/policies/opa"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
)

type EvaluateSinglePolicyRequest struct {
	PolicyID       string `json:"policy_id" validate:"required"`
	PolicyName     string `json:"policy_name"`
	Contents       string `json:"contents" validate:"required"`
	InputJSON      []byte `json:"input_json" validate:"required"`
	InputIndex     int    `json:"input_index"`                // Index of the input document being evaluated
	InputIdentity  string `json:"input_identity" validate:""` // Human-readable input reference
	AllowSkipApply bool   `json:"allow_skip_apply,omitempty"`
}

type EvaluateSinglePolicyResult struct {
	Violations   []PolicyViolation  `json:"violations" temporaljson:"violations,omitempty"`
	Checks       []policyopa.Result `json:"checks,omitempty" temporaljson:"checks,omitempty"`
	SkipFindings []PolicyViolation  `json:"skip_findings,omitempty" temporaljson:"skip_findings,omitempty"`
}

// @temporal-gen-v2 activity
// @max-retries 1
// @schedule-to-close-timeout 2m
// @start-to-close-timeout 1m30s
func (a *Activities) EvaluateSinglePolicy(ctx context.Context, req *EvaluateSinglePolicyRequest) (*EvaluateSinglePolicyResult, error) {
	started := time.Now()
	decision := "pass"
	errorStage := ""
	defer func() { a.policyEvaluationMetrics.record(ctx, started, decision, errorStage) }()

	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(zap.String("policy_id", req.PolicyID))

	l.Info("evaluating policy")

	policy, err := policyopa.Parse(req.Contents)
	if err != nil {
		errorStage = "policy_validation"
		l.Error("invalid OPA policy", zap.Error(err))
		return nil, errors.Wrap(err, "invalid OPA policy")
	}

	for _, check := range policy.Checks {
		if check.Action == "skip_apply" && !req.AllowSkipApply {
			errorStage = "policy_validation"
			return nil, fmt.Errorf("check %s: skip_apply is only supported for Terraform component and sandbox plans", check.ID)
		}
	}

	if len(policy.Checks) > 0 {
		result := &EvaluateSinglePolicyResult{Checks: policy.Evaluate(ctx, req.InputJSON, req.InputIndex, req.InputIdentity)}
		for i, check := range result.Checks {
			for _, finding := range check.Findings {
				if policy.Checks[i].Action == "skip_apply" {
					result.SkipFindings = append(result.SkipFindings, PolicyViolation{
						PolicyID: req.PolicyID, PolicyName: req.PolicyName, CheckID: check.CheckID,
						InputIndex: finding.InputIndex, InputIdentity: finding.InputIdentity,
						Message: finding.Message,
					})
					continue
				}
				result.Violations = append(result.Violations, PolicyViolation{
					PolicyID: req.PolicyID, PolicyName: req.PolicyName, CheckID: check.CheckID,
					InputIndex: finding.InputIndex, InputIdentity: finding.InputIdentity,
					Message: finding.Message, Severity: policy.Checks[i].Action,
				})
			}
			for _, evaluationErr := range check.Errors {
				result.Violations = append(result.Violations, PolicyViolation{
					PolicyID: req.PolicyID, PolicyName: req.PolicyName, CheckID: check.CheckID,
					InputIndex: evaluationErr.InputIndex, InputIdentity: evaluationErr.InputIdentity,
					Message: fmt.Sprintf("Check %s could not be evaluated: %s", check.CheckID, evaluationErr.Message), Severity: "deny",
				})
				errorStage = "check_evaluation"
			}
			if check.Decision == "deny" || check.Decision == "error" {
				decision = "deny"
			} else if check.Decision == "skip_apply" && decision != "deny" {
				decision = "skip_apply"
			} else if check.Decision == "warn" && decision == "pass" {
				decision = "warn"
			}
		}
		l.Info("named policy checks evaluated", zap.Int("check_count", len(result.Checks)), zap.String("decision", decision))
		return result, nil
	}

	var input any
	if err := json.Unmarshal(req.InputJSON, &input); err != nil {
		errorStage = "input_validation"
		l.Error("unable to parse input JSON", zap.Error(err))
		return nil, errors.Wrap(err, "unable to parse input JSON")
	}

	l.Debug("input JSON parsed successfully")

	var violations []PolicyViolation

	denyViolations, err := a.evaluateRule(ctx, l, req.Contents, input, "data.nuon.deny", "deny")
	if err != nil {
		errorStage = "deny_evaluation"
		return nil, errors.Wrap(err, "unable to evaluate deny rules")
	}
	violations = append(violations, denyViolations...)

	warnViolations, err := a.evaluateRule(ctx, l, req.Contents, input, "data.nuon.warn", "warn")
	if err != nil {
		errorStage = "warn_evaluation"
		return nil, errors.Wrap(err, "unable to evaluate warn rules")
	}
	violations = append(violations, warnViolations...)
	if len(denyViolations) > 0 {
		decision = "deny"
	} else if len(warnViolations) > 0 {
		decision = "warn"
	}

	for i := range violations {
		violations[i].PolicyID = req.PolicyID
		violations[i].PolicyName = req.PolicyName
		violations[i].InputIndex = req.InputIndex
		violations[i].InputIdentity = req.InputIdentity
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
				return nil, fmt.Errorf("OPA rule %s must return a set or array of violation messages, got %T", queryStr, expr.Value)
			}
			for _, item := range ruleResults {
				violation := PolicyViolation{
					Severity: severity,
				}

				switch v := item.(type) {
				case string:
					violation.Message = v
				case map[string]interface{}:
					msg, ok := v["msg"].(string)
					if !ok {
						return nil, fmt.Errorf("OPA rule %s returned an object without a string msg field", queryStr)
					}
					violation.Message = msg
				default:
					return nil, fmt.Errorf("OPA rule %s returned unsupported violation type %T", queryStr, item)
				}

				if violation.Message == "" {
					return nil, fmt.Errorf("OPA rule %s returned an empty violation message", queryStr)
				}
				violations = append(violations, violation)
			}
		}
	}

	return violations, nil
}
