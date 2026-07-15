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

// schemaTypeForFile maps a generated config file (relative path) to its schema
// type slug.
func schemaTypeForFile(rel string) string {
	switch {
	case strings.HasPrefix(rel, "input_groups/"):
		return "input-group"
	case strings.HasPrefix(rel, "inputs/"):
		return "input"
	case strings.HasPrefix(rel, "policies/"):
		return "policy"
	case strings.HasPrefix(rel, "actions/"):
		return "action"
	case strings.HasPrefix(rel, "components/"):
		switch {
		case strings.Contains(rel, "helm"):
			return "helm"
		case strings.Contains(rel, "terraform"):
			return "terraform"
		case strings.Contains(rel, "kubernetes"):
			return "kubernetes-manifest"
		}
		return ""
	}
	name := filepath.Base(rel)
	switch {
	case name == "metadata.toml":
		return "metadata"
	case name == "sandbox.toml":
		return "sandbox"
	case name == "runner.toml":
		return "runner"
	case name == "stack.toml":
		return "stack"
	case name == "installer.toml":
		return "installer"
	case name == "secrets.toml":
		return "secrets"
	case name == "break_glass.toml":
		return "break-glass"
	case name == "provision.toml", name == "maintenance.toml", name == "deprovision.toml":
		return "permissions"
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
		typ := schemaTypeForFile(rel)
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
