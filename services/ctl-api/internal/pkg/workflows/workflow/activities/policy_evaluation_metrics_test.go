package activities

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.temporal.io/sdk/testsuite"
)

func TestEvaluateSinglePolicyMetrics(t *testing.T) {
	tests := []struct {
		name       string
		policy     string
		input      []byte
		decision   string
		errorStage string
	}{
		{
			name:     "pass",
			policy:   "package nuon\ndeny := []\nwarn := []",
			input:    []byte(`{"identity":"sensitive-input"}`),
			decision: "pass",
		},
		{
			name:     "warn",
			policy:   "package nuon\ndeny := []\nwarn := [\"warning\"]",
			input:    []byte(`{}`),
			decision: "warn",
		},
		{
			name:     "deny takes precedence over warn",
			policy:   "package nuon\ndeny := [\"denied\"]\nwarn := [\"warning\"]",
			input:    []byte(`{}`),
			decision: "deny",
		},
		{
			name:       "evaluation error",
			policy:     "package nuon\ndeny := true\nwarn := []",
			input:      []byte(`{}`),
			errorStage: "deny_evaluation",
		},
		{
			name: "invalid policy", policy: "invalid rego", input: []byte(`{}`), errorStage: "policy_validation",
		},
		{
			name: "invalid input", policy: "package nuon\ndeny := []\nwarn := []", input: []byte(`{`), errorStage: "input_validation",
		},
		{
			name: "warn evaluation fails after deny", policy: "package nuon\ndeny := [\"denied\"]\nwarn := true", input: []byte(`{}`), errorStage: "warn_evaluation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := sdkmetric.NewManualReader()
			provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
			activities := New(Params{MeterProvider: provider})
			req := &EvaluateSinglePolicyRequest{
				PolicyID:      "sensitive-policy-id",
				PolicyName:    "sensitive-policy-name",
				Contents:      tt.policy,
				InputJSON:     tt.input,
				InputIdentity: "sensitive-input-identity",
			}

			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestActivityEnvironment()
			env.RegisterActivity(activities.EvaluateSinglePolicy)
			encoded, err := env.ExecuteActivity(activities.EvaluateSinglePolicy, req)
			if tt.errorStage == "" {
				require.NoError(t, err)
				var result EvaluateSinglePolicyResult
				require.NoError(t, encoded.Get(&result))
			} else {
				require.Error(t, err)
			}

			var resourceMetrics metricdata.ResourceMetrics
			require.NoError(t, reader.Collect(context.Background(), &resourceMetrics))
			metrics := metricDataByName(resourceMetrics)
			require.Contains(t, metrics, "nuon.policy.evaluation.count")
			require.Contains(t, metrics, "nuon.policy.evaluation.duration")

			count := metrics["nuon.policy.evaluation.count"].(metricdata.Sum[int64])
			require.Len(t, count.DataPoints, 1)
			assert.EqualValues(t, 1, count.DataPoints[0].Value)
			attrs := count.DataPoints[0].Attributes
			if tt.errorStage == "" {
				assert.Equal(t, attribute.NewSet(
					attribute.String("outcome", "success"),
					attribute.String("decision", tt.decision),
				), attrs)
			} else {
				assert.Equal(t, attribute.NewSet(
					attribute.String("outcome", "error"),
					attribute.String("error.type", tt.errorStage),
				), attrs)
			}

			histogram := metrics["nuon.policy.evaluation.duration"].(metricdata.Histogram[float64])
			require.Len(t, histogram.DataPoints, 1)
			assert.EqualValues(t, 1, histogram.DataPoints[0].Count)
			assert.Equal(t, attrs, histogram.DataPoints[0].Attributes)
			labels := fmt.Sprint(attrs.ToSlice())
			assert.NotContains(t, labels, req.PolicyID)
			assert.NotContains(t, labels, req.PolicyName)
			assert.NotContains(t, labels, req.InputIdentity)
			assert.NotContains(t, labels, "sensitive-input")
		})
	}
}

func metricDataByName(resourceMetrics metricdata.ResourceMetrics) map[string]metricdata.Aggregation {
	metrics := make(map[string]metricdata.Aggregation)
	for _, scope := range resourceMetrics.ScopeMetrics {
		for _, metric := range scope.Metrics {
			metrics[metric.Name] = metric.Data
		}
	}
	return metrics
}
