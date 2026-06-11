package config

import "testing"

func TestValidateInputValueSyntax(t *testing.T) {
	tests := []struct {
		name      string
		inputType string
		value     string
		wantErr   bool
	}{
		{"empty yaml is ok", InputTypeYAML, "", false},
		{"whitespace yaml is ok", InputTypeYAML, "   \n  ", false},
		{"valid yaml mapping", InputTypeYAML, "replicaCount: 5\nresources:\n  cpu: \"150m\"\n", false},
		{"valid yaml scalar", InputTypeYAML, "5", false},
		{"invalid yaml", InputTypeYAML, "foo: [bar\n", true},
		{"empty hcl is ok", InputTypeHCL, "", false},
		{"valid hcl tfvars", InputTypeHCL, "cidr_block = \"10.1.0.0/16\"\ncount = 3\n", false},
		{"valid json tfvars", InputTypeHCL, "{\"cidr_block\": \"10.1.0.0/16\"}", false},
		{"invalid hcl syntax", InputTypeHCL, "cidr_block = =\n", true},
		{"hcl block is not tfvars", InputTypeHCL, "resource \"x\" {}\n", true},
		{"unknown type is not checked", "string", "anything {[ goes", false},
		{"unknown type empty", "number", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInputValueSyntax(tt.inputType, tt.value)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestComponentOverrideKindInputType(t *testing.T) {
	if got := ComponentOverrideKindHelmValues.InputType(); got != InputTypeYAML {
		t.Fatalf("helm_values kind: want %q, got %q", InputTypeYAML, got)
	}
	if got := ComponentOverrideKindTFVars.InputType(); got != InputTypeHCL {
		t.Fatalf("tf_vars kind: want %q, got %q", InputTypeHCL, got)
	}
}
