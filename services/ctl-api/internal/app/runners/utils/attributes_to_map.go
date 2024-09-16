package utils

import (
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
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
