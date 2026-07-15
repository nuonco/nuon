package generator_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/xeipuuv/gojsonschema"

	"github.com/nuonco/nuon/pkg/config/generator"
	"github.com/nuonco/nuon/pkg/config/schema"
)

// schemaTypeForFile maps a generated config filename to its schema type slug.
func schemaTypeForFile(name string) string {
	switch {
	case name == "sandbox.toml":
		return "sandbox"
	case name == "runner.toml":
		return "runner"
	case name == "stack.toml":
		return "stack"
	case name == "installer.toml":
		return "installer"
	case name == "inputs.toml":
		return "inputs"
	case name == "policies.toml":
		return "policies"
	case name == "secrets.toml":
		return "secrets"
	case name == "break_glass.toml":
		return "break-glass"
	case name == "provision.toml", name == "maintenance.toml", name == "deprovision.toml":
		return "permissions"
	case strings.Contains(name, "helm"):
		return "helm"
	case strings.Contains(name, "terraform"):
		return "terraform"
	case strings.Contains(name, "kubernetes"):
		return "kubernetes-manifest"
	case strings.HasPrefix(name, "example_action"):
		return "action"
	case strings.HasPrefix(name, "example_install"):
		return "install"
	}
	return ""
}

// TestGeneratedConfigsValidateAgainstSchemas guards that every file produced by
// the default `nuon apps init` scaffold validates against its dedicated schema.
func TestGeneratedConfigsValidateAgainstSchemas(t *testing.T) {
	dir := t.TempDir()

	// Mirror the default `nuon apps init` flags (SkipNonRequired=true).
	gen := generator.NewConfigGen(false, false, false, true, true, "")
	if err := gen.Gen(dir, generator.DefaultAppConfigConfigStructure(dir)); err != nil {
		t.Fatalf("generating config: %v", err)
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".toml") {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		typ := schemaTypeForFile(filepath.Base(path))
		if typ == "" {
			t.Fatalf("no schema type mapping for generated file %s", rel)
		}
		schm, serr := schema.LookupSchemaType(typ)
		if serr != nil || schm == nil {
			t.Fatalf("no schema for type %q (%s): %v", typ, rel, serr)
		}

		raw, _ := os.ReadFile(path)
		doc := map[string]any{}
		if derr := toml.Unmarshal(raw, &doc); derr != nil {
			t.Errorf("%s: invalid TOML: %v", rel, derr)
			return nil
		}
		jb, _ := json.Marshal(doc)
		res, verr := gojsonschema.Validate(gojsonschema.NewGoLoader(schm), gojsonschema.NewBytesLoader(jb))
		if verr != nil {
			t.Errorf("%s: validation error: %v", rel, verr)
			return nil
		}
		if !res.Valid() {
			for _, e := range res.Errors() {
				t.Errorf("%s (%s) invalid: %s", rel, typ, e)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
