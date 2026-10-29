package opa_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	policyopa "github.com/nuonco/nuon/pkg/policies/opa"
)

func TestNamedChecks(t *testing.T) {
	policy, err := policyopa.Parse(`package nuon.checks
import data.nuon.lib.check

# METADATA
# title: Bounded replicas
# custom:
#   action: warn
check_replicas contains check.finding(sprintf("too many replicas: %d", [input.replicas])) if {
    input.replicas > 3
}
check_replicas contains check.finding("no replicas") if { input.replicas == 0 }

check_public contains check.finding(sprintf("public resource: %s", [input.name])) if { input.public }
check_json contains check.finding("known violation")
check_json contains check.finding("parsed") if { json.unmarshal(input.document).blocked }
helper(x) := x + 1
`)
	require.NoError(t, err)
	require.Len(t, policy.Checks, 3)
	first := policy.Evaluate(context.Background(), []byte(`{"replicas":5,"public":true,"name":"example","document":"{}"}`), 0, "Deployment/default/example")
	second := policy.Evaluate(context.Background(), []byte(`{"replicas":1,"public":false,"document":"invalid"}`), 1, "Deployment/default/other")
	results := policyopa.Aggregate(append(first, second...))
	require.Len(t, results, 3)
	assert.Equal(t, "check_json", results[0].CheckID)
	assert.Equal(t, "error", results[0].Decision)
	require.Len(t, results[0].Findings, 1)
	assert.Equal(t, "known violation", results[0].Findings[0].Message)
	assert.Equal(t, "Deployment/default/example", results[0].Findings[0].InputIdentity)
	require.Len(t, results[0].Errors, 1)
	assert.Equal(t, "evaluation_failed", results[0].Errors[0].Code)
	assert.Equal(t, "Deployment/default/other", results[0].Errors[0].InputIdentity)
	assert.Equal(t, "deny", results[1].Decision)
	require.Len(t, results[1].Findings, 1)
	encoded, err := json.Marshal(results[1].Findings[0])
	require.NoError(t, err)
	assert.JSONEq(t, `{"message":"public resource: example","input_index":0,"input_identity":"Deployment/default/example"}`, string(encoded))
	assert.Equal(t, "Bounded replicas", results[2].Title)
	assert.Equal(t, "warn", results[2].Decision)
	require.Len(t, results[2].Findings, 1)
	assert.Equal(t, "too many replicas: 5", results[2].Findings[0].Message)
	for _, result := range results {
		assert.Equal(t, 2, result.InputCount)
	}
	assert.Equal(t, "deny", first[0].Decision)
	assert.Empty(t, first[0].Errors)
	assert.Equal(t, "error", second[0].Decision)
	assert.Empty(t, second[0].Findings)
	require.Len(t, second[0].Errors, 1)
	assert.Equal(t, "evaluation_failed", second[0].Errors[0].Code)
	assert.Equal(t, "pass", second[1].Decision)
	assert.Equal(t, "pass", second[2].Decision)

	for _, input := range []string{"null", "[]", "invalid", "{} {}"} {
		results := policy.Evaluate(context.Background(), []byte(input), 0, "example")
		require.Len(t, results, 3)
		for _, result := range results {
			assert.Equal(t, "error", result.Decision)
			assert.Equal(t, "input_invalid", result.Errors[0].Code)
		}
	}
}

func TestPackageActionInheritance(t *testing.T) {
	for _, tt := range []struct {
		name     string
		metadata string
		decision string
	}{
		{name: "no action", decision: "deny"},
		{name: "deny default", metadata: "# custom:\n#   action: deny\n", decision: "deny"},
		{name: "warn default", metadata: "# custom:\n#   action: warn\n", decision: "warn"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := policyopa.Parse(`# METADATA
# title: Package title
# description: Package description
` + tt.metadata + `package nuon.checks
import data.nuon.lib.check

check_bare contains check.finding("bare")
# METADATA
# custom:
#   action: deny
check_deny contains check.finding("deny")
check_empty contains check.finding("never") if { false }
check_error contains check.finding("invalid JSON") if { json.unmarshal(input.document) }

check_titled contains check.finding("first clause")
# METADATA
# title: Check title
# description: Check description
check_titled contains check.finding("second clause")

check_warn contains check.finding("first warning")
# METADATA
# custom:
#   action: warn
check_warn contains check.finding("second warning")

# METADATA
# title: Helper title
helper := true
`)
			require.NoError(t, err)
			results := policy.Evaluate(context.Background(), []byte(`{"document":"invalid"}`), 0, "example")
			require.Len(t, results, 6)
			assert.Equal(t, "check_bare", results[0].Title)
			assert.Empty(t, results[0].Description)
			assert.Equal(t, tt.decision, results[0].Decision)
			assert.Equal(t, "deny", results[1].Decision)
			assert.Equal(t, "pass", results[2].Decision)
			assert.Equal(t, "error", results[3].Decision)
			require.Len(t, results[3].Errors, 1)
			assert.Equal(t, "evaluation_failed", results[3].Errors[0].Code)
			assert.Equal(t, "Check title", results[4].Title)
			assert.Equal(t, "Check description", results[4].Description)
			assert.Equal(t, tt.decision, results[4].Decision)
			assert.Len(t, results[4].Findings, 2)
			assert.Equal(t, "warn", results[5].Decision)
			assert.Len(t, results[5].Findings, 2)

			separate, err := policyopa.Parse(`package nuon.checks
import data.nuon.lib.check
check_bare contains check.finding("separate policy")`)
			require.NoError(t, err)
			assert.Equal(t, "deny", separate.Evaluate(context.Background(), []byte(`{}`), 0, "example")[0].Decision)
		})
	}
}

