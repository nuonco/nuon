package configdiff

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
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
		"env_vars":          func(c *app.AppSandboxConfig) { c.EnvVars = hstore(map[string]string{"LOG": "debug"}) },
		"references":        func(c *app.AppSandboxConfig) { c.References = []string{"comp.other"} },
		"operation_roles":   func(c *app.AppSandboxConfig) { c.OperationRoles = hstore(map[string]string{"apply": "admin"}) },
		"aws_region_type":   func(c *app.AppSandboxConfig) { c.AWSRegionType = generics.NewNullString("global") },
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

// A sandbox source bump leaves every scalar field identical but changes the
// infrastructure that gets deployed. Missing it means the new sandbox never
// reaches installs, because no reprovision step is emitted.
func TestSandboxConfigEqualDetectsSourceChange(t *testing.T) {
	base := func() app.AppSandboxConfig {
		return app.AppSandboxConfig{
			Type:             "terraform",
			TerraformVersion: "1.5.7",
			PublicGitVCSConfig: &app.PublicGitVCSConfig{
				Repo:      "nuonco/sandboxes",
				Directory: "aws-eks",
				Branch:    "v1.0.0",
			},
		}
	}

	cases := map[string]func(c *app.AppSandboxConfig){
		"branch":    func(c *app.AppSandboxConfig) { c.PublicGitVCSConfig.Branch = "v1.1.0" },
		"directory": func(c *app.AppSandboxConfig) { c.PublicGitVCSConfig.Directory = "aws-ecs" },
		"repo":      func(c *app.AppSandboxConfig) { c.PublicGitVCSConfig.Repo = "nuonco/other" },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			a, b := base(), base()
			mutate(&b)
			if sandboxConfigEqual(a, b) {
				t.Fatalf("expected a sandbox source %s change to be detected", name)
			}
		})
	}

	t.Run("switching source kind", func(t *testing.T) {
		a := base()
		b := base()
		b.PublicGitVCSConfig = nil
		b.ConnectedGithubVCSConfig = &app.ConnectedGithubVCSConfig{
			Repo:      "nuonco/sandboxes",
			Directory: "aws-eks",
			Branch:    "v1.0.0",
		}
		if sandboxConfigEqual(a, b) {
			t.Fatal("expected a public-git to connected-github switch to be detected")
		}
	})

	t.Run("identical source with different row ids", func(t *testing.T) {
		a, b := base(), base()
		a.PublicGitVCSConfig.ID = "pgvcold"
		b.PublicGitVCSConfig.ID = "pgvcnew"
		if !sandboxConfigEqual(a, b) {
			t.Fatal("an unchanged source should not read as changed just because its row is new")
		}
	})
}

// Retry and approval knobs are orchestration, not infrastructure — changing one
// must not trigger a full sandbox reprovision.
func TestSandboxConfigEqualIgnoresOrchestrationKnobs(t *testing.T) {
	a := app.AppSandboxConfig{Type: "terraform", TerraformVersion: "1.5.7"}

	retries := 5
	skipNoops := true
	b := a
	b.MaxAutoRetries = &retries
	b.SkipNoops = &skipNoops

	if !sandboxConfigEqual(a, b) {
		t.Fatal("orchestration knobs should not count as a sandbox content change")
	}
}

func TestStackConfigEqualIgnoresID(t *testing.T) {
	a := app.AppStackConfig{
		Type: app.StackTypeGCP,
		Name: "acme-{{.nuon.install.id}}",
	}
	a.ID = "appstackcfgold"

	b := a
	b.ID = "appstackcfgnew"

	if !StackConfigEqual(a, b) {
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
		"deployment_scope":           func(c *app.AppStackConfig) { c.DeploymentScope = app.StackDeploymentScopeSubscription },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			b := a
			mutate(&b)
			if StackConfigEqual(a, b) {
				t.Fatalf("expected a %s change to be detected", name)
			}
		})
	}
}
