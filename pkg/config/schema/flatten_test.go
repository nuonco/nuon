package schema

import (
	"encoding/json"
	"strings"
	"testing"

	sjs "github.com/santhosh-tekuri/jsonschema/v6"
)

func compileStrict(t *testing.T, schemaType string) *sjs.Schema {
	t.Helper()

	schm, err := LookupSchemaType(schemaType)
	if err != nil {
		t.Fatalf("unable to build schema %s: %v", schemaType, err)
	}
	if schm == nil {
		t.Fatalf("no schema registered for %s", schemaType)
	}

	b, err := json.Marshal(schm)
	if err != nil {
		t.Fatalf("unable to marshal schema %s: %v", schemaType, err)
	}

	doc, err := sjs.UnmarshalJSON(strings.NewReader(string(b)))
	if err != nil {
		t.Fatalf("unable to parse schema %s: %v", schemaType, err)
	}

	c := sjs.NewCompiler()
	if err := c.AddResource("schema.json", doc); err != nil {
		t.Fatalf("unable to add schema %s: %v", schemaType, err)
	}
	compiled, err := c.Compile("schema.json")
	if err != nil {
		t.Fatalf("unable to compile schema %s: %v", schemaType, err)
	}
	return compiled
}

func validateDoc(t *testing.T, compiled *sjs.Schema, doc string) error {
	t.Helper()

	inst, err := sjs.UnmarshalJSON(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("invalid test document: %v", err)
	}
	return compiled.Validate(inst)
}

// TestComponentSchemasAcceptRealConfigs validates documents mirroring the TOML
// templates served by the API (config_templates.go) against the published
// schemas with a strict draft 2020-12 validator. The previous allOf-composed
// schemas were formally unsatisfiable: no document could validate.
func TestComponentSchemasAcceptRealConfigs(t *testing.T) {
	testCases := []struct {
		schemaType string
		name       string
		doc        string
	}{
		{
			schemaType: "container-image",
			name:       "public image",
			doc: `{
				"name": "toml_container_image",
				"type": "container_image",
				"public": {"image_url": "kennethreitz/httpbin", "tag": "latest"}
			}`,
		},
		{
			schemaType: "container-image",
			name:       "aws ecr image",
			doc: `{
				"name": "toml_container_image_ecr",
				"type": "container_image",
				"aws_ecr": {
					"iam_role_arn": "iam_role_arn",
					"image_url": "ecr-url",
					"tag": "latest",
					"region": "us-west-2"
				}
			}`,
		},
		{
			schemaType: "docker-build",
			name:       "docker build",
			doc: `{
				"name": "toml_docker_build",
				"type": "docker_build",
				"dockerfile": "Dockerfile",
				"connected_repo": {"directory": "deployment", "repo": "nuonco/nuon", "branch": "main"}
			}`,
		},
		{
			schemaType: "terraform",
			name:       "terraform module",
			doc: `{
				"name": "toml_terraform",
				"type": "terraform_module",
				"terraform_version": "1.7.5",
				"connected_repo": {"directory": "infra", "repo": "nuonco/nuon", "branch": "main"},
				"vars": {"AWS_REGION": "{{.nuon.install.sandbox.account.region}}"},
				"dependencies": ["infra"]
			}`,
		},
		{
			schemaType: "helm",
			name:       "helm chart",
			doc: `{
				"name": "toml_helm",
				"type": "helm_chart",
				"chart_name": "e2e-helm",
				"connected_repo": {"directory": "deployment", "repo": "nuonco/nuon", "branch": "main"},
				"values_file": [{"contents": "image.tag = latest"}],
				"values": {"api.ingresses.public_domain": "example.com"}
			}`,
		},
		{
			schemaType: "job",
			name:       "job",
			doc: `{
				"name": "toml_job",
				"type": "job",
				"image_url": "{{.nuon.components.e2e_docker_build.image.repository.uri}}",
				"tag": "latest",
				"cmd": ["printenv"],
				"args": [""],
				"env_vars": {"PUBLIC_DOMAIN": "example.com"}
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.schemaType+"/"+tc.name, func(t *testing.T) {
			compiled := compileStrict(t, tc.schemaType)
			if err := validateDoc(t, compiled, tc.doc); err != nil {
				t.Fatalf("expected document to validate: %v", err)
			}
		})
	}
}

// TestComponentSchemasStillRejectInvalidConfigs ensures flattening did not
// loosen the schemas: unknown keys and missing required fields must still fail.
func TestComponentSchemasStillRejectInvalidConfigs(t *testing.T) {
	testCases := []struct {
		schemaType string
		name       string
		doc        string
	}{
		{
			schemaType: "container-image",
			name:       "missing type and name",
			doc:        `{"public": {"image_url": "kennethreitz/httpbin", "tag": "latest"}}`,
		},
		{
			schemaType: "container-image",
			name:       "unknown property",
			doc: `{
				"name": "img",
				"type": "container_image",
				"public": {"image_url": "kennethreitz/httpbin", "tag": "latest"},
				"bogus": true
			}`,
		},
		{
			schemaType: "container-image",
			name:       "no image source",
			doc:        `{"name": "img", "type": "container_image"}`,
		},
		{
			schemaType: "terraform",
			name:       "unknown property",
			doc:        `{"name": "tf", "type": "terraform_module", "bogus": true}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.schemaType+"/"+tc.name, func(t *testing.T) {
			compiled := compileStrict(t, tc.schemaType)
			if err := validateDoc(t, compiled, tc.doc); err == nil {
				t.Fatal("expected document to fail validation")
			}
		})
	}
}

// TestComponentSchemasAreFlattened guards against reintroducing the
// unsatisfiable allOf composition.
func TestComponentSchemasAreFlattened(t *testing.T) {
	for _, schemaType := range []string{
		"container-image", "docker-build", "helm", "job", "kubernetes-manifest", "terraform",
	} {
		t.Run(schemaType, func(t *testing.T) {
			schm, err := LookupSchemaType(schemaType)
			if err != nil {
				t.Fatal(err)
			}
			if len(schm.AllOf) > 0 {
				t.Fatal("schema must not use top-level allOf composition")
			}
			if _, ok := schm.Properties.Get("name"); !ok {
				t.Fatal("schema must declare the shared component property \"name\"")
			}
			if _, ok := schm.Properties.Get("type"); !ok {
				t.Fatal("schema must declare the shared component property \"type\"")
			}
			for _, req := range []string{"type", "name"} {
				found := false
				for _, r := range schm.Required {
					if r == req {
						found = true
					}
				}
				if !found {
					t.Fatalf("schema must require %q", req)
				}
			}
		})
	}
}
