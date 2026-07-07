package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	stack "github.com/nuonco/nuon/sdks/stack"
)

func TestStackDataSourceSchema(t *testing.T) {
	ctx := context.Background()
	d := NewStackDataSource()
	var resp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %+v", resp.Diagnostics)
	}
	if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("invalid schema implementation: %+v", diags)
	}
}

func TestFlattenConfigGCP(t *testing.T) {
	cfg := &stack.Config{
		InstallID:     "inst123",
		OrgID:         "org123",
		AppID:         "app123",
		Cloud:         stack.CloudGCP,
		RunnerID:      "runner123",
		RunnerAPIURL:  "https://runner.example.com",
		PhoneHomeURL:  "https://api.example.com/v1/installs/inst123/phone-home/ph123",
		InstallInputs: map[string]string{"domain": "example.com"},
		Secrets: map[string]stack.SecretInput{
			"db_password": {Description: "db", Required: true, Value: "hunter2"},
		},
		GCP: &stack.GCPConfig{
			RunnerInitScriptURL:  "https://init.sh",
			RunnerAPIToken:       "tok",
			ProvisionPermissions: []string{"compute.instances.create"},
			CustomRoles: map[string]stack.GCPRole{
				"extra": {Permissions: []string{"storage.buckets.get"}, Enabled: true},
			},
		},
	}

	var data stackDataSourceModel
	data.PhoneHomeID = types.StringValue("ph123")
	flattenConfig(&data, cfg)

	if data.InstallID.ValueString() != "inst123" {
		t.Errorf("install_id = %q", data.InstallID.ValueString())
	}
	if data.Cloud.ValueString() != "gcp" {
		t.Errorf("cloud = %q", data.Cloud.ValueString())
	}
	if data.PhoneHomeID.ValueString() != "ph123" {
		t.Errorf("phone_home_id overwritten: %q", data.PhoneHomeID.ValueString())
	}
	if data.Secrets["db_password"].Value != "hunter2" {
		t.Errorf("secret value = %q", data.Secrets["db_password"].Value)
	}
	if data.GCP == nil || data.GCP.RunnerAPIToken != "tok" {
		t.Fatalf("gcp not flattened: %+v", data.GCP)
	}
	if r := data.GCP.CustomRoles["extra"]; !r.Enabled || len(r.Permissions) != 1 {
		t.Errorf("custom role = %+v", r)
	}
}
