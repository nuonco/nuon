package config

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/pkg/config/diff"
)

func TestStructuredHelmValuesDiff(t *testing.T) {
	t.Run("nested change renders per-key diff", func(t *testing.T) {
		oldYAML := "replicaCount: 4\nresources:\n  requests:\n    cpu: \"150m\"\n    memory: 64Mi\n"
		newYAML := "replicaCount: 5\nresources:\n  requests:\n    cpu: \"150m\"\n    memory: 64Mi\n"

		d, ok := structuredHelmValuesDiff("key", oldYAML, newYAML)
		if !ok {
			t.Fatal("expected structured diff, got fallback")
		}

		got := d.String("")
		want := "key:\n" +
			"\treplicaCount: '4' -> '5'\n" +
			"\tresources:\n" +
			"\t\trequests:\n" +
			"\t\t\tcpu: '\"150m\"' (unchanged)\n" +
			"\t\t\tmemory: '\"64Mi\"' (unchanged)\n"
		if got != want {
			t.Fatalf("unexpected diff:\n got:\n%q\nwant:\n%q", got, want)
		}

		summary := d.Summary()
		if summary.Changed != 1 || summary.Unchanged != 2 {
			t.Fatalf("unexpected summary: %+v", summary)
		}
	})

	t.Run("new install shows all keys as additions", func(t *testing.T) {
		d, ok := structuredHelmValuesDiff("key", "", "replicaCount: 5\n")
		if !ok {
			t.Fatal("expected structured diff, got fallback")
		}
		summary := d.Summary()
		if summary.Added != 1 {
			t.Fatalf("expected 1 add, got %+v", summary)
		}
	})

	t.Run("removed key", func(t *testing.T) {
		d, ok := structuredHelmValuesDiff("key", "a: 1\nb: 2\n", "a: 1\n")
		if !ok {
			t.Fatal("expected structured diff, got fallback")
		}
		summary := d.Summary()
		if summary.Removed != 1 || summary.Unchanged != 1 {
			t.Fatalf("unexpected summary: %+v", summary)
		}
	})

	t.Run("list values render compactly", func(t *testing.T) {
		d, ok := structuredHelmValuesDiff("key", "args:\n  - a\n  - b\n", "args:\n  - a\n  - c\n")
		if !ok {
			t.Fatal("expected structured diff, got fallback")
		}
		got := d.String("")
		want := "key:\n\targs: '[\"a\",\"b\"]' -> '[\"a\",\"c\"]'\n"
		if got != want {
			t.Fatalf("unexpected diff:\n got: %q\nwant: %q", got, want)
		}
	})

	t.Run("adding an empty-string value is detected", func(t *testing.T) {
		d, ok := structuredHelmValuesDiff("key", "a: 1\n", "a: 1\nb: \"\"\n")
		if !ok {
			t.Fatal("expected structured diff, got fallback")
		}
		if s := d.Summary(); !s.HasChanged || s.Added != 1 {
			t.Fatalf("expected 1 add, got %+v", s)
		}
	})

	t.Run("adding a null value is detected", func(t *testing.T) {
		d, ok := structuredHelmValuesDiff("key", "a: 1\n", "a: 1\nb: null\n")
		if !ok {
			t.Fatal("expected structured diff, got fallback")
		}
		if s := d.Summary(); !s.HasChanged || s.Added != 1 {
			t.Fatalf("expected 1 add, got %+v", s)
		}
	})

	t.Run("adding an empty mapping is detected", func(t *testing.T) {
		d, ok := structuredHelmValuesDiff("key", "a: 1\n", "a: 1\nb: {}\n")
		if !ok {
			t.Fatal("expected structured diff, got fallback")
		}
		if s := d.Summary(); !s.HasChanged || s.Added != 1 {
			t.Fatalf("expected 1 add, got %+v", s)
		}
	})

	t.Run("non-mapping falls back to string diff", func(t *testing.T) {
		if _, ok := structuredHelmValuesDiff("key", "just a scalar", "another scalar"); ok {
			t.Fatal("expected fallback for non-mapping yaml")
		}
	})

	t.Run("invalid yaml falls back to string diff", func(t *testing.T) {
		if _, ok := structuredHelmValuesDiff("key", "a: 1", "::: not yaml :::\n  - broken"); ok {
			t.Fatal("expected fallback for invalid yaml")
		}
	})
}

