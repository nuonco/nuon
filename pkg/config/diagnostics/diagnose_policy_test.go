package diagnostics

import (
	"strings"
	"testing"
)

func TestDiagnoseFileDuplicatePolicyNames(t *testing.T) {
	text := `# policies

[[policy]]
name = "warn-public-sensitive-ports"
type = "sandbox"
engine = "opa"
contents = "package nuon"

[[policy]]
name = "warn-public-sensitive-ports"
type = "terraform_module"
engine = "opa"
contents = "package nuon"
`
	if !hasMessage(DiagnoseFile("policies.toml", text), `duplicate policy name "warn-public-sensitive-ports"`) {
		t.Fatalf("missing duplicate name diagnostic: %#v", DiagnoseFile("policies.toml", text))
	}
}

func TestDiagnoseFileDuplicateDerivedPolicyNames(t *testing.T) {
	text := `# policies

[[policy]]
type = "sandbox"
engine = "opa"
contents = "./policies/warn-public-sensitive-ports.rego"

[[policy]]
type = "terraform_module"
engine = "opa"
contents = "./policies/warn-public-sensitive-ports.rego"
`
	if !hasMessage(DiagnoseFile("policies.toml", text), `duplicate policy name "warn-public-sensitive-ports"`) {
		t.Fatalf("missing derived duplicate name diagnostic: %#v", DiagnoseFile("policies.toml", text))
	}
}

func TestDiagnoseFileDistinctPolicyNames(t *testing.T) {
	text := `# policies

[[policy]]
name = "sandbox-ports"
type = "sandbox"
engine = "opa"
contents = "package nuon"

[[policy]]
name = "module-ports"
type = "terraform_module"
engine = "opa"
contents = "package nuon"
`
	if hasMessage(DiagnoseFile("policies.toml", text), "duplicate policy name") {
		t.Fatalf("distinct names were reported as duplicates: %#v", DiagnoseFile("policies.toml", text))
	}
}

func hasMessage(diags []Diagnostic, message string) bool {
	for _, diag := range diags {
		if strings.Contains(diag.Message, message) {
			return true
		}
	}
	return false
}
