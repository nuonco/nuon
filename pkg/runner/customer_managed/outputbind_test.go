package customermanaged

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComponentOutputPlaceholderDeterministicAndCollisionResistant(t *testing.T) {
	first := ComponentOutputPlaceholder("certificate", "public_domain_certificate_arn")
	require.Equal(t, first, ComponentOutputPlaceholder("certificate", "public_domain_certificate_arn"))
	require.Contains(t, first, "__NUON_CUSTOMER_MANAGED_COMPONENT_")

	require.NotEqual(t, ComponentOutputPlaceholder("api", "a.b"), ComponentOutputPlaceholder("api", "a_b"))
	require.NotEqual(t, ComponentOutputPlaceholder("api", "a.b"), ComponentOutputPlaceholder("api-a", "b"))
}

func TestResolveOutputPath(t *testing.T) {
	outputs := map[string]any{
		"lambda_function": map[string]any{"lambda_function_arn": "arn:aws:lambda:us-west-2:123:function:demo"},
		"endpoint":        "https://example.test",
	}

	value, ok := ResolveOutputPath(outputs, "lambda_function.lambda_function_arn")
	require.True(t, ok)
	require.Equal(t, "arn:aws:lambda:us-west-2:123:function:demo", value)

	value, ok = ResolveOutputPath(outputs, "endpoint")
	require.True(t, ok)
	require.Equal(t, "https://example.test", value)

	_, ok = ResolveOutputPath(outputs, "lambda_function.missing")
	require.False(t, ok)
	_, ok = ResolveOutputPath(outputs, "endpoint.nested")
	require.False(t, ok)
}

func TestOutputValueString(t *testing.T) {
	str, err := OutputValueString("plain")
	require.NoError(t, err)
	require.Equal(t, "plain", str)

	str, err = OutputValueString(float64(8080))
	require.NoError(t, err)
	require.Equal(t, "8080", str)

	str, err = OutputValueString(true)
	require.NoError(t, err)
	require.Equal(t, "true", str)

	str, err = OutputValueString(map[string]any{"a": []any{"b"}})
	require.NoError(t, err)
	require.JSONEq(t, `{"a":["b"]}`, str)

	_, err = OutputValueString(nil)
	require.ErrorContains(t, err, "null")
}

func TestSubstituteComponentOutputsReplacesEmbeddedTokens(t *testing.T) {
	token := ComponentOutputPlaceholder("certificate", "arn")
	var plan map[string]any
	require.NoError(t, json.Unmarshal([]byte(`{"vars":{"exact":"`+token+`","embedded":"prefix `+token+` "}}`), &plan))

	SubstituteComponentOutputs(plan, map[string]string{token: "arn:aws:acm:cert"})

	vars := plan["vars"].(map[string]any)
	require.Equal(t, "arn:aws:acm:cert", vars["exact"])
	require.Equal(t, "prefix arn:aws:acm:cert ", vars["embedded"])
}

func TestUnresolvedComponentOutputs(t *testing.T) {
	resolved := ComponentOutputPlaceholder("a", "x")
	unresolved := ComponentOutputPlaceholder("b", "y")
	bindings := []OutputBinding{
		{Token: resolved, ComponentName: "a", StepID: "s", OutputPath: "x"},
		{Token: unresolved, ComponentName: "b", StepID: "s", OutputPath: "y"},
	}
	plan := map[string]any{"value": "left: " + unresolved}

	missing := UnresolvedComponentOutputs(plan, bindings)
	require.Len(t, missing, 1)
	require.Equal(t, "b", missing[0].ComponentName)
}

func TestEnvelopeValidateOutputBindings(t *testing.T) {
	base := func() *Envelope {
		envelope := latebindEnvelope(t)
		envelope.OutputBindings = []OutputBinding{{Token: "__T1__", ComponentName: "a", StepID: "apply", OutputPath: "x"}}
		return envelope
	}
	require.NoError(t, base().Validate())

	envelope := base()
	envelope.OutputBindings[0].OutputPath = ""
	require.ErrorContains(t, envelope.Validate(), "incomplete")

	envelope = base()
	envelope.OutputBindings = append(envelope.OutputBindings, envelope.OutputBindings[0])
	require.ErrorContains(t, envelope.Validate(), "duplicate output binding token")

	envelope = base()
	envelope.OutputBindings[0].StepID = "missing-step"
	require.ErrorContains(t, envelope.Validate(), "unknown step")
}
