package installs

import (
	"testing"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
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

func TestInputDefaults(t *testing.T) {
	inputs := []*models.AppAppInput{
		{Name: "optional_with_default", Default: "opt"},
		{Name: "required_with_default", Required: true, Default: "req"},
		{Name: "required_no_default", Required: true},
		{Name: "sensitive_with_default", Required: true, Default: "secret", Sensitive: true},
		{Name: "customer_owned", Default: "cust", Source: string(models.AppAppInputSourceCustomer)},
	}

	got := inputDefaults(inputs)
	want := map[string]string{
		"optional_with_default": "opt",
		"required_with_default": "req",
	}

	if len(got) != len(want) {
		t.Fatalf("inputDefaults() = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("inputDefaults()[%q] = %q, want %q", k, got[k], v)
		}
	}
}

func TestResolveRequiredInputs(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []*models.AppAppInput
		defined  map[string]string
		current  map[string]string
		wantFill map[string]string
		wantErrs []string
	}{
		{
			name:     "required input with a default is not needed in the config",
			inputs:   []*models.AppAppInput{{Name: "runner_image_url", Required: true, Default: "upstream"}},
			defined:  map[string]string{},
			current:  map[string]string{},
			wantFill: map[string]string{"runner_image_url": "upstream"},
		},
		{
			name:     "required input with no default is still an error",
			inputs:   []*models.AppAppInput{{Name: "root_domain", Required: true}},
			defined:  map[string]string{},
			current:  map[string]string{},
			wantFill: map[string]string{},
			wantErrs: []string{"missing required input root_domain"},
		},
		{
			name:     "value already on the install is left alone, not reverted to the default",
			inputs:   []*models.AppAppInput{{Name: "runner_image_url", Required: true, Default: "upstream"}},
			defined:  map[string]string{},
			current:  map[string]string{"runner_image_url": "vendored"},
			wantFill: map[string]string{},
		},
		{
			name:     "config value wins over the default",
			inputs:   []*models.AppAppInput{{Name: "runner_image_url", Required: true, Default: "upstream"}},
			defined:  map[string]string{"runner_image_url": "explicit"},
			current:  map[string]string{},
			wantFill: map[string]string{},
		},
		{
			name: "customer owned input still cannot be set from the config",
			inputs: []*models.AppAppInput{
				{Name: "region", Required: true, Default: "us-west-2", Source: string(models.AppAppInputSourceCustomer)},
			},
			defined:  map[string]string{"region": "us-east-1"},
			current:  map[string]string{},
			wantFill: map[string]string{},
			wantErrs: []string{"refusing to set user_configurable input region"},
		},
		{
			name: "optional input with no value is not filled",
			inputs: []*models.AppAppInput{
				{Name: "sub_domain", Default: "app"},
			},
			defined:  map[string]string{},
			current:  map[string]string{},
			wantFill: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fill, errs := resolveRequiredInputs(tt.inputs, tt.defined, tt.current)

			if len(fill) != len(tt.wantFill) {
				t.Fatalf("fill = %v, want %v", fill, tt.wantFill)
			}
			for k, v := range tt.wantFill {
				if fill[k] != v {
					t.Errorf("fill[%q] = %q, want %q", k, fill[k], v)
				}
			}

			if len(errs) != len(tt.wantErrs) {
				t.Fatalf("errs = %v, want %v", errs, tt.wantErrs)
			}
			for i, want := range tt.wantErrs {
				if errs[i].Error() != want {
					t.Errorf("errs[%d] = %q, want %q", i, errs[i].Error(), want)
				}
			}
		})
	}
}
