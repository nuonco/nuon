package log

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

func getResource(logStreamID string, kvs map[string]string) *resource.Resource {
	attrs := []attribute.KeyValue{}
	for k, v := range kvs {
		attrs = append(attrs, attribute.KeyValue{
			Key:   attribute.Key(k),
			Value: attribute.StringValue(v),
		})
	}
	attrs = append(attrs,
		attribute.KeyValue{
			Key:   "log_stream.id",
			Value: attribute.StringValue(logStreamID),
		},
		attribute.KeyValue{
			Key:   "service.name",
			Value: attribute.StringValue("api"),
		},
	)

	return resource.NewSchemaless(attrs...)
}
