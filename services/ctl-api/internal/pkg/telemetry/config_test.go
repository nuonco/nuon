package telemetry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/nuonco/nuon/pkg/services/config"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/stretchr/testify/require"
)

func TestConfigValidation(t *testing.T) {
	for _, tc := range []struct {
		name, endpoint, protocol, wantError string
	}{
		{name: "unconfigured"},
		{name: "http", endpoint: "http://localhost:4318"},
		{name: "https with base path", endpoint: "https://example.com/otel/", protocol: "http/protobuf"},
		{name: "missing scheme", endpoint: "localhost:4318", wantError: "OTLP endpoint"},
		{name: "missing host", endpoint: "https:///otel", wantError: "OTLP endpoint"},
		{name: "invalid URL", endpoint: "http://%", wantError: "OTLP endpoint"},
		{name: "credentials", endpoint: "https://user:secret@example.com", wantError: "OTLP endpoint"},
		{name: "query", endpoint: "https://example.com?token=secret", wantError: "OTLP endpoint"},
		{name: "fragment", endpoint: "https://example.com#metrics", wantError: "OTLP endpoint"},
		{name: "unsupported protocol", endpoint: "http://localhost:4318", protocol: "grpc", wantError: "http/protobuf"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := NewConfig(&internal.Config{
				OTELExporterOTLPEndpoint: tc.endpoint,
				OTELExporterOTLPProtocol: tc.protocol,
			})
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				require.Nil(t, cfg)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.endpoint, cfg.Endpoint)
		})
	}
}

func TestSharedResourceDefaults(t *testing.T) {
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "")
	t.Setenv("OTEL_SERVICE_NAME", "")
	serviceConfig := &internal.Config{
		ServiceName:       "ctl-api",
		Version:           "test-version",
		ServiceType:       "api",
		ServiceDeployment: "runner",
	}
	first, err := NewConfig(serviceConfig)
	require.NoError(t, err)
	attrs := map[string]string{}
	for _, attr := range first.Resource.Attributes() {
		attrs[string(attr.Key)] = attr.Value.AsString()
	}
	require.Equal(t, "ctl-api", attrs["service.name"])
	require.Equal(t, "test-version", attrs["service.version"])
	require.Equal(t, "api", attrs["nuon.service.type"])
	require.Equal(t, "runner", attrs["nuon.service.deployment"])
	_, err = uuid.Parse(attrs["service.instance.id"])
	require.NoError(t, err)
	second, err := NewConfig(serviceConfig)
	require.NoError(t, err)
	require.False(t, first.Resource.Equal(second.Resource))
}

func TestSharedConfigFileAndEnvironment(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "")
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("otel_exporter_otlp_endpoint: https://file.example.com/otel\notel_exporter_otlp_protocol: http/protobuf\n"), 0600))
	var serviceConfig internal.Config
	require.NoError(t, config.NewFileLoader(path).LoadInto(nil, &serviceConfig))
	cfg, err := NewConfig(&serviceConfig)
	require.NoError(t, err)
	require.Equal(t, "https://file.example.com/otel", cfg.Endpoint)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://env.example.com/otel")
	require.NoError(t, config.NewFileLoader(path).LoadInto(nil, &serviceConfig))
	cfg, err = NewConfig(&serviceConfig)
	require.NoError(t, err)
	require.Equal(t, "https://env.example.com/otel", cfg.Endpoint)
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	require.NoError(t, config.NewFileLoader(path).LoadInto(nil, &serviceConfig))
	_, err = NewConfig(&serviceConfig)
	require.ErrorContains(t, err, "http/protobuf")
}

func TestRejectSignalSpecificTransport(t *testing.T) {
	for _, suffix := range []string{"ENDPOINT", "PROTOCOL", "HEADERS", "CERTIFICATE", "CLIENT_CERTIFICATE", "CLIENT_KEY", "TIMEOUT", "COMPRESSION", "INSECURE"} {
		t.Run(suffix, func(t *testing.T) {
			key := "OTEL_EXPORTER_OTLP_METRICS_" + suffix
			t.Setenv(key, "secret-value")
			_, err := NewConfig(&internal.Config{OTELExporterOTLPEndpoint: "https://example.com"})
			require.ErrorContains(t, err, key)
			require.NotContains(t, err.Error(), "secret-value")
			_, err = NewConfig(&internal.Config{})
			require.NoError(t, err)
		})
	}
}

func TestInvalidGenericTransportFailsClosed(t *testing.T) {
	for _, tc := range []struct{ key, value string }{
		{"HEADERS", "secret-value"},
		{"HEADERS", "bad key=secret-value"},
		{"HEADERS", "Authorization=%secret-value"},
		{"HEADERS", "Authorization=secret-value%0d%0aInjected:value"},
		{"CERTIFICATE", "/nonexistent/secret-value"},
		{"CLIENT_CERTIFICATE", "/nonexistent/secret-value"},
		{"CLIENT_KEY", "/nonexistent/secret-value"},
		{"TIMEOUT", "secret-value"},
		{"TIMEOUT", "0"},
		{"TIMEOUT", "-1"},
		{"TIMEOUT", "9223372036854775807"},
		{"COMPRESSION", "secret-value"},
		{"INSECURE", "secret-value"},
		{"INSECURE", "true"},
	} {
		t.Run(tc.key+"/"+tc.value, func(t *testing.T) {
			key := "OTEL_EXPORTER_OTLP_" + tc.key
			t.Setenv(key, tc.value)
			_, err := NewConfig(&internal.Config{OTELExporterOTLPEndpoint: "https://example.com"})
			require.ErrorContains(t, err, "OTEL_EXPORTER_OTLP_")
			require.NotContains(t, err.Error(), "secret-value")
			_, err = NewConfig(&internal.Config{})
			require.NoError(t, err)
		})
	}
	path := filepath.Join(t.TempDir(), "invalid.pem")
	require.NoError(t, os.WriteFile(path, []byte("secret-value"), 0600))
	t.Setenv("OTEL_EXPORTER_OTLP_CERTIFICATE", path)
	_, err := NewConfig(&internal.Config{OTELExporterOTLPEndpoint: "https://example.com"})
	require.ErrorContains(t, err, "OTEL_EXPORTER_OTLP_CERTIFICATE")
	require.NotContains(t, err.Error(), path)
}
