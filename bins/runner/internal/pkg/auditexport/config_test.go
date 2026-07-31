package auditexport

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestParseSecret(t *testing.T) {
	valid := "exporters:\n  otlphttp:\n    endpoint: https://otlp.example.com\n    headers:\n      Authorization: Bearer token\n"
	if _, err := parseSecret(valid); err != nil {
		t.Fatalf("expected valid configuration: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(valid))
	if _, err := parseSecret(encoded); err != nil {
		t.Fatalf("expected valid base64-encoded configuration: %v", err)
	}

	tests := []string{
		"exporters:\n  otlphttp:\n    endpoint: http://otlp.example.com\n",
		"exporters:\n  otlphttp:\n    endpoint: https://user@example.com\n",
		"exporters:\n  otlphttp:\n    endpoint: https://example.com/#fragment\n",
		"exporters:\n  otlphttp:\n    endpoint: https://example.com/?token=value\n",
		"exporters:\n  otlphttp:\n    endpoint: https://${env:RUNNER_API_TOKEN}\n",
		"exporters:\n  otlphttp:\n    endpoint: https://example.com\n    unknown: true\n",
		"exporters:\n  otlphttp:\n    endpoint: https://example.com\n---\nexporters: {}\n",
		"exporters:\n  otlphttp:\n    endpoint: https://example.com\n    headers:\n      'bad header': value\n",
		"exporters:\n  otlphttp:\n    endpoint: https://example.com\n    headers:\n      Authorization: ''\n",
		"exporters:\n  otlphttp:\n    endpoint: https://example.com\n    headers:\n      Authorization: '${env:RUNNER_API_TOKEN}'\n",
	}
	for _, value := range tests {
		if _, err := parseSecret(value); err == nil {
			t.Errorf("expected configuration to be rejected: %q", value)
		}
	}
}

func TestChildEnvironmentOmitsRunnerCredentials(t *testing.T) {
	t.Setenv("RUNNER_API_TOKEN", "runner-secret")
	t.Setenv("HTTPS_PROXY", "https://proxy.example.com")
	environment := childEnvironment([]string{"NUON_AUDIT_HEADER_0=customer-secret"})
	joined := strings.Join(environment, "\n")
	if strings.Contains(joined, "runner-secret") || strings.Contains(joined, "RUNNER_API_TOKEN") {
		t.Fatal("collector child inherited the runner API credential")
	}
	if !strings.Contains(joined, "HTTPS_PROXY=https://proxy.example.com") {
		t.Fatal("collector child did not inherit HTTPS proxy configuration")
	}
	if !strings.Contains(joined, "NUON_AUDIT_HEADER_0=customer-secret") {
		t.Fatal("collector child did not receive the customer header")
	}
}

func TestCollectorConfigKeepsHeaderValuesOutOfFile(t *testing.T) {
	cfg, err := parseSecret("exporters:\n  otlphttp:\n    endpoint: https://otlp.example.com\n    headers:\n      Authorization: super-secret-value\n")
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
	if !strings.Contains(string(contents), "${env:NUON_AUDIT_HEADER_0}") {
		t.Fatal("generated collector configuration does not reference header environment variable")
	}
	if len(environment) != 1 || !strings.HasSuffix(environment[0], "=super-secret-value") {
		t.Fatalf("unexpected child environment: %q", environment)
	}
	if !strings.Contains(string(contents), "127.0.0.1:4318") || !strings.Contains(string(contents), "logs:") {
		t.Fatal("generated collector configuration lacks loopback logs pipeline")
	}
	if !strings.Contains(string(contents), `attributes["nuon.audit"] != "true"`) {
		t.Fatal("generated collector configuration does not filter non-audit logs")
	}
	if !strings.Contains(string(contents), "- filter/audit") {
		t.Fatal("generated collector pipeline does not use the audit filter")
	}
}