func TestStructuredTFVarsDiff(t *testing.T) {
	t.Run("hcl scalar change renders per-key diff", func(t *testing.T) {
		oldVars := "domain_name = \"whoami.nuon.run\"\nreplicas = 2\n"
		newVars := "domain_name = \"whoami.example.com\"\nreplicas = 2\n"

		d, ok := structuredTFVarsDiff("key", oldVars, newVars)
		if !ok {
			t.Fatal("expected structured diff, got fallback")
		}

		got := d.String("")
		want := "key:\n" +
			"\tdomain_name: '\"whoami.nuon.run\"' -> '\"whoami.example.com\"'\n" +
			"\treplicas: '2' (unchanged)\n"
		if got != want {
			t.Fatalf("unexpected diff:\n got:\n%q\nwant:\n%q", got, want)
		}

		if s := d.Summary(); s.Changed != 1 || s.Unchanged != 1 {
			t.Fatalf("unexpected summary: %+v", s)
		}
	})

	t.Run("nested object recurses", func(t *testing.T) {
		oldVars := "tags = { env = \"dev\", team = \"core\" }\n"
		newVars := "tags = { env = \"prod\", team = \"core\" }\n"

		d, ok := structuredTFVarsDiff("key", oldVars, newVars)
		if !ok {
			t.Fatal("expected structured diff, got fallback")
		}
		got := d.String("")
		want := "key:\n" +
			"\ttags:\n" +
			"\t\tenv: '\"dev\"' -> '\"prod\"'\n" +
			"\t\tteam: '\"core\"' (unchanged)\n"
		if got != want {
			t.Fatalf("unexpected diff:\n got:\n%q\nwant:\n%q", got, want)
		}
	})

	t.Run("json tfvars", func(t *testing.T) {
		oldVars := `{"domain_name": "a.com"}`
		newVars := `{"domain_name": "b.com"}`

		d, ok := structuredTFVarsDiff("key", oldVars, newVars)
		if !ok {
			t.Fatal("expected structured diff, got fallback")
		}
		got := d.String("")
		want := "key:\n\tdomain_name: '\"a.com\"' -> '\"b.com\"'\n"
		if got != want {
			t.Fatalf("unexpected diff:\n got: %q\nwant: %q", got, want)
		}
	})

	t.Run("new install shows additions", func(t *testing.T) {
		d, ok := structuredTFVarsDiff("key", "", "domain_name = \"a.com\"\n")
		if !ok {
			t.Fatal("expected structured diff, got fallback")
		}
		if s := d.Summary(); s.Added != 1 {
			t.Fatalf("expected 1 add, got %+v", s)
		}
	})

	t.Run("list values render compactly", func(t *testing.T) {
		d, ok := structuredTFVarsDiff("key", "zones = [\"a\", \"b\"]\n", "zones = [\"a\", \"c\"]\n")
		if !ok {
			t.Fatal("expected structured diff, got fallback")
		}
		got := d.String("")
		want := "key:\n\tzones: '[\"a\",\"b\"]' -> '[\"a\",\"c\"]'\n"
		if got != want {
			t.Fatalf("unexpected diff:\n got: %q\nwant: %q", got, want)
		}
	})

	t.Run("invalid hcl falls back to string diff", func(t *testing.T) {
		if _, ok := structuredTFVarsDiff("key", "domain = \"a\"\n", "this is = = not valid"); ok {
			t.Fatal("expected fallback for invalid hcl")
		}
	})

	t.Run("non-literal expression falls back to string diff", func(t *testing.T) {
		if _, ok := structuredTFVarsDiff("key", "a = 1\n", "a = var.something\n"); ok {
			t.Fatal("expected fallback for non-literal tfvars")
		}
	})
}

func TestStructuredHelmValuesDiff_Unchanged(t *testing.T) {
	yamlStr := "replicaCount: 4\n"
	d, ok := structuredHelmValuesDiff("key", yamlStr, yamlStr)
	if !ok {
		t.Fatal("expected structured diff")
	}
	if s := d.Summary(); s.HasChanged {
		t.Fatalf("expected no changes, got %+v", s)
	}
	_ = diff.OpNoop
}

func TestInstallDiff_StructuredValueTypeProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		number := strconv.FormatInt(hegel.Draw(t, hegel.Integers[int64](0, 1<<60)), 10)
		word := hegel.Draw(t, hegel.FromRegex(`[a-z]{1,12}`, true))
		for _, literal := range []string{"true", "false", "null", number, "[]", "{}", `["` + word + `"]`, `{"item":"` + word + `"}`} {
			oldValue, newValue := literal, strconv.Quote(literal)
			for range hegel.Draw(t, hegel.Integers(0, 3)) {
				oldValue, newValue = `{"nested":`+oldValue+`}`, `{"nested":`+newValue+`}`
			}
			oldDocument := `{"value":` + oldValue + `,"control":"unchanged"}`
			newDocument := `{"value":` + newValue + `,"control":"unchanged"}`
			for _, kind := range []ComponentOverrideKind{ComponentOverrideKindHelmValues, ComponentOverrideKindTFVars} {
				makeInstall := func(document string) *Install {
					override := ComponentOverride{HelmValues: document}
					if kind == ComponentOverrideKindTFVars {
						override = ComponentOverride{TFVars: document}
					}
					return &Install{Components: map[string]ComponentOverride{"api": override}}
				}
				old, new := makeInstall(oldDocument), makeInstall(newDocument)
				for _, pair := range [][2]*Install{{old, new}, {new, old}} {
					got, err := pair[1].Diff(pair[0])
					require.NoError(t, err)
					require.Equal(t, 1, got.Summary().Changed, "%s must distinguish %s from a string containing it", kind, literal)
					require.NotEmpty(t, got.FormatChanged(""))
				}
				unchanged, err := old.Diff(makeInstall("\n  " + oldDocument + "\n"))
				require.NoError(t, err)
				require.False(t, unchanged.Summary().HasChanged, "formatting must not change typed literals")
			}
		}
	})
}

func TestInstallDiff_StructuredFallbackProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		value := hegel.Draw(t, hegel.FromRegex(`[a-z]{1,12}`, true))
		for _, fixture := range []struct {
			key, old, next string
		}{
			{HelmValuesOverrideInputName("api"), "replicas: 1\n---\nvalue: " + value, "replicas: 1\n---\nvalue: changed_" + value},
			{HelmValuesOverrideInputName("api"), "replicas: 1\n---\nvalue: " + value, "replicas: 1\n---\n{"},
			{HelmValuesOverrideInputName("api"), "- " + value, "- changed_" + value},
			{TFVarsOverrideInputName("api"), "value = var." + value, "value = var.changed_" + value},
		} {
			old := &Install{InputGroups: []InputGroup{{Inputs: map[string]string{fixture.key: fixture.old}}}}
			next := &Install{InputGroups: []InputGroup{{Inputs: map[string]string{fixture.key: fixture.next}}}}
			got, err := next.Diff(old)
			require.NoError(t, err)
			require.Equal(t, 1, got.Summary().Changed, "unsupported structured content must retain its textual change")
			require.NotEmpty(t, got.FormatChanged(""))
		}
	})
}
