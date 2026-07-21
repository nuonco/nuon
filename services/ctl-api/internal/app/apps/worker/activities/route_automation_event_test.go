package activities

import (
	"encoding/json"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestRuleMatchesAutomationEvent(t *testing.T) {
	payload, err := decodeAutomationPayload(json.RawMessage(`{"image":{"repository":"acme/api","tag":"v1"},"attempt":2}`))
	require.NoError(t, err)
	rule := &app.EventAutomationRule{
		EventTypes: pq.StringArray{"com.acme.image.pushed.v1"},
		Filters: []app.EventAutomationFilter{
			{Op: app.EventAutomationFilterTypeEq, Path: "/image/repository", Value: "acme/api"},
			{Op: app.EventAutomationFilterTypeEq, Path: "/attempt", Value: 2},
		},
	}

	require.True(t, ruleMatchesEvent(rule, "com.acme.image.pushed.v1", payload))
	require.False(t, ruleMatchesEvent(rule, "com.acme.image.deleted.v1", payload))
	rule.Filters[0].Value = "acme/other"
	require.False(t, ruleMatchesEvent(rule, "com.acme.image.pushed.v1", payload))
}

func TestResolveJSONPointerEscapes(t *testing.T) {
	payload := map[string]any{"a/b": []any{map[string]any{"~key": true}}}
	value, ok := resolveJSONPointer(payload, "/a~1b/0/~0key")
	require.True(t, ok)
	require.Equal(t, true, value)
}

func TestResolveJSONPointerRejectsNonCanonicalArrayIndexes(t *testing.T) {
	payload := []any{"first", "second"}
	for _, pointer := range []string{"/01", "/+1", "/-0", "/-"} {
		_, ok := resolveJSONPointer(payload, pointer)
		require.False(t, ok, pointer)
	}
	value, ok := resolveJSONPointer(payload, "/1")
	require.True(t, ok)
	require.Equal(t, "second", value)
}

func TestJSONValuesEqualNumbers(t *testing.T) {
	require.True(t, jsonValuesEqual(json.Number("1.0"), 1))
	require.True(t, jsonValuesEqual(json.Number("9007199254740993"), json.Number("9007199254740993.0")))
	require.False(t, jsonValuesEqual(json.Number("9007199254740993"), json.Number("9007199254740992")))
}
