package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	stack "github.com/nuonco/nuon/sdks/stack"
)

func TestStackResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewStackResource()
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %+v", resp.Diagnostics)
	}
	if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("invalid schema implementation: %+v", diags)
	}
}

func TestFlattenOutputs(t *testing.T) {
	out := &stack.Outputs{
		Cloud:         stack.CloudAWS,
		InstallInputs: map[string]string{"domain": "example.com"},
		AWS: &stack.AWSOutputs{
			AccountID:        "123456789012",
			Region:           "us-west-2",
			VPCID:            "vpc-abc",
			PublicSubnetIDs:  []string{"subnet-2", "subnet-1"},
			RunnerIAMRoleARN: "arn:aws:iam::123:role/r",
			SecretARNs:       map[string]string{"db_arn": "arn:aws:secretsmanager:x"},
		},
	}
	m := flattenOutputs(out)

	checks := map[string]string{
		"cloud":               "aws",
		"account_id":          "123456789012",
		"region":              "us-west-2",
		"vpc_id":              "vpc-abc",
		"runner_iam_role_arn": "arn:aws:iam::123:role/r",
		"public_subnet_ids":   "subnet-1,subnet-2",
		"input.domain":        "example.com",
		"secret.db_arn":       "arn:aws:secretsmanager:x",
	}
	for k, want := range checks {
		if got := m[k]; got != want {
			t.Errorf("outputs[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestFlattenOutputsNil(t *testing.T) {
	if m := flattenOutputs(nil); len(m) != 0 {
		t.Errorf("expected empty map, got %+v", m)
	}
}
