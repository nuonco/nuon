package telemetryexport

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func auditEnabledConfig(endpoint string) string {
	return fmt.Sprintf(`version: v1
telemetry:
  logs:
    audit:
      enabled: true
exporters:
  otlphttp:
    endpoint: %s
`, endpoint)
}

func TestParseSecretV1(t *testing.T) {
	valid := auditEnabledConfig("https://otlp.example.com") + "    headers:\n      Authorization: Bearer token\n"
	if _, err := parseSecret(valid); err != nil {
		t.Fatalf("expected valid configuration: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(valid))
	if _, err := parseSecret(encoded); err != nil {
		t.Fatalf("expected valid base64-encoded configuration: %v", err)
	}
	cfg, err := parseSecret(valid)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.AuditLogsEnabled || cfg.OTLPHTTP.Endpoint != "https://otlp.example.com" {
		t.Fatalf("unexpected parsed configuration: %#v", cfg)
	}
}

func TestParseSecretAllowsAuditLogsToBeDisabled(t *testing.T) {
	tests := []string{
		"version: v1\n",
		"version: v1\ntelemetry:\n  logs:\n    audit:\n      enabled: false\n",
	}
	for _, value := range tests {
		cfg, err := parseSecret(value)
		if err != nil {
			t.Fatalf("expected disabled configuration to be valid: %v", err)
		}
		if cfg.AuditLogsEnabled {
			t.Fatal("disabled configuration enabled audit logs")
		}
	}
}

func TestParseSecretRejectsInvalidV1(t *testing.T) {
	tests := map[string]string{
		"missing version":            strings.TrimPrefix(auditEnabledConfig("https://example.com"), "version: v1\n"),
		"unsupported version":        "version: v2\n",
		"enabled without exporter":   "version: v1\ntelemetry:\n  logs:\n    audit:\n      enabled: true\n",
		"insecure endpoint":          auditEnabledConfig("http://otlp.example.com"),
		"endpoint with user info":    auditEnabledConfig("https://user@example.com"),
		"endpoint with fragment":     auditEnabledConfig("https://example.com/#fragment"),
		"endpoint with query":        auditEnabledConfig("https://example.com/?token=value"),
		"endpoint with expansion":    auditEnabledConfig("https://${env:RUNNER_API_TOKEN}"),
		"unknown exporter field":     auditEnabledConfig("https://example.com") + "    unknown: true\n",
		"unknown telemetry category": "version: v1\ntelemetry:\n  logs:\n    billing:\n      enabled: true\n",
		"unknown top-level field":    "version: v1\nunknown: true\n",
		"trailing document":          auditEnabledConfig("https://example.com") + "---\nversion: v1\n",
		"invalid header name":        auditEnabledConfig("https://example.com") + "    headers:\n      'bad header': value\n",
		"empty header value":         auditEnabledConfig("https://example.com") + "    headers:\n      Authorization: ''\n",
		"header with expansion":      auditEnabledConfig("https://example.com") + "    headers:\n      Authorization: '${env:RUNNER_API_TOKEN}'\n",
	}
	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := parseSecret(value); err == nil {
				t.Errorf("expected configuration to be rejected: %q", value)
			}
		})
	}
}

func TestChildEnvironmentOmitsRunnerCredentials(t *testing.T) {
	t.Setenv("RUNNER_API_TOKEN", "runner-secret")
	t.Setenv("HTTPS_PROXY", "https://proxy.example.com")
	environment := childEnvironment([]string{"NUON_TELEMETRY_EXPORT_HEADER_0=customer-secret"})
	joined := strings.Join(environment, "\n")
	if strings.Contains(joined, "runner-secret") || strings.Contains(joined, "RUNNER_API_TOKEN") {
		t.Fatal("collector child inherited the runner API credential")
	}
	if !strings.Contains(joined, "HTTPS_PROXY=https://proxy.example.com") {
		t.Fatal("collector child did not inherit HTTPS proxy configuration")
	}
	if !strings.Contains(joined, "NUON_TELEMETRY_EXPORT_HEADER_0=customer-secret") {
		t.Fatal("collector child did not receive the customer header")
	}
}

func TestCollectorConfigKeepsHeaderValuesOutOfFile(t *testing.T) {
	cfg, err := parseSecret(auditEnabledConfig("https://otlp.example.com") + "    headers:\n      Authorization: super-secret-value\n")
	if err != nil {
		t.Fatal(err)
	}
	contents, environment, err := collectorConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(contents), "super-secret-value") {
		t.Fatal("generated collector configuration contains a header value")
	}
	if !strings.Contains(string(contents), "${env:NUON_TELEMETRY_EXPORT_HEADER_0}") {
		t.Fatal("generated collector configuration does not reference header environment variable")
	}
	if len(environment) != 1 || !strings.HasSuffix(environment[0], "=super-secret-value") {
		t.Fatalf("unexpected child environment: %q", environment)
	}
	if !strings.Contains(string(contents), "127.0.0.1:4318") || !strings.Contains(string(contents), "logs/audit:") {
		t.Fatal("generated collector configuration lacks loopback logs pipeline")
	}
	if !strings.Contains(string(contents), `attributes["nuon.audit"] != "true"`) {
		t.Fatal("generated collector configuration does not filter non-audit logs")
	}
	if !strings.Contains(string(contents), "- filter/audit") {
		t.Fatal("generated collector pipeline does not use the audit filter")
	}
}

func TestCollectorConfigOmitsAuditPipelineWhenDisabled(t *testing.T) {
	cfg, err := parseSecret("version: v1\n")
	if err != nil {
		t.Fatal(err)
	}
	contents, environment, err := collectorConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(environment) != 0 {
		t.Fatalf("disabled audit pipeline produced an environment: %q", environment)
	}
	config := string(contents)
	if !strings.Contains(config, "health_check:") || !strings.Contains(config, "pipelines: {}") {
		t.Fatalf("generated collector configuration does not contain an empty service: %s", config)
	}
	for _, unexpected := range []string{"127.0.0.1:4318", "logs/audit:", "filter/audit:", "otlp_http:"} {
		if strings.Contains(config, unexpected) {
			t.Fatalf("generated collector configuration contains disabled audit component %q", unexpected)
		}
	}
}
