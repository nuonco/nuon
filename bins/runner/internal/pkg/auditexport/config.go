package auditexport

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
)

var headerNamePattern = regexp.MustCompile(`^[!#$%&'*+.^_` + "`" + `|~0-9A-Za-z-]+$`)

const (
	maxSecretSize  = 64 * 1024
	maxEndpointLen = 4096
	maxHeaders     = 32
	maxHeaderLen   = 4096
)

type secretConfig struct {
	Exporters struct {
		OTLPHTTP struct {
			Endpoint string            `yaml:"endpoint"`
			Headers  map[string]string `yaml:"headers,omitempty"`
		} `yaml:"otlphttp"`
	} `yaml:"exporters"`
}

func parseSecret(value string) (secretConfig, error) {
	var cfg secretConfig
	if len(value) == 0 || len(value) > maxSecretSize {
		return cfg, fmt.Errorf("audit export configuration has invalid size")
	}
	if decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value)); err == nil {
		value = string(decoded)
	}
	if len(value) == 0 || len(value) > maxSecretSize {
		return cfg, fmt.Errorf("decoded audit export configuration has invalid size")
	}
	decoder := yaml.NewDecoder(strings.NewReader(value))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("decode audit export configuration: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err == nil {
		return cfg, fmt.Errorf("audit export configuration contains trailing document")
	} else if !errors.Is(err, io.EOF) {
		return cfg, fmt.Errorf("decode trailing audit export configuration: %w", err)
	}

	endpoint := cfg.Exporters.OTLPHTTP.Endpoint
	if endpoint == "" || len(endpoint) > maxEndpointLen {
		return cfg, fmt.Errorf("audit export endpoint is required or too long")
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" || u.RawQuery != "" || strings.Contains(endpoint, "${") {
		return cfg, fmt.Errorf("audit export endpoint must be an HTTPS URL with a host and no userinfo, query, fragment, or environment expansion")
	}
	if len(cfg.Exporters.OTLPHTTP.Headers) > maxHeaders {
		return cfg, fmt.Errorf("audit export configuration has too many headers")
	}
	for name, value := range cfg.Exporters.OTLPHTTP.Headers {
		if name == "" || len(name) > 256 || http.CanonicalHeaderKey(name) == "" || !headerNamePattern.MatchString(name) {
			return cfg, fmt.Errorf("audit export configuration has an invalid header name")
		}
		if value == "" || len(value) > maxHeaderLen || strings.ContainsAny(value, "\r\n\x00") || strings.Contains(value, "${") {
			return cfg, fmt.Errorf("audit export configuration has an invalid header value")
		}
	}
	return cfg, nil
}

func collectorConfig(cfg secretConfig) ([]byte, []string, error) {
	headers := make(map[string]string, len(cfg.Exporters.OTLPHTTP.Headers))
	headerNames := make([]string, 0, len(cfg.Exporters.OTLPHTTP.Headers))
	for name := range cfg.Exporters.OTLPHTTP.Headers {
		headerNames = append(headerNames, name)
	}
	sort.Strings(headerNames)

	environment := make([]string, 0, len(headerNames))
	for i, name := range headerNames {
		value := cfg.Exporters.OTLPHTTP.Headers[name]
		envName := fmt.Sprintf("NUON_AUDIT_HEADER_%d", i)
		headers[name] = "${env:" + envName + "}"
		environment = append(environment, envName+"="+value)
	}
	document := map[string]any{
		"receivers": map[string]any{"otlp": map[string]any{"protocols": map[string]any{"http": map[string]any{"endpoint": "127.0.0.1:4318"}}}},
		"processors": map[string]any{
			"memory_limiter": map[string]any{"check_interval": "1s", "limit_mib": 128, "spike_limit_mib": 32},
			"filter/audit": map[string]any{
				"error_mode": "ignore",
				"logs": map[string]any{
					"log_record": []string{`attributes["nuon.audit"] != "true"`},
				},
			},
			"batch": map[string]any{"send_batch_size": 512, "timeout": "5s"},
		},
		"exporters": map[string]any{"otlp_http": map[string]any{
			"endpoint": cfg.Exporters.OTLPHTTP.Endpoint, "headers": headers, "compression": "gzip",
			"sending_queue":    map[string]any{"enabled": true, "queue_size": 1000, "num_consumers": 2},
			"retry_on_failure": map[string]any{"enabled": true, "initial_interval": "1s", "max_interval": "30s", "max_elapsed_time": "5m"},
		}},
		"extensions": map[string]any{"health_check": map[string]any{"endpoint": "127.0.0.1:13133"}},
		"service": map[string]any{
			"extensions": []string{"health_check"},
			"pipelines":  map[string]any{"logs": map[string]any{"receivers": []string{"otlp"}, "processors": []string{"memory_limiter", "filter/audit", "batch"}, "exporters": []string{"otlp_http"}}},
			"telemetry":  map[string]any{"logs": map[string]any{"level": "warn"}},
		},
	}
	contents, err := yaml.Marshal(document)
	return contents, environment, err
}
