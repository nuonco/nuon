package telemetryexport

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/audit"
)

var headerNamePattern = regexp.MustCompile(`^[!#$%&'*+.^_` + "`" + `|~0-9A-Za-z-]+$`)

const (
	configVersionV1              = "v1"
	auditFileStorageExtensionID  = "file_storage/audit"
	vendorFileStorageExtensionID = "file_storage/vendor"
	vendorBearerAuthExtensionID  = "bearertokenauth/vendor"
	collectorStorageDir          = "/var/lib/nuon/telemetry-export"
	vendorStorageDir             = collectorStorageDir + "/vendor"
	vendorTokenDir               = collectorStorageDir + "/vendor-auth"
	vendorTokenPath              = vendorTokenDir + "/access-token"
	vendorOTLPGRPCAddress        = "0.0.0.0:4317"
	vendorOTLPHTTPAddress        = "0.0.0.0:4318"
	vendorCollectorHealthAddress = "127.0.0.1:13134"
	auditQueueSize               = 10_000
	auditQueueConsumers          = 2
	vendorQueueSizeBytes         = 1 << 30
	vendorQueueConsumers         = 2
	maxOTLPRequestBodySize       = 4 << 20
	maxSecretSize                = 64 * 1024
	maxEndpointLen               = 4096
	maxHeaders                   = 32
	maxHeaderLen                 = 4096
)

type configEnvelope struct {
	Version string `yaml:"version"`
}

type configV1 struct {
	Version   string            `yaml:"version"`
	Telemetry telemetryConfigV1 `yaml:"telemetry,omitempty"`
	Exporters struct {
		OTLPHTTP otlpHTTPExporter `yaml:"otlphttp,omitempty"`
	} `yaml:"exporters"`
}

