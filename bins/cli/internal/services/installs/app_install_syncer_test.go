package installs

import (
	"strings"
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

func TestCheckImmutableTargetAccount(t *testing.T) {
	tests := []struct {
		name      string
		local     *config.Install
		upstream  *config.Install
		wantErr   bool
		wantMatch string
	}{
		{
			name:     "no upstream is a create, nothing to protect",
			local:    &config.Install{AWSAccount: &config.AWSAccount{AccountID: "123456789012"}},
			upstream: nil,
		},
		{
			name:     "unchanged account id is fine",
			local:    &config.Install{AWSAccount: &config.AWSAccount{AccountID: "123456789012"}},
			upstream: &config.Install{AWSAccount: &config.AWSAccount{AccountID: "123456789012"}},
		},
		{
			name:     "config omitting the account id means don't care, not clear it",
			local:    &config.Install{AWSAccount: &config.AWSAccount{Region: "us-west-2"}},
			upstream: &config.Install{AWSAccount: &config.AWSAccount{AccountID: "123456789012"}},
		},
		{
			name:      "changed aws account id is refused",
			local:     &config.Install{AWSAccount: &config.AWSAccount{AccountID: "999999999999"}},
			upstream:  &config.Install{AWSAccount: &config.AWSAccount{AccountID: "123456789012"}},
			wantErr:   true,
			wantMatch: "aws_account.account_id",
		},
		{
			name:      "setting an account id on an install that has none is refused",
			local:     &config.Install{AWSAccount: &config.AWSAccount{AccountID: "123456789012"}},
			upstream:  &config.Install{AWSAccount: &config.AWSAccount{}},
			wantErr:   true,
			wantMatch: "aws_account.account_id",
		},
		{
			name:      "changed azure subscription id is refused",
			local:     &config.Install{AzureAccount: &config.AzureAccount{SubscriptionID: "sub-b"}},
			upstream:  &config.Install{AzureAccount: &config.AzureAccount{SubscriptionID: "sub-a"}},
			wantErr:   true,
			wantMatch: "azure_account.subscription_id",
		},
		{
			name:      "changed gcp project id is refused",
			local:     &config.Install{GCPAccount: &config.GCPAccount{ProjectID: "proj-b"}},
			upstream:  &config.Install{GCPAccount: &config.GCPAccount{ProjectID: "proj-a"}},
			wantErr:   true,
			wantMatch: "gcp_account.project_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkImmutableTargetAccount(tt.local, tt.upstream)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantMatch) {
					t.Errorf("error %q does not mention %q", err, tt.wantMatch)
				}
				if !strings.Contains(err.Error(), "immutable") {
					t.Errorf("error %q should explain the field is immutable", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
