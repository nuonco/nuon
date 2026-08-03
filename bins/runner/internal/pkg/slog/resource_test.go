package slog

import (
	"testing"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"

	"github.com/nuonco/nuon/pkg/runner/settings"
	"github.com/nuonco/nuon/pkg/runner/version"
)

func TestResourceIncludesNuonServiceIdentity(t *testing.T) {
	set := &settings.Settings{Metadata: map[string]string{
		"runner.id":         "run123",
		"service.name":      "overridden",
		"service.namespace": "overridden",
	}}
	rsrc := getResource(set, "log123")

	want := map[attribute.Key]string{
		semconv.ServiceNamespaceKey:    "nuon",
		semconv.ServiceNameKey:         "runner",
		semconv.ServiceVersionKey:      version.Version,
		semconv.ServiceInstanceIDKey:   "run123",
		attribute.Key("log_stream.id"): "log123",
	}
	for key, value := range want {
		got, ok := rsrc.Set().Value(key)
		if !ok || got.AsString() != value {
			t.Errorf("%s = %q, want %q", key, got.AsString(), value)
		}
	}
}

func TestResourceOmitsEmptyServiceInstanceID(t *testing.T) {
	rsrc := getResource(&settings.Settings{Metadata: map[string]string{}}, "")
	if _, ok := rsrc.Set().Value(semconv.ServiceInstanceIDKey); ok {
		t.Fatal("empty service.instance.id was included")
	}
}
