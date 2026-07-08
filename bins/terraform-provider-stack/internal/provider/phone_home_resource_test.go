package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestPhoneHomeResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewPhoneHomeResource()
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %+v", resp.Diagnostics)
	}
	if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("invalid schema implementation: %+v", diags)
	}
}

func TestPhoneHomeResourceTypeName(t *testing.T) {
	ctx := context.Background()
	r := NewPhoneHomeResource()
	var resp resource.MetadataResponse
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "stack"}, &resp)
	if resp.TypeName != "stack_phone_home" {
		t.Errorf("type name = %q, want stack_phone_home", resp.TypeName)
	}
}
