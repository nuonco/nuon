package utils

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
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
