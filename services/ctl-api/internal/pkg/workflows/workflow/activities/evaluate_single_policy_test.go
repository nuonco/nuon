package activities

import (
	"context"
	_ "embed"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.uber.org/zap"
)

//go:embed testdata/fake_terraform_plan_display_contents.json
var fakeTerraformPlanDisplayContents string

func TestEvaluateRule_TerraformPolicy(t *testing.T) {
	ctx := context.Background()
	l := zap.NewNop()

	a := &Activities{}

	policyPath := "testdata/sample_terraform_policy.rego"

	pctx := &policyContext{
		AppConfigID:   "test-app-config",
		AppID:         "test-app",
		ComponentName: "test-component",
	}

	inputs, _, err := a.prepareTerraformPolicyInputs([]byte(fakeTerraformPlanDisplayContents), pctx)
	require.NoError(t, err, "failed to prepare terraform policy inputs")
	require.Len(t, inputs, 1, "expected exactly one input")

	var input any
	err = json.Unmarshal(inputs[0], &input)
	require.NoError(t, err, "failed to unmarshal input JSON")

	policyContents, err := os.ReadFile(policyPath)
	require.NoError(t, err, "failed to read policy file")

	t.Run("deny rules detect violations", func(t *testing.T) {
		violations, err := a.evaluateRule(ctx, l, string(policyContents), input, "data.nuon.deny", "deny")
		require.NoError(t, err, "evaluateRule failed for deny")

		assert.NotEmpty(t, violations, "expected deny violations")

		var kmsViolation, sgViolation bool
		for _, v := range violations {
			assert.Equal(t, "deny", v.Severity)
			if strings.Contains(v.Message, "aws_kms_key") && strings.Contains(v.Message, "key rotation") {
				kmsViolation = true
			}
			if strings.Contains(v.Message, "aws_security_group_rule") && strings.Contains(v.Message, "all traffic") {
				sgViolation = true
			}
		}
		assert.True(t, kmsViolation, "expected KMS key rotation violation")
		assert.True(t, sgViolation, "expected security group rule violation")
	})

	t.Run("warn rules detect violations", func(t *testing.T) {
		violations, err := a.evaluateRule(ctx, l, string(policyContents), input, "data.nuon.warn", "warn")
		require.NoError(t, err, "evaluateRule failed for warn")

		assert.NotEmpty(t, violations, "expected warn violations")
		for _, v := range violations {
			assert.Equal(t, "warn", v.Severity)
			assert.Contains(t, v.Message, "Environment tag")
		}
	})

	t.Run("non-existent rule returns no violations", func(t *testing.T) {
		violations, err := a.evaluateRule(ctx, l, string(policyContents), input, "data.nuon.nonexistent", "info")
		require.NoError(t, err, "evaluateRule should not error for non-existent rules")
		assert.Empty(t, violations, "non-existent rule should return no violations")
	})
}

func TestEvaluateRule_InvalidPolicy(t *testing.T) {
	ctx := context.Background()
	l := zap.NewNop()

	a := &Activities{}

	invalidPolicy := `package nuon
deny contains msg if {
	this is not valid rego syntax
}`

	input := map[string]interface{}{}

	_, err := a.evaluateRule(ctx, l, invalidPolicy, input, "data.nuon.deny", "deny")
	assert.Error(t, err, "expected error for invalid policy syntax")
}

func TestEvaluateSinglePolicyRejectsWrongPackage(t *testing.T) {
	_, err := executeEvaluateSinglePolicy(t, &EvaluateSinglePolicyRequest{
		PolicyID:  "policy-id",
		Contents:  "package other\ndeny := []",
		InputJSON: []byte("{}"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must use package nuon")
}

func TestEvaluateSinglePolicyRejectsMissingDecisions(t *testing.T) {
	_, err := executeEvaluateSinglePolicy(t, &EvaluateSinglePolicyRequest{
		PolicyID:  "policy-id",
		Contents:  "package nuon\nallow := true",
		InputJSON: []byte("{}"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must define at least one deny or warn rule")
}

func TestEvaluateSinglePolicyAllowsOneDecisionType(t *testing.T) {
	result, err := executeEvaluateSinglePolicy(t, &EvaluateSinglePolicyRequest{
		PolicyID:  "policy-id",
		Contents:  "package nuon\ndeny := []",
		InputJSON: []byte("{}"),
	})
	require.NoError(t, err)
	assert.Empty(t, result.Violations)
}

func executeEvaluateSinglePolicy(t *testing.T, req *EvaluateSinglePolicyRequest) (*EvaluateSinglePolicyResult, error) {
	t.Helper()

	activities := &Activities{}
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestActivityEnvironment()
	env.RegisterActivity(activities.EvaluateSinglePolicy)

	encoded, err := env.ExecuteActivity(activities.EvaluateSinglePolicy, req)
	if err != nil {
		return nil, err
	}

	var result EvaluateSinglePolicyResult
	require.NoError(t, encoded.Get(&result))
	return &result, nil
}

func TestEvaluateRuleRejectsInvalidOutput(t *testing.T) {
	tests := map[string]string{
		"scalar rule": `package nuon
deny := true`,
		"unsupported violation type": `package nuon
deny contains 42 if { true }`,
		"object without message": `package nuon
deny contains {"reason": "blocked"} if { true }`,
		"empty message": `package nuon
deny contains "" if { true }`,
	}

	for name, policy := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := (&Activities{}).evaluateRule(context.Background(), zap.NewNop(), policy, map[string]interface{}{}, "data.nuon.deny", "deny")
			require.Error(t, err)
		})
	}
}

func TestEvaluateRule_EmptyInput(t *testing.T) {
	ctx := context.Background()
	l := zap.NewNop()

	a := &Activities{}

	policy := `package nuon

deny contains msg if {
	input.something == true
	msg := "found something"
}`

	input := map[string]interface{}{}

	violations, err := a.evaluateRule(ctx, l, policy, input, "data.nuon.deny", "deny")
	require.NoError(t, err)
	assert.Empty(t, violations, "empty input should not trigger violations")
}
