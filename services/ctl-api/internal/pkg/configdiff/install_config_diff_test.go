package configdiff

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func hstore(kv map[string]string) pgtype.Hstore {
	h := pgtype.Hstore{}
	for k, v := range kv {
		val := v
		h[k] = &val
	}
	return h
}

// Every app config sync writes fresh sandbox and stack config rows, so the IDs
// always differ between two versions. Only a content change is a real change —
// comparing IDs reported a sandbox reprovision and a stack regeneration on
// every branch run.
func TestSandboxConfigEqualIgnoresID(t *testing.T) {
	a := app.AppSandboxConfig{
		Type:             "terraform",
		TerraformVersion: "1.5.7",
		Variables:        hstore(map[string]string{"region": "us-west-2"}),
	}
	a.ID = "appsandboxcfgold"

	b := a
	b.ID = "appsandboxcfgnew"

	if !sandboxConfigEqual(a, b) {
		t.Fatal("sandbox configs with identical content should compare equal regardless of ID")
	}
}

func TestSandboxConfigEqualDetectsContentChange(t *testing.T) {
	a := app.AppSandboxConfig{
		Type:             "terraform",
		TerraformVersion: "1.5.7",
		Variables:        hstore(map[string]string{"region": "us-west-2"}),
	}

	cases := map[string]func(c *app.AppSandboxConfig){
		"variables":         func(c *app.AppSandboxConfig) { c.Variables = hstore(map[string]string{"region": "us-east-1"}) },
		"terraform_version": func(c *app.AppSandboxConfig) { c.TerraformVersion = "1.6.0" },
		"type":              func(c *app.AppSandboxConfig) { c.Type = "pulumi" },
		"drift_schedule":    func(c *app.AppSandboxConfig) { c.DriftSchedule = "0 * * * *" },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			b := a
			mutate(&b)
			if sandboxConfigEqual(a, b) {
				t.Fatalf("expected a %s change to be detected", name)
			}
		})
	}
}

func TestStackConfigEqualIgnoresID(t *testing.T) {
	a := app.AppStackConfig{
		Type: app.StackTypeGCP,
		Name: "aqd-{{.nuon.install.id}}",
	}
	a.ID = "appstackcfgold"

	b := a
	b.ID = "appstackcfgnew"

	if !stackConfigEqual(a, b) {
		t.Fatal("stack configs with identical content should compare equal regardless of ID")
	}
}

func TestStackConfigEqualDetectsContentChange(t *testing.T) {
	a := app.AppStackConfig{
		Type:                    app.StackTypeAWS,
		Name:                    "nuon-install",
		RunnerNestedTemplateURL: "https://example.com/runner-v1.yaml",
	}

	cases := map[string]func(c *app.AppStackConfig){
		"type":                       func(c *app.AppStackConfig) { c.Type = app.StackTypeAzure },
		"name":                       func(c *app.AppStackConfig) { c.Name = "nuon-install-2" },
		"runner_nested_template_url": func(c *app.AppStackConfig) { c.RunnerNestedTemplateURL = "https://example.com/runner-v2.yaml" },
		"vpc_nested_template_url":    func(c *app.AppStackConfig) { c.VPCNestedTemplateURL = "https://example.com/vpc.yaml" },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			b := a
			mutate(&b)
			if stackConfigEqual(a, b) {
				t.Fatalf("expected a %s change to be detected", name)
			}
		})
	}
}
