package otel

import (
	"encoding/hex"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// src: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/exporter_logs.go#L132-L139
func AttributesToMap(attributes pcommon.Map) map[string]string {
	m := make(map[string]string, attributes.Len())
	attributes.Range(func(k string, v pcommon.Value) bool {
		m[k] = v.AsString()
		return true
	})
	return m
}

// src: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/exporter_traces.go#L130
func ConvertEvents(events ptrace.SpanEventSlice) ([]time.Time, []string, []map[string]string) {
	var (
		times []time.Time
		names []string
		attrs []map[string]string
	)
	for i := 0; i < events.Len(); i++ {
		event := events.At(i)
		times = append(times, event.Timestamp().AsTime())
		names = append(names, event.Name())
		attrs = append(attrs, AttributesToMap(event.Attributes()))
	}
	return times, names, attrs
}

// src: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/exporter_traces.go#L145
func ConvertLinks(links ptrace.SpanLinkSlice) ([]string, []string, []string, []map[string]string) {
	var (
		traceIDs []string
		spanIDs  []string
		states   []string
		attrs    []map[string]string
	)
	for i := 0; i < links.Len(); i++ {
		link := links.At(i)
		traceIDs = append(traceIDs, TraceIDToHexOrEmptyString(link.TraceID()))
		spanIDs = append(spanIDs, SpanIDToHexOrEmptyString(link.SpanID()))
		states = append(states, link.TraceState().AsRaw())
		attrs = append(attrs, AttributesToMap(link.Attributes()))
	}
	return traceIDs, spanIDs, states, attrs
}

// src: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/internal/metrics_model.go#L105C1-L124C2
func ConvertExemplars(exemplars pmetric.ExemplarSlice) (clickhouse.ArraySet, clickhouse.ArraySet, clickhouse.ArraySet, clickhouse.ArraySet, clickhouse.ArraySet) {
	var (
		attrs    clickhouse.ArraySet
		times    clickhouse.ArraySet
		values   clickhouse.ArraySet
		traceIDs clickhouse.ArraySet
		spanIDs  clickhouse.ArraySet
	)
	for i := 0; i < exemplars.Len(); i++ {
		exemplar := exemplars.At(i)
		attrs = append(attrs, AttributesToMap(exemplar.FilteredAttributes()))
		times = append(times, exemplar.Timestamp().AsTime())
		values = append(values, GetValue(exemplar.IntValue(), exemplar.DoubleValue(), exemplar.ValueType()))

		traceID, spanID := exemplar.TraceID(), exemplar.SpanID()
		traceIDs = append(traceIDs, hex.EncodeToString(traceID[:]))
		spanIDs = append(spanIDs, hex.EncodeToString(spanID[:]))
	}
	return attrs, times, values, traceIDs, spanIDs
}

// src: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/internal/metrics_model.go#L171
func ConvertSliceToArraySet[T any](slice []T) clickhouse.ArraySet {
	var set clickhouse.ArraySet
	for _, item := range slice {
		set = append(set, item)
	}
	return set
}

// src: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/internal/metrics_model.go#L179
func ConvertValueAtQuantile(valueAtQuantile pmetric.SummaryDataPointValueAtQuantileSlice) (clickhouse.ArraySet, clickhouse.ArraySet) {
	var (
		quantiles clickhouse.ArraySet
		values    clickhouse.ArraySet
	)
	for i := 0; i < valueAtQuantile.Len(); i++ {
		value := valueAtQuantile.At(i)
		quantiles = append(quantiles, value.Quantile())
		values = append(values, value.Value())
	}
	return quantiles, values
}

// SpanKindStr returns a string representation of the SpanKind as it's defined in the proto.
// The function provides old behavior of ptrace.SpanKind.String() to support graceful adoption of
// https://github.com/open-telemetry/opentelemetry-collector/pull/6250.
func SpanKindStr(sk ptrace.SpanKind) string {
	switch sk {
	case ptrace.SpanKindUnspecified:
		return "SPAN_KIND_UNSPECIFIED"
	case ptrace.SpanKindInternal:
		return "SPAN_KIND_INTERNAL"
	case ptrace.SpanKindServer:
		return "SPAN_KIND_SERVER"
	case ptrace.SpanKindClient:
		return "SPAN_KIND_CLIENT"
	case ptrace.SpanKindProducer:
		return "SPAN_KIND_PRODUCER"
	case ptrace.SpanKindConsumer:
		return "SPAN_KIND_CONSUMER"
	}
	return ""
}

// StatusCodeStr returns a string representation of the StatusCode as it's defined in the proto.
// The function provides old behavior of ptrace.StatusCode.String() to support graceful adoption of
// https://github.com/open-telemetry/opentelemetry-collector/pull/6250.
func StatusCodeStr(sk ptrace.StatusCode) string {
	switch sk {
	case ptrace.StatusCodeUnset:
		return "STATUS_CODE_UNSET"
	case ptrace.StatusCodeOk:
		return "STATUS_CODE_OK"
	case ptrace.StatusCodeError:
		return "STATUS_CODE_ERROR"
	}
	return ""
}
