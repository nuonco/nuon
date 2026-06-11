package installs

import (
	"strings"
	"testing"
)

// TestParseInstallConfig_RejectsMalformedOverride proves the CLI parse path
// validates component-override syntax (Parse + Validate) before any API call,
// so a malformed Helm values / tfvars override fails fast at config load.
func TestParseInstallConfig_RejectsMalformedOverride(t *testing.T) {
	t.Run("invalid helm_values", func(t *testing.T) {
		raw := `name = "abcd"

[components.whoami]
helm_values = """
replicaCount: [5
"""
`
		_, err := parseInstallConfig(strings.NewReader(raw))
		if err == nil {
			t.Fatal("expected error for malformed helm_values, got nil")
		}
		if !strings.Contains(err.Error(), "helm_values") {
			t.Fatalf("error should mention helm_values, got: %v", err)
		}
	})

	t.Run("invalid tf_vars", func(t *testing.T) {
		raw := `name = "abcd"

[components.vpc]
tf_vars = """
cidr = =
"""
`
		_, err := parseInstallConfig(strings.NewReader(raw))
		if err == nil {
			t.Fatal("expected error for malformed tf_vars, got nil")
		}
		if !strings.Contains(err.Error(), "tf_vars") {
			t.Fatalf("error should mention tf_vars, got: %v", err)
		}
	})

	t.Run("valid overrides parse", func(t *testing.T) {
		raw := `name = "abcd"

[components.whoami]
helm_values = """
replicaCount: 5
"""

[components.vpc]
tf_vars = """
cidr = "10.0.0.0/16"
"""
`
		cfg, err := parseInstallConfig(strings.NewReader(raw))
		if err != nil {
			t.Fatalf("expected valid config to parse, got: %v", err)
		}
		if len(cfg.Components) != 2 {
			t.Fatalf("expected 2 component overrides, got %d", len(cfg.Components))
		}
	})
}
