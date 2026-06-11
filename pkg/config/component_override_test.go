package config

import "testing"

func TestComponentOverrideInputName_RoundTrip(t *testing.T) {
	cases := []struct {
		name string
		kind ComponentOverrideKind
		comp string
	}{
		{"helm simple", ComponentOverrideKindHelmValues, "clickhouse"},
		{"tf simple", ComponentOverrideKindTFVars, "vpc"},
		{"hyphen", ComponentOverrideKindHelmValues, "clickhouse-operator"},
		{"underscore", ComponentOverrideKindTFVars, "api_server"},
		{"dots", ComponentOverrideKindHelmValues, "a.b.c"},
		{"mixed", ComponentOverrideKindTFVars, "My-Comp_1.2"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			switch tc.kind {
			case ComponentOverrideKindHelmValues:
				got = HelmValuesOverrideInputName(tc.comp)
			case ComponentOverrideKindTFVars:
				got = TFVarsOverrideInputName(tc.comp)
			}

			if !IsComponentOverrideInputName(got) {
				t.Fatalf("IsComponentOverrideInputName(%q) = false", got)
			}

			kind, comp, ok := ParseComponentOverrideInputName(got)
			if !ok {
				t.Fatalf("ParseComponentOverrideInputName(%q) ok=false", got)
			}
			if kind != tc.kind {
				t.Errorf("kind = %q, want %q", kind, tc.kind)
			}
			if comp != tc.comp {
				t.Errorf("comp = %q, want %q", comp, tc.comp)
			}
		})
	}
}

func TestComponentOverrideInputName_HyphenUnderscoreNoCollision(t *testing.T) {
	a := HelmValuesOverrideInputName("foo-bar")
	b := HelmValuesOverrideInputName("foo_bar")
	if a == b {
		t.Fatalf("expected distinct keys for foo-bar and foo_bar, both = %q", a)
	}
}

func TestComponentOverrideInputName_KindNoCollision(t *testing.T) {
	helm := HelmValuesOverrideInputName("vpc")
	tf := TFVarsOverrideInputName("vpc")
	if helm == tf {
		t.Fatalf("helm and tf override names collided for same component: %q", helm)
	}
}

func TestParseComponentOverrideInputName_NonOverride(t *testing.T) {
	for _, name := range []string{"replicas", "log_level", "nuon_component_override_v1_bogus_zz"} {
		if _, _, ok := ParseComponentOverrideInputName(name); ok {
			t.Errorf("ParseComponentOverrideInputName(%q) ok=true, want false", name)
		}
	}
}

func TestSyntheticComponentOverrideInputs(t *testing.T) {
	components := ComponentList{
		{Name: "vpc", Type: TerraformModuleComponentType},
		{Name: "clickhouse", Type: HelmChartComponentType},
		{Name: "api", Type: DockerBuildComponentType},             // ignored
		{Name: "manifest", Type: KubernetesManifestComponentType}, // ignored
		nil, // ignored
	}

	got := SyntheticComponentOverrideInputs(components)
	if len(got) != 2 {
		t.Fatalf("got %d synthetic inputs, want 2: %+v", len(got), got)
	}

	if got[0].Component != "vpc" || got[0].Kind != ComponentOverrideKindTFVars {
		t.Errorf("got[0] = %+v, want vpc/tf_vars", got[0])
	}
	if got[0].Name != TFVarsOverrideInputName("vpc") {
		t.Errorf("got[0].Name = %q, want %q", got[0].Name, TFVarsOverrideInputName("vpc"))
	}
	if got[1].Component != "clickhouse" || got[1].Kind != ComponentOverrideKindHelmValues {
		t.Errorf("got[1] = %+v, want clickhouse/helm_values", got[1])
	}

	// Indices must be distinct and non-zero so server-side `required` index
	// validation passes and ordering is stable.
	if got[0].Index == got[1].Index {
		t.Errorf("indices collided: %d", got[0].Index)
	}
	if got[0].Index == 0 || got[1].Index == 0 {
		t.Errorf("indices must be non-zero: %d, %d", got[0].Index, got[1].Index)
	}
}

func TestSyntheticComponentOverrideInputs_Empty(t *testing.T) {
	if got := SyntheticComponentOverrideInputs(nil); len(got) != 0 {
		t.Fatalf("got %d, want 0", len(got))
	}
}

func TestValidateInputs_RejectsReservedPrefix(t *testing.T) {
	cfg := AppInputConfig{
		Inputs: []AppInput{
			{Name: HelmValuesOverrideInputName("vpc"), Group: ""},
		},
	}
	if err := cfg.ValidateInputs(); err == nil {
		t.Fatal("expected ValidateInputs to reject reserved-prefix input name")
	}
}
