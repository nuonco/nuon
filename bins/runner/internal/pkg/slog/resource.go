package slog

import (
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/runner/settings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

func getResource(set *settings.Settings, logStreamID string) *resource.Resource {
	attrs := []attribute.KeyValue{}
	builtInAttrs := map[string]string{
		"service.name": "runner",
	}
	if logStreamID != "" {
		builtInAttrs["log_stream.id"] = logStreamID
	}

	for k, v := range generics.MergeMap(set.Metadata, builtInAttrs) {
		attrs = append(attrs, attribute.KeyValue{
			Key:   attribute.Key(k),
			Value: attribute.StringValue(v),
		})
	}

	return resource.NewWithAttributes(set.OtelSchemaURL, attrs...)
}
