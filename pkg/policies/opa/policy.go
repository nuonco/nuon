package opa

import (
	"context"
	_ "embed"
	"fmt"
	"sort"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/util"
)

//go:embed check.rego
var Library string

type Check struct {
	ID          string
	Title       string
	Description string
	Action      string
}

type Policy struct {
	Checks   []Check
	compiler *ast.Compiler
}

type Finding struct {
	Message       string `json:"message"`
	InputIndex    int    `json:"input_index"`
	InputIdentity string `json:"input_identity,omitempty"`
}

type EvaluationError struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	InputIndex    int    `json:"input_index"`
	InputIdentity string `json:"input_identity,omitempty"`
}

type Result struct {
	CheckID     string            `json:"check_id"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Decision    string            `json:"decision"`
	Findings    []Finding         `json:"findings"`
	Errors      []EvaluationError `json:"errors"`
	InputCount  int               `json:"input_count"`
}

func UsesNamedChecks(contents string) bool {
	module, err := ast.ParseModule("policy.rego", contents)
	return err == nil && module.Package.Path.String() == "data.nuon.checks"
}

func Parse(contents string) (*Policy, error) {
	module, err := ast.ParseModule("policy.rego", contents)
	if err != nil {
		return nil, err
	}
	if module.Package.Path.String() == "data.nuon" {
		for _, rule := range module.Rules {
			if rule.Head.Name == "deny" || rule.Head.Name == "warn" {
				return &Policy{}, nil
			}
		}
		return nil, fmt.Errorf("OPA policy must define at least one deny or warn rule")
	}
	if module.Package.Path.String() != "data.nuon.checks" {
		return nil, fmt.Errorf("OPA policy must use package nuon (legacy) or nuon.checks (named checks), got %s", module.Package.Path.String())
	}

	module, err = ast.ParseModuleWithOpts("policy.rego", contents, ast.ParserOptions{ProcessAnnotation: true})
	if err != nil {
		return nil, err
	}
	checks := make(map[string]Check)
	for _, rule := range module.Rules {
		name := rule.Head.Ref()[0].String()
		if name == "deny" || name == "warn" {
			return nil, fmt.Errorf("package nuon.checks cannot contain legacy %s rules", name)
		}
		if !strings.HasPrefix(name, "check_") {
			continue
		}
		if name == "check_" || len(rule.Head.Ref()) != 1 || len(rule.Head.Args) != 0 || rule.Head.RuleKind() != ast.MultiValue || rule.Default || rule.Else != nil {
			return nil, fmt.Errorf("%s must be a named finding set: check_name contains finding if { ... }", name)
		}
		checks[name] = Check{ID: name, Title: name}
	}
	if len(checks) == 0 {
		return nil, fmt.Errorf("package nuon.checks must define at least one check_* finding set")
	}

	defaultAction := "deny"
	annotated := make(map[string]bool)
	for _, annotation := range module.Annotations {
		path := annotation.GetTargetPath().String()
		name := strings.TrimPrefix(path, "data.nuon.checks.")
		check, ok := checks[name]
		action, hasAction := annotation.Custom["action"]
		if hasAction {
			if path != "data.nuon.checks" && !ok {
				return nil, fmt.Errorf("%s custom.action is only supported on package nuon.checks or a check_* finding set", path)
			}
			if action != "deny" && action != "warn" && action != "skip_apply" {
				return nil, fmt.Errorf("%s custom.action must be deny, warn, or skip_apply", path)
			}
		}
		if path == "data.nuon.checks" {
			if hasAction {
				defaultAction = action.(string)
			}
			continue
		}
		if !ok {
			continue
		}
		if annotated[name] {
			return nil, fmt.Errorf("%s must have at most one metadata declaration", name)
		}
		annotated[name] = true
		if annotation.Title != "" {
			check.Title = annotation.Title
		}
		check.Description = annotation.Description
		if hasAction {
			check.Action = action.(string)
		}
		checks[name] = check
	}

	library, err := ast.ParseModule("nuon-check.rego", Library)
	if err != nil {
		return nil, err
	}
	compiler := ast.NewCompiler()
	compiler.Compile(map[string]*ast.Module{"policy.rego": module, "nuon-check.rego": library})
	if compiler.Failed() {
		return nil, compiler.Errors
	}
	policy := &Policy{compiler: compiler}
	for _, check := range checks {
		if check.Action == "" {
			check.Action = defaultAction
		}
		policy.Checks = append(policy.Checks, check)
	}
	sort.Slice(policy.Checks, func(i, j int) bool { return policy.Checks[i].ID < policy.Checks[j].ID })
	return policy, nil
}

func (p *Policy) Evaluate(ctx context.Context, inputJSON []byte, inputIndex int, inputIdentity string) []Result {
	var input map[string]any
	inputErr := util.UnmarshalJSON(inputJSON, &input)
	if inputErr == nil && input == nil {
		inputErr = fmt.Errorf("policy input must be a JSON object")
	}
	results := make([]Result, 0, len(p.Checks))
	for _, check := range p.Checks {
		result := Result{
			CheckID: check.ID, Title: check.Title, Description: check.Description,
			Decision: "pass", Findings: []Finding{}, Errors: []EvaluationError{}, InputCount: 1,
		}
		code := "input_invalid"
		err := inputErr
		if err == nil {
			code = "evaluation_failed"
			var query rego.PreparedEvalQuery
			query, err = rego.New(
				rego.Compiler(p.compiler),
				rego.Query("data.nuon.checks."+check.ID),
				rego.StrictBuiltinErrors(true),
			).PrepareForEval(ctx)
			if err == nil {
				var output rego.ResultSet
				output, err = query.Eval(ctx, rego.EvalInput(input))
				if err == nil {
					code = "invalid_result"
					result.Findings, err = decodeFindings(output, inputIndex, inputIdentity)
				}
			}
		}
		if len(result.Findings) > 0 {
			result.Decision = check.Action
		}
		if err != nil {
			result.Decision = "error"
			result.Errors = append(result.Errors, EvaluationError{Code: code, Message: err.Error(), InputIndex: inputIndex, InputIdentity: inputIdentity})
		}
		results = append(results, result)
	}
	return results
}

func decodeFindings(output rego.ResultSet, inputIndex int, inputIdentity string) ([]Finding, error) {
	findings := make([]Finding, 0)
	if len(output) != 1 || len(output[0].Expressions) != 1 {
		return findings, fmt.Errorf("check did not return a finding set")
	}
	items, ok := output[0].Expressions[0].Value.([]any)
	if !ok {
		return findings, fmt.Errorf("check must return a finding set")
	}
	var invalid bool
	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			invalid = true
			continue
		}
		message, messageOK := object["message"].(string)
		if object["kind"] != "finding" || !messageOK || strings.TrimSpace(message) == "" {
			invalid = true
			continue
		}
		findings = append(findings, Finding{Message: message, InputIndex: inputIndex, InputIdentity: inputIdentity})
	}
	if invalid {
		return findings, fmt.Errorf("each finding must have kind finding and a nonempty message; authored errors are not supported")
	}
	return findings, nil
}

func Aggregate(results []Result) []Result {
	byID := make(map[string]Result)
	for _, result := range results {
		current, exists := byID[result.CheckID]
		if !exists {
			current = result
			current.Findings = append([]Finding{}, result.Findings...)
			current.Errors = append([]EvaluationError{}, result.Errors...)
		} else {
			current.Findings = append(current.Findings, result.Findings...)
			current.Errors = append(current.Errors, result.Errors...)
			current.InputCount += result.InputCount
			if decisionRank(result.Decision) > decisionRank(current.Decision) {
				current.Decision = result.Decision
			}
		}
		byID[result.CheckID] = current
	}
	merged := make([]Result, 0, len(byID))
	for _, result := range byID {
		merged = append(merged, result)
	}
	sort.Slice(merged, func(i, j int) bool { return merged[i].CheckID < merged[j].CheckID })
	return merged
}

func decisionRank(decision string) int {
	switch decision {
	case "error":
		return 4
	case "deny":
		return 3
	case "skip_apply":
		return 2
	case "warn":
		return 1
	default:
		return 0
	}
}