func TestActionMetadataValidation(t *testing.T) {
	for _, action := range []string{"skip_apply", "null", "true", "[deny]", "{name: deny}"} {
		for _, target := range []string{"package", "check"} {
			t.Run(target+"/"+action, func(t *testing.T) {
				metadata := fmt.Sprintf("# METADATA\n# custom:\n#   action: %s\n", action)
				source := metadata + "package nuon.checks\ncheck_example contains {}"
				if target == "check" {
					source = "package nuon.checks\n" + metadata + "check_example contains {}"
				}
				_, err := policyopa.Parse(source)
				require.ErrorContains(t, err, "custom.action must be deny or warn")
			})
		}
	}
	for _, helper := range []string{"helper := true", "helper(x) := x", "helper.nested := true"} {
		t.Run(helper, func(t *testing.T) {
			_, err := policyopa.Parse("package nuon.checks\ncheck_example contains {}\n# METADATA\n# custom:\n#   action: warn\n" + helper)
			require.ErrorContains(t, err, "custom.action is only supported on package nuon.checks or a check_* finding set")
		})
	}
}

func TestInvalidFindingsRetainValidEvidence(t *testing.T) {
	policy, err := policyopa.Parse(`package nuon.checks
import data.nuon.lib.check
check_output contains check.finding("blocked")
check_output contains {"kind":"error", "code":"custom"}
check_clean contains check.finding("never") if { false }
`)
	require.NoError(t, err)
	results := policy.Evaluate(context.Background(), []byte(`{}`), 2, "example")
	require.Len(t, results, 2)
	assert.Equal(t, "pass", results[0].Decision)
	assert.Equal(t, "error", results[1].Decision)
	require.Len(t, results[1].Findings, 1)
	assert.Equal(t, "blocked", results[1].Findings[0].Message)
	assert.Equal(t, "invalid_result", results[1].Errors[0].Code)
}

func TestPolicyValidationAndCompatibility(t *testing.T) {
	for name, source := range map[string]string{
		"legacy deny":         "package nuon\ndeny := []",
		"legacy warn":         "package nuon\nwarn contains {\"msg\":\"review\"}",
		"legacy check helper": "package nuon\ncheck_local := true\ndeny contains \"blocked\" if check_local",
		"legacy metadata":     "# METADATA\n# custom:\n#   action: anything\npackage nuon\n# METADATA\n# custom:\n#   action: anything\nhelper := true\ndeny contains \"blocked\" if helper",
	} {
		t.Run(name, func(t *testing.T) {
			policy, err := policyopa.Parse(source)
			require.NoError(t, err)
			assert.Empty(t, policy.Checks)
		})
	}
	for name, source := range map[string]string{
		"wrong package":           "package other\ndeny := []",
		"missing legacy decision": "package nuon\nallow := true",
		"missing checks":          "package nuon.checks\nhelper := true",
		"mixed format":            "package nuon.checks\ndeny := []\ncheck_example contains {}",
		"nested legacy rule":      "package nuon.checks\ndeny.unsafe contains {}\ncheck_example contains {}",
		"complete result":         "package nuon.checks\ncheck_example := []",
		"function":                "package nuon.checks\ncheck_example(x) := x",
		"nested result":           "package nuon.checks\ncheck_example.sub contains {}\ncheck_other contains {}",
		"duplicate metadata":      "package nuon.checks\n# METADATA\n# title: First\ncheck_example contains {}\n# METADATA\n# title: Second\ncheck_example contains {}",
		"undefined helper":        "package nuon.checks\ncheck_example contains nonexistent(input)",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := policyopa.Parse(source)
			require.Error(t, err)
		})
	}
}
