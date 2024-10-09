package slog

import (
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

func getResource(set *settings.Settings) *resource.Resource {
	attrs := []attribute.KeyValue{}
	for k, v := range set.Metadata {
		attrs = append(attrs, attribute.KeyValue{
			Key:   attribute.Key(k),
			Value: attribute.StringValue(v),
		})
	}
	return resource.NewWithAttributes(set.OtelSchemaURL, attrs...)
}
