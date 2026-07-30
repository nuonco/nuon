package eventfilter

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRestrictedPaths(t *testing.T) {
	for _, path := range []string{`$.items[*].name`, `$['build.version']`, `$.items[0]`} {
		_, err := ParsePath(path, true)
		require.NoError(t, err, path)
	}
	for _, path := range []string{`$..name`, `$.items[0:2]`, `$['a','b']`, `$.items[?@.ok]`, `$.items[?length(@.name)>0]`, `$.items[-1]`} {
		_, err := ParsePath(path, true)
		require.Error(t, err, path)
	}
	_, err := ParsePath(`$.items[*]`, false)
	require.Error(t, err)
}

func TestEvaluateOperatorsAndWildcardCardinality(t *testing.T) {
	payload := map[string]any{"names": []any{"alpha", "release-main"}, "mixed": []any{"other", 1}, "count": json.Number("9007199254740993"), "fraction": json.Number("1.25"), "null": nil}
	tests := []struct {
		filter  Filter
		matched bool
	}{
		{Filter{Path: `$.names[*]`, Op: OperatorSuffix, Value: "main"}, true},
		{Filter{Path: `$.names[*]`, Op: OperatorPrefix, Value: "release-"}, true},
		{Filter{Path: `$.names[*]`, Op: OperatorContains, Value: "ease-m"}, true},
		{Filter{Path: `$.names[*]`, Op: OperatorIn, Value: []any{"beta", "alpha"}}, true},
		{Filter{Path: `$.names[*]`, Op: OperatorNEq, Value: "alpha"}, false},
		{Filter{Path: `$.names[*]`, Op: OperatorNEq, Value: "other"}, true},
		{Filter{Path: `$.mixed[*]`, Op: OperatorNEq, Value: "alpha"}, false},
		{Filter{Path: `$.missing`, Op: OperatorNEq, Value: "other"}, false},
		{Filter{Path: `$.count`, Op: OperatorGT, Value: int64(9007199254740992)}, true},
		{Filter{Path: `$.fraction`, Op: OperatorGTE, Value: 1.25}, true},
		{Filter{Path: `$.fraction`, Op: OperatorLT, Value: 2}, true},
		{Filter{Path: `$.fraction`, Op: OperatorLTE, Value: 1.24}, false},
		{Filter{Path: `$.fraction`, Op: OperatorEq, Value: 1.25}, true},
		{Filter{Path: `$.null`, Op: OperatorExists}, true},
		{Filter{Path: `$.missing`, Op: OperatorNotExists}, true},
		{Filter{Path: `$.names[*]`, Op: OperatorRegex, Value: `^release-`}, true},
	}
	for _, test := range tests {
		compiled, err := Compile(test.filter)
		require.NoError(t, err)
		require.Equal(t, test.matched, compiled.Evaluate(payload, nil).Matched, test.filter)
	}
}

func TestHeaderFilterIsCaseInsensitiveAndMultiValue(t *testing.T) {
	compiled, err := Compile(Filter{From: SourceHeaders, Path: "x-event", Op: OperatorEq, Value: "second"})
	require.NoError(t, err)
	headers := http.Header{}
	headers.Add("X-Event", "first")
	headers.Add("X-Event", "second")
	require.True(t, compiled.Evaluate(nil, headers).Matched)
	neq, err := Compile(Filter{From: SourceHeaders, Path: "x-event", Op: OperatorNEq, Value: "other"})
	require.NoError(t, err)
	headers.Add("X-Event", "other")
	require.False(t, neq.Evaluate(nil, headers).Matched)
}
