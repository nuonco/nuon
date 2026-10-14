package telemetryexport

import (
	"encoding/pem"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const testCollectorDirectory = "/var/lib/nuon/telemetry-export"

func defaultTestConfig(t *testing.T) Config {
	t.Helper()
	t.Setenv("SSL_CERT_FILE", "")
	return DefaultConfig(testCollectorDirectory)
}

func TestCollectorConfigBuildsPersistentPipelines(t *testing.T) {
	contents, err := CollectorConfig(defaultTestConfig(t), "https://relay.example.com", nil)
	require.NoError(t, err)
	require.NotContains(t, string(contents), "access_token", "generated collector configuration contains credential material")

	var generated struct {
		Extensions map[string]struct {
			Endpoint  string `yaml:"endpoint"`
			Directory string `yaml:"directory"`
			Filename  string `yaml:"filename"`
		} `yaml:"extensions"`
		Receivers map[string]struct {
			Protocols struct {
				GRPC struct {
					Endpoint          string `yaml:"endpoint"`
					MaxReceiveSizeMiB int    `yaml:"max_recv_msg_size_mib"`
				} `yaml:"grpc"`
				HTTP struct {
					Endpoint           string `yaml:"endpoint"`
					MaxRequestBodySize int    `yaml:"max_request_body_size"`
				} `yaml:"http"`
			} `yaml:"protocols"`
		} `yaml:"receivers"`
		Exporters map[string]struct {
			Endpoint string `yaml:"endpoint"`
			Auth     struct {
				Authenticator string `yaml:"authenticator"`
			} `yaml:"auth"`
			TLS          map[string]any `yaml:"tls"`
			SendingQueue struct {
				Enabled         bool   `yaml:"enabled"`
				Sizer           string `yaml:"sizer"`
				QueueSize       int    `yaml:"queue_size"`
				NumConsumers    int    `yaml:"num_consumers"`
				Storage         string `yaml:"storage"`
				BlockOnOverflow bool   `yaml:"block_on_overflow"`
			} `yaml:"sending_queue"`
			Retry struct {
				Enabled        bool   `yaml:"enabled"`
				MaxElapsedTime string `yaml:"max_elapsed_time"`
			} `yaml:"retry_on_failure"`
		} `yaml:"exporters"`
		Service struct {
			Extensions []string       `yaml:"extensions"`
			Pipelines  map[string]any `yaml:"pipelines"`
		} `yaml:"service"`
	}
	require.NoError(t, yaml.Unmarshal(contents, &generated))

	storage := generated.Extensions["file_storage/vendor"]
	require.Equal(t, testCollectorDirectory+"/vendor", storage.Directory, "vendor storage is not isolated")
	require.Equal(t, "127.0.0.1:13134", generated.Extensions["health_check"].Endpoint)
	require.Equal(t, testCollectorDirectory+"/vendor-auth/access-token", generated.Extensions["bearertokenauth/vendor"].Filename)
	require.Subset(t, generated.Service.Extensions, []string{"health_check", "file_storage/vendor", "bearertokenauth/vendor"})

	receiver := generated.Receivers["otlp"]
	require.Equal(t, "0.0.0.0:4317", receiver.Protocols.GRPC.Endpoint)
	require.Equal(t, 4, receiver.Protocols.GRPC.MaxReceiveSizeMiB)
	require.Equal(t, "0.0.0.0:4318", receiver.Protocols.HTTP.Endpoint)
	require.Equal(t, 4<<20, receiver.Protocols.HTTP.MaxRequestBodySize)

	exporter := generated.Exporters["otlp_http/vendor"]
	require.Equal(t, "https://relay.example.com", exporter.Endpoint)
	require.Equal(t, "bearertokenauth/vendor", exporter.Auth.Authenticator)
	require.Nil(t, exporter.TLS, "exporter overrides TLS trust without a custom CA")
	queue := exporter.SendingQueue
	if !queue.Enabled || queue.Sizer != "bytes" || queue.QueueSize != 1<<30 || queue.NumConsumers != 2 || queue.Storage != "file_storage/vendor" || queue.BlockOnOverflow || !exporter.Retry.Enabled || exporter.Retry.MaxElapsedTime != "0s" {
		t.Fatalf("vendor exporter is not durable and asynchronous: %#v", exporter)
	}
	for _, pipeline := range []string{"logs", "metrics", "traces"} {
		require.Contains(t, generated.Service.Pipelines, pipeline)
	}
	for _, unexpected := range []string{"logs/audit", "otlp_http/async", "otlp_http/sync", "file_storage/audit", "${env:"} {
		require.NotContains(t, string(contents), unexpected)
	}
}

func TestCollectorConfigCAFileSupplementsSystemTrust(t *testing.T) {
	t.Setenv("SSL_CERT_FILE", "/etc/custom-$ca/bundle.pem")
	cfg := DefaultConfig(testCollectorDirectory)
	require.Equal(t, "/etc/custom-$ca/bundle.pem", cfg.CAFile)
	contents, err := CollectorConfig(cfg, "https://relay.example.com", nil)
	require.NoError(t, err)
	var generated struct {
		Exporters map[string]struct {
			TLS map[string]any `yaml:"tls"`
		} `yaml:"exporters"`
	}
	require.NoError(t, yaml.Unmarshal(contents, &generated))
	require.Equal(t, map[string]any{
		"ca_file":                      "/etc/custom-$$ca/bundle.pem",
		"include_system_ca_certs_pool": true,
	}, generated.Exporters["otlp_http/vendor"].TLS)
}

func TestCollectorConfigResourceActions(t *testing.T) {
	cfg := defaultTestConfig(t)
	for _, attributes := range []map[string]string{nil, {}, {
		"nuon.install.name":                        "production-eu",
		"nuon.install.labels.literal-${env:LABEL}": "${env:VALUE}\n$ $$",
	}} {
		contents, err := CollectorConfig(cfg, "https://relay.example.com", attributes)
		require.NoError(t, err)
		var generated struct {
			Processors map[string]struct {
				Attributes []map[string]any `yaml:"attributes"`
			} `yaml:"processors"`
			Service struct {
				Pipelines map[string]struct {
					Processors []string `yaml:"processors"`
				} `yaml:"pipelines"`
			} `yaml:"service"`
		}
		require.NoError(t, yaml.Unmarshal(contents, &generated))
		wantProcessors := []string{"memory_limiter"}
		if len(attributes) == 0 {
			require.NotContains(t, generated.Processors, "resource/install")
		} else {
			wantProcessors = append(wantProcessors, "resource/install")
			require.Equal(t, []map[string]any{
				{"action": "delete", "pattern": `^nuon\.install\.labels\.`},
				{"action": "upsert", "key": "nuon.install.labels.literal-$${env:LABEL}", "value": "$${env:VALUE}\n$$ $$$$"},
				{"action": "upsert", "key": "nuon.install.name", "value": "production-eu"},
			}, generated.Processors["resource/install"].Attributes)
		}
		for _, signal := range []string{"logs", "metrics", "traces"} {
			require.Equal(t, wantProcessors, generated.Service.Pipelines[signal].Processors, signal)
		}
	}
}

func TestCollectorConfigRejectsInvalidEndpoint(t *testing.T) {
	cfg := defaultTestConfig(t)
	for _, endpoint := range []string{"", "http://relay.example.com", "https://user@relay.example.com", "https://relay.example.com/?q=1", "https://relay.example.com/#f", "https://${env:HOST}"} {
		_, err := CollectorConfig(cfg, endpoint, nil)
		require.Error(t, err, endpoint)
	}
}

// NUON_TEST_OTELCOL is the pinned telemetry-agent Collector (see the Dockerfile test target).
func TestCollectorConfigValidatesWithCollector(t *testing.T) {
	binary := os.Getenv("NUON_TEST_OTELCOL")
	if binary == "" {
		t.Skip("set NUON_TEST_OTELCOL to the built telemetry-agent Collector binary")
	}
	directory := t.TempDir()
	server := httptest.NewTLSServer(nil)
	caFile := filepath.Join(directory, "ca.pem")
	require.NoError(t, os.WriteFile(caFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0o600))
	server.Close()
	for name, ca := range map[string]string{"system trust": "", "supplemental ca": caFile} {
		t.Run(name, func(t *testing.T) {
			cfg := DefaultConfig(filepath.Join(directory, strings.ReplaceAll(name, " ", "-")))
			cfg.CAFile = ca
			contents, err := CollectorConfig(cfg, "https://relay.example.com", map[string]string{
				"nuon.org.name": "acme", "nuon.app.name": "payments", "nuon.install.name": "production-eu",
				"nuon.install.labels.tier": "enterprise",
			})
			require.NoError(t, err)
			require.Equal(t, ca != "", slices.Contains(strings.Fields(string(contents)), "include_system_ca_certs_pool:"))
			path := filepath.Join(t.TempDir(), "collector.yaml")
			require.NoError(t, os.WriteFile(path, contents, 0o600))
			if output, err := exec.Command(binary, "validate", "--config", path).CombinedOutput(); err != nil {
				t.Fatalf("collector configuration is invalid: %v\n%s", err, output)
			}
		})
	}
}
