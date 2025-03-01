package slog

import (
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
	"github.com/powertoolsdev/mono/pkg/generics"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

func getResource(set *settings.Settings) *resource.Resource {
	attrs := []attribute.KeyValue{}
	builtInAttrs := map[string]string{
		"service.name": "runner",
	}

	for k, v := range generics.MergeMap(set.Metadata, builtInAttrs) {
		attrs = append(attrs, attribute.KeyValue{
			Key:   attribute.Key(k),
			Value: attribute.StringValue(v),
		})
	}

	return resource.NewWithAttributes(set.OtelSchemaURL, attrs...)
}
