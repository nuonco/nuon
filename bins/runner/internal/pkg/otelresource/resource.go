package otelresource

import (
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/runner/settings"
	"github.com/nuonco/nuon/pkg/runner/version"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

func New(set *settings.Settings, logStreamID string) *resource.Resource {
	attrs := []attribute.KeyValue{}
	builtInAttrs := map[string]string{
		string(semconv.ServiceNamespaceKey): "nuon",
		string(semconv.ServiceNameKey):      "runner",
	}
	if version.Version != "" {
		builtInAttrs[string(semconv.ServiceVersionKey)] = version.Version
	}
	if runnerID := set.Metadata["runner.id"]; runnerID != "" {
		builtInAttrs[string(semconv.ServiceInstanceIDKey)] = runnerID
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
