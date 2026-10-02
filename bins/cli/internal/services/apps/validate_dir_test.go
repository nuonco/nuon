package apps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/config/diagnostics"
)

func TestDiagnoseConfigDirReportsEveryFile(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, contents string) {
		t.Helper()
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("components/0-curl-tools.toml", "# container image\nname = \"curl_tools\"\ntype = \"container_image\"\n")
	write("components/1-ec2.toml", "# terraform-module\nname = \"ec2\"\ntype = \"terraform_module\"\n")
	write("actions/ok.toml", "# action\nname = \"ok\"\ntimeout = \"30s\"\n\n[[triggers]]\ntype = \"manual\"\n\n[[steps]]\nname = \"run\"\ncommand = \"true\"\n")
	write("runner.toml", "# runner\nrunner_type = 1\nnot_a_field = \"break-schema\"\n")
	write("metadata.toml", "version = \"v1\"\n")
	write(".git/hidden.toml", "# container image\nname = \"hidden\"\n")

	found, err := diagnoseConfigDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	var errors, warnings []string
	for _, diag := range found {
		line := formatConfigDiagnostic(diag)
		if strings.Contains(line, ".git/") {
			t.Fatalf("scanned hidden file: %s", line)
		}
		if diag.Severity == diagnostics.SeverityWarning {
			warnings = append(warnings, line)
			continue
		}
		errors = append(errors, line)
	}

	joined := strings.Join(errors, "\n")
	for _, want := range []string{
		"components/0-curl-tools.toml:1:1: error: Unknown schema type 'container image'",
		"components/1-ec2.toml:1:1: error: Unknown schema type 'terraform-module'",
		"runner.toml:2:1: error: Type mismatch for 'runner_type'",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "actions/ok.toml") || strings.Contains(joined, "metadata.toml") {
		t.Errorf("valid files reported as errors:\n%s", joined)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "runner.toml") || !strings.Contains(warnings[0], "Unknown key: not_a_field") {
		t.Errorf("warnings = %#v", warnings)
	}
}
