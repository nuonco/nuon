package telemetryexport

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	vendorFileStorageExtensionID = "file_storage/vendor"
	vendorBearerAuthExtensionID  = "bearertokenauth/vendor"
	vendorQueueSizeBytes         = 1 << 30
	vendorQueueConsumers         = 2
	maxOTLPRequestBodySize       = 4 << 20
)

// Config controls local listeners and state, not the remotely selected relay destination.
// TokenDirectory must be dedicated to relay credentials; disabling export removes it.
type Config struct {
	StorageDirectory string
	TokenDirectory   string
	GRPCAddress      string
	HTTPAddress      string
	HealthAddress    string
	// CAFile supplements, rather than replaces, the system trust pool
	// (including SSL_CERT_FILE and SSL_CERT_DIR roots) for relay connections.
	CAFile string
}

func DefaultConfig(directory string) Config {
	return Config{
		StorageDirectory: filepath.Join(directory, "vendor"),
		TokenDirectory:   filepath.Join(directory, "vendor-auth"),
		GRPCAddress:      "0.0.0.0:4317",
		HTTPAddress:      "0.0.0.0:4318",
		HealthAddress:    "127.0.0.1:13134",
		CAFile:           os.Getenv("SSL_CERT_FILE"),
	}
}

func validateEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if endpoint == "" || len(endpoint) > 4096 || err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" || u.RawQuery != "" || strings.Contains(endpoint, "${") {
		return fmt.Errorf("telemetry export endpoint must be an HTTPS URL with a host and no userinfo, query, fragment, or environment expansion")
	}
	return nil
}

// CollectorConfig produces the same persistent OTLP pipelines for runner and standalone collectors.
// Literal metadata and local paths are escaped against Collector environment expansion.
func CollectorConfig(cfg Config, endpoint string, attributes map[string]string) ([]byte, error) {
	if err := validateEndpoint(endpoint); err != nil {
		return nil, err
	}
	literal := func(value string) string { return strings.ReplaceAll(value, "$", "$$") }
	receiver := map[string]any{"protocols": map[string]any{
		"grpc": map[string]any{"endpoint": cfg.GRPCAddress, "max_recv_msg_size_mib": maxOTLPRequestBodySize >> 20},
		"http": map[string]any{"endpoint": cfg.HTTPAddress, "max_request_body_size": maxOTLPRequestBodySize},
	}}
	exporter := map[string]any{
		"endpoint": literal(endpoint), "compression": "gzip", "timeout": "30s",
		"auth":             map[string]any{"authenticator": vendorBearerAuthExtensionID},
		"sending_queue":    map[string]any{"enabled": true, "sizer": "bytes", "queue_size": vendorQueueSizeBytes, "num_consumers": vendorQueueConsumers, "storage": vendorFileStorageExtensionID, "block_on_overflow": false},
		"retry_on_failure": map[string]any{"enabled": true, "initial_interval": "1s", "max_interval": "30s", "max_elapsed_time": "0s"},
	}
	if cfg.CAFile != "" {
		// The Collector otherwise trusts only ca_file, dropping system and SSL_CERT_DIR roots.
		exporter["tls"] = map[string]any{"ca_file": literal(cfg.CAFile), "include_system_ca_certs_pool": true}
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
			actions = append(actions, map[string]any{"action": "upsert", "key": literal(key), "value": literal(attributes[key])})
		}
		processors["resource/install"] = map[string]any{"attributes": actions}
		pipeline["processors"] = []string{"memory_limiter", "resource/install"}
	}
	directory := literal(cfg.StorageDirectory)
	document := map[string]any{
		"extensions": map[string]any{
			"health_check": map[string]any{"endpoint": cfg.HealthAddress},
			vendorFileStorageExtensionID: map[string]any{
				"directory": directory, "create_directory": true, "directory_permissions": "0700", "fsync": true,
				"compaction": map[string]any{"on_rebound": true, "directory": directory, "cleanup_on_start": true},
			},
			vendorBearerAuthExtensionID: map[string]any{"filename": literal(filepath.Join(cfg.TokenDirectory, "access-token"))},
		},
		"receivers":  map[string]any{"otlp": receiver},
		"processors": processors,
		"exporters":  map[string]any{"otlp_http/vendor": exporter},
		"service": map[string]any{
			"extensions": []string{"health_check", vendorFileStorageExtensionID, vendorBearerAuthExtensionID},
			"pipelines":  map[string]any{"logs": pipeline, "metrics": pipeline, "traces": pipeline},
			"telemetry":  map[string]any{"logs": map[string]any{"level": "warn"}},
		},
	}
	return yaml.Marshal(document)
}