type telemetryConfigV1 struct {
	Logs struct {
		Audit struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"audit,omitempty"`
	} `yaml:"logs,omitempty"`
}

type config struct {
	AuditLogsEnabled bool
	OTLPHTTP         otlpHTTPExporter
}

type otlpHTTPExporter struct {
	Endpoint string            `yaml:"endpoint"`
	Headers  map[string]string `yaml:"headers,omitempty"`
}

func parseSecret(value string) (config, error) {
	var cfg config
	if len(value) == 0 || len(value) > maxSecretSize {
		return cfg, fmt.Errorf("telemetry export configuration has invalid size")
	}
	if decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value)); err == nil {
		value = string(decoded)
	}
	if len(value) == 0 || len(value) > maxSecretSize {
		return cfg, fmt.Errorf("decoded telemetry export configuration has invalid size")
	}

	var envelope configEnvelope
	if err := decodeConfigDocument(value, &envelope, false); err != nil {
		return cfg, fmt.Errorf("decode telemetry export configuration version: %w", err)
	}
	switch envelope.Version {
	case "":
		return cfg, fmt.Errorf("telemetry export configuration version is required")
	case configVersionV1:
		return parseConfigV1(value)
	default:
		return cfg, fmt.Errorf("unsupported telemetry export configuration version %q", envelope.Version)
	}
}

func parseConfigV1(value string) (config, error) {
	var wire configV1
	if err := decodeConfigDocument(value, &wire, true); err != nil {
		return config{}, fmt.Errorf("decode telemetry export configuration: %w", err)
	}
	cfg := config{
		AuditLogsEnabled: wire.Telemetry.Logs.Audit.Enabled,
		OTLPHTTP:         wire.Exporters.OTLPHTTP,
	}
	if cfg.AuditLogsEnabled {
		if err := validateOTLPHTTPExporter(cfg.OTLPHTTP); err != nil {
			return config{}, err
		}
	}
	return cfg, nil
}

func decodeConfigDocument(value string, destination any, knownFields bool) error {
	decoder := yaml.NewDecoder(strings.NewReader(value))
	decoder.KnownFields(knownFields)
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err == nil {
		return fmt.Errorf("telemetry export configuration contains trailing document")
	} else if !errors.Is(err, io.EOF) {
		return fmt.Errorf("decode trailing telemetry export configuration: %w", err)
	}
	return nil
}

func validateOTLPHTTPExporter(exporter otlpHTTPExporter) error {
	endpoint := exporter.Endpoint
	if endpoint == "" || len(endpoint) > maxEndpointLen {
		return fmt.Errorf("telemetry export endpoint is required or too long")
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" || u.RawQuery != "" || strings.Contains(endpoint, "${") {
		return fmt.Errorf("telemetry export endpoint must be an HTTPS URL with a host and no userinfo, query, fragment, or environment expansion")
	}
	if len(exporter.Headers) > maxHeaders {
		return fmt.Errorf("telemetry export configuration has too many headers")
	}
	for name, value := range exporter.Headers {
		if name == "" || len(name) > 256 || http.CanonicalHeaderKey(name) == "" || !headerNamePattern.MatchString(name) {
			return fmt.Errorf("telemetry export configuration has an invalid header name")
		}
		if value == "" || len(value) > maxHeaderLen || strings.ContainsAny(value, "\r\n\x00") || strings.Contains(value, "${") {
			return fmt.Errorf("telemetry export configuration has an invalid header value")
		}
	}
	return nil
}

func collectorConfig(cfg config) ([]byte, []string, error) {
	pipelines := make(map[string]any)
	extensions := map[string]any{"health_check": map[string]any{"endpoint": "127.0.0.1:13133"}}
	receivers := make(map[string]any)
	exporters := make(map[string]any)
	serviceExtensions := []string{"health_check"}
	service := map[string]any{
		"extensions": serviceExtensions,
		"pipelines":  pipelines,
		// TODO: Export Collector queue and delivery metrics through the operator telemetry pipeline for alerting.
		"telemetry": map[string]any{"logs": map[string]any{"level": "warn"}},
	}
	document := map[string]any{
		"extensions": extensions,
		"service":    service,
	}
	environment := make([]string, 0, len(cfg.OTLPHTTP.Headers))

	if cfg.AuditLogsEnabled {
		headers, values := headerEnvironment(cfg.OTLPHTTP.Headers, "NUON_TELEMETRY_EXPORT_HEADER")
		environment = append(environment, values...)
		extensions[auditFileStorageExtensionID] = fileStorageConfig(collectorStorageDir)
		serviceExtensions = append(serviceExtensions, auditFileStorageExtensionID)
		receivers["otlp/async"] = map[string]any{"protocols": map[string]any{"http": map[string]any{"endpoint": audit.AsyncRouteAddress}}}
		receivers["otlp/sync"] = map[string]any{"protocols": map[string]any{"http": map[string]any{"endpoint": audit.SyncRouteAddress}}}
		exporters["otlp_http/async"] = map[string]any{
			"endpoint": cfg.OTLPHTTP.Endpoint, "headers": headers, "compression": "gzip",
			"sending_queue":    map[string]any{"enabled": true, "queue_size": auditQueueSize, "num_consumers": auditQueueConsumers, "storage": auditFileStorageExtensionID, "block_on_overflow": false},
			"retry_on_failure": map[string]any{"enabled": true, "initial_interval": "1s", "max_interval": "30s", "max_elapsed_time": "0s"},
		}
		exporters["otlp_http/sync"] = map[string]any{
			"endpoint": cfg.OTLPHTTP.Endpoint, "headers": headers, "compression": "gzip", "timeout": audit.SyncExportTimeout.String(),
			"sending_queue":    map[string]any{"enabled": false},
			"retry_on_failure": map[string]any{"enabled": false},
		}
		pipelines["logs/audit_async"] = map[string]any{"receivers": []string{"otlp/async"}, "processors": []string{"memory_limiter"}, "exporters": []string{"otlp_http/async"}}
		pipelines["logs/audit_sync"] = map[string]any{"receivers": []string{"otlp/sync"}, "processors": []string{"memory_limiter"}, "exporters": []string{"otlp_http/sync"}}
	}

	service["extensions"] = serviceExtensions
	if len(pipelines) != 0 {
		document["receivers"] = receivers
		document["processors"] = map[string]any{"memory_limiter": map[string]any{"check_interval": "1s", "limit_mib": 128, "spike_limit_mib": 32}}
		document["exporters"] = exporters
	}

	contents, err := yaml.Marshal(document)
	return contents, environment, err
}

func vendorCollectorConfig(endpoint string, attributes map[string]string) ([]byte, error) {
	if err := validateOTLPHTTPExporter(otlpHTTPExporter{Endpoint: endpoint}); err != nil {
		return nil, fmt.Errorf("invalid vendor OTLP/HTTP exporter: %w", err)
	}

	receiver := map[string]any{"protocols": map[string]any{
		"grpc": map[string]any{"endpoint": vendorOTLPGRPCAddress, "max_recv_msg_size_mib": maxOTLPRequestBodySize >> 20},
		"http": map[string]any{"endpoint": vendorOTLPHTTPAddress, "max_request_body_size": maxOTLPRequestBodySize},
	}}
	exporter := map[string]any{
		"endpoint": endpoint, "compression": "gzip", "timeout": "30s",
		"auth":             map[string]any{"authenticator": vendorBearerAuthExtensionID},
		"sending_queue":    map[string]any{"enabled": true, "sizer": "bytes", "queue_size": vendorQueueSizeBytes, "num_consumers": vendorQueueConsumers, "storage": vendorFileStorageExtensionID, "block_on_overflow": false},
		"retry_on_failure": map[string]any{"enabled": true, "initial_interval": "1s", "max_interval": "30s", "max_elapsed_time": "0s"},
	}
	pipeline := map[string]any{"receivers": []string{"otlp"}, "processors": []string{"memory_limiter"}, "exporters": []string{"otlp_http/vendor"}}
	processors := map[string]any{"memory_limiter": map[string]any{"check_interval": "1s", "limit_mib": 128, "spike_limit_mib": 32}}
	// Older APIs omit the snapshot; preserve their existing passthrough behavior.
	if len(attributes) > 0 {
		keys := make([]string, 0, len(attributes))
		for key := range attributes {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		actions := []map[string]any{{"action": "delete", "pattern": `^nuon\.install\.labels\.`}}
		for _, key := range keys {
			// Confmap expands environment references even in quoted YAML strings.
			actions = append(actions, map[string]any{
				"action": "upsert",
				"key":    strings.ReplaceAll(key, "$", "$$"),
				"value":  strings.ReplaceAll(attributes[key], "$", "$$"),
			})
		}
		processors["resource/install"] = map[string]any{"attributes": actions}
		pipeline["processors"] = []string{"memory_limiter", "resource/install"}
	}
	document := map[string]any{
		"extensions": map[string]any{
			"health_check":               map[string]any{"endpoint": vendorCollectorHealthAddress},
			vendorFileStorageExtensionID: fileStorageConfig(vendorStorageDir),
			vendorBearerAuthExtensionID:  map[string]any{"filename": vendorTokenPath},
		},
		"receivers":  map[string]any{"otlp": receiver},
		"processors": processors,
		"exporters":  map[string]any{"otlp_http/vendor": exporter},
		"service": map[string]any{
			"extensions": []string{"health_check", vendorFileStorageExtensionID, vendorBearerAuthExtensionID},
			"pipelines": map[string]any{
				"logs":    pipeline,
				"metrics": pipeline,
				"traces":  pipeline,
			},
			"telemetry": map[string]any{"logs": map[string]any{"level": "warn"}},
		},
	}

	return yaml.Marshal(document)
}

func headerEnvironment(values map[string]string, prefix string) (map[string]string, []string) {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)

	headers := make(map[string]string, len(names))
	environment := make([]string, 0, len(names))
	for i, name := range names {
		envName := fmt.Sprintf("%s_%d", prefix, i)
		headers[name] = "${env:" + envName + "}"
		environment = append(environment, envName+"="+values[name])
	}
	return headers, environment
}

func fileStorageConfig(directory string) map[string]any {
	return map[string]any{
		"directory":             directory,
		"create_directory":      true,
		"directory_permissions": "0700",
		"fsync":                 true,
		"compaction": map[string]any{
			"on_rebound":       true,
			"directory":        directory,
			"cleanup_on_start": true,
		},
	}
}
