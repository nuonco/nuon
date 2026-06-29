package arm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// nestedBracketRe matches ARM expressions that contain nested square brackets,
// e.g. "[guid(..., [take(...)], ...)]". A well-formed ARM expression has exactly
// one outermost pair of square brackets; any inner '[' is a syntax error.
var nestedBracketRe = regexp.MustCompile(`\[[^\]]*\[`)

func TestGetAzureTemplate_NoNestedBrackets(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	inp := minimalTemplateInput()
	armTmpl, err := tmpl.getAzureTemplate(inp)
	if err != nil {
		t.Fatalf("getAzureTemplate returned error: %v", err)
	}

	tmplBytes, err := json.MarshalIndent(armTmpl, "", "  ")
	if err != nil {
		t.Fatalf("unable to marshal ARM template: %v", err)
	}

	assertNoNestedBrackets(t, tmplBytes)
}

func TestGetAzureTemplate_WithSecrets(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	inp := minimalTemplateInput()
	inp.AppCfg.SecretsConfig.Secrets = []app.AppSecretConfig{
		{Name: "github_app_key", Required: true},
		{Name: "db_password", Required: false},
	}

	armTmpl, err := tmpl.getAzureTemplate(inp)
	if err != nil {
		t.Fatalf("getAzureTemplate returned error: %v", err)
	}

	tmplBytes, err := json.MarshalIndent(armTmpl, "", "  ")
	if err != nil {
		t.Fatalf("unable to marshal ARM template: %v", err)
	}

	assertNoNestedBrackets(t, tmplBytes)
}

// TestDefaultVNet_NonPruningSubnetSemantics guards the fix for
// InUseSubnetCannotBeDeleted on reprovision: subnets are declared as standalone
// child resources and omitted from the VNet's own properties, which only
// preserves existing subnets on a VNet PUT when the API version is >= 2023-09-01.
func TestDefaultVNet_NonPruningSubnetSemantics(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	deployment := tmpl.getDefaultVNetDeployment(minimalTemplateInput())

	props, ok := deployment["properties"].(map[string]any)
	if !ok {
		t.Fatalf("vnetDeployment missing properties")
	}
	innerTmpl, ok := props["template"].(map[string]any)
	if !ok {
		t.Fatalf("vnetDeployment missing inner template")
	}
	resources, ok := innerTmpl["resources"].([]any)
	if !ok {
		t.Fatalf("inner template missing resources")
	}

	var foundVNet bool
	for _, r := range resources {
		res, ok := r.(map[string]any)
		if !ok || res["type"] != "Microsoft.Network/virtualNetworks" {
			continue
		}
		foundVNet = true

		if got := res["apiVersion"]; got != azureVNetAPIVersion {
			t.Errorf("VNet apiVersion = %v, want %s", got, azureVNetAPIVersion)
		}
		if azureVNetAPIVersion < "2023-09-01" {
			t.Errorf("azureVNetAPIVersion %s is < 2023-09-01; omitting subnets on a VNet PUT would prune in-use subnets", azureVNetAPIVersion)
		}

		vnetProps, ok := res["properties"].(map[string]any)
		if !ok {
			t.Fatalf("VNet missing properties")
		}
		if _, hasSubnets := vnetProps["subnets"]; hasSubnets {
			t.Error("VNet properties must NOT declare inline subnets; they are standalone child resources to avoid pruning foreign/out-of-band subnets")
		}
	}
	if !foundVNet {
		t.Fatal("no Microsoft.Network/virtualNetworks resource found in default VNet deployment")
	}
}

func assertNoNestedBrackets(t *testing.T, tmplBytes []byte) {
	t.Helper()

	lines := splitLines(tmplBytes)
	for i, line := range lines {
		if nestedBracketRe.MatchString(line) {
			t.Errorf("nested ARM brackets at line %d: %s", i+1, line)
		}
	}
}

func splitLines(b []byte) []string {
	var lines []string
	start := 0
	for i, c := range b {
		if c == '\n' {
			lines = append(lines, string(b[start:i]))
			start = i + 1
		}
	}
	if start < len(b) {
		lines = append(lines, string(b[start:]))
	}
	return lines
}

func minimalTemplateInput() *stacks.TemplateInput {
	return &stacks.TemplateInput{
		Install: &app.Install{
			ID:    "test-install-id-00000001",
			AppID: "test-app-id-000000000001",
			AzureAccount: &app.AzureAccount{
				Location: "eastus",
			},
		},
		CloudFormationStackVersion: &app.InstallStackVersion{
			PhoneHomeURL: fmt.Sprintf("https://api.example.com/phone-home/%s", "phid-test"),
		},
		Runner: &app.Runner{
			ID:    "test-runner-id-0000000001",
			OrgID: "test-org-id-00000000001",
		},
		Settings: &app.RunnerGroupSettings{
			ContainerImageURL: "example.com/runner",
			ContainerImageTag: "latest",
		},
		AppCfg: &app.AppConfig{},
	}
}
