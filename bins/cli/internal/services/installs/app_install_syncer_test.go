package installs

import (
	"testing"

	"github.com/nuonco/nuon/pkg/config"
)

func TestInstallDiffKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "plain input passes through unchanged",
			key:  "sub_domain",
			want: "sub_domain",
		},
		{
			name: "helm values override decodes to components.<name>.helm_values",
			key:  config.HelmValuesOverrideInputName("whoami"),
			want: "components.whoami.helm_values",
		},
		{
			name: "tf vars override decodes to components.<name>.tf_vars",
			key:  config.TFVarsOverrideInputName("certificate"),
			want: "components.certificate.tf_vars",
		},
		{
			name: "component name with underscores/dashes round-trips",
			key:  config.HelmValuesOverrideInputName("foo-bar_baz"),
			want: "components.foo-bar_baz.helm_values",
		},
		{
			name: "non-override reserved-looking key passes through",
			key:  "inputs",
			want: "inputs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := installDiffKey(tt.key); got != tt.want {
				t.Fatalf("installDiffKey(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}
