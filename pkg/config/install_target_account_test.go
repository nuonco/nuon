package config

import (
	"strings"
	"testing"
)

func TestCheckImmutableTargetAccount(t *testing.T) {
	tests := []struct {
		name      string
		local     *Install
		upstream  *Install
		wantErr   bool
		wantMatch string
	}{
		{
			name:     "no upstream is a create, nothing to protect",
			local:    &Install{AWSAccount: &AWSAccount{AccountID: "123456789012"}},
			upstream: nil,
		},
		{
			name:     "unchanged account id is fine",
			local:    &Install{AWSAccount: &AWSAccount{AccountID: "123456789012"}},
			upstream: &Install{AWSAccount: &AWSAccount{AccountID: "123456789012"}},
		},
		{
			name:     "config omitting the account id means don't care, not clear it",
			local:    &Install{AWSAccount: &AWSAccount{Region: "us-west-2"}},
			upstream: &Install{AWSAccount: &AWSAccount{AccountID: "123456789012"}},
		},
		{
			name:      "changed aws account id is refused",
			local:     &Install{AWSAccount: &AWSAccount{AccountID: "999999999999"}},
			upstream:  &Install{AWSAccount: &AWSAccount{AccountID: "123456789012"}},
			wantErr:   true,
			wantMatch: "aws_account.account_id",
		},
		{
			name:      "setting an account id on an install that has none is refused",
			local:     &Install{AWSAccount: &AWSAccount{AccountID: "123456789012"}},
			upstream:  &Install{AWSAccount: &AWSAccount{}},
			wantErr:   true,
			wantMatch: "aws_account.account_id",
		},
		{
			name:      "changed azure subscription id is refused",
			local:     &Install{AzureAccount: &AzureAccount{SubscriptionID: "sub-b"}},
			upstream:  &Install{AzureAccount: &AzureAccount{SubscriptionID: "sub-a"}},
			wantErr:   true,
			wantMatch: "azure_account.subscription_id",
		},
		{
			name:      "changed gcp project id is refused",
			local:     &Install{GCPAccount: &GCPAccount{ProjectID: "proj-b"}},
			upstream:  &Install{GCPAccount: &GCPAccount{ProjectID: "proj-a"}},
			wantErr:   true,
			wantMatch: "gcp_account.project_id",
		},
		{
			name:      "every changed identifier is reported, not just the first",
			local:     &Install{AWSAccount: &AWSAccount{AccountID: "999999999999"}, GCPAccount: &GCPAccount{ProjectID: "proj-b"}},
			upstream:  &Install{AWSAccount: &AWSAccount{AccountID: "123456789012"}, GCPAccount: &GCPAccount{ProjectID: "proj-a"}},
			wantErr:   true,
			wantMatch: "gcp_account.project_id",
		},
		{
			name:     "a nil receiver is a no-op rather than a panic",
			local:    nil,
			upstream: &Install{AWSAccount: &AWSAccount{AccountID: "123456789012"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.local.CheckImmutableTargetAccount(tt.upstream)
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
