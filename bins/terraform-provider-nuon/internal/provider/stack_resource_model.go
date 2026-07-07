package provider

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	stack "github.com/nuonco/nuon/sdks/stack"
)

// stackResourceModel is the Terraform state/plan shape for nuon_stack.
type stackResourceModel struct {
	PhoneHomeID      types.String   `tfsdk:"phone_home_id"`
	TerraformVersion types.String   `tfsdk:"terraform_version"`
	ModuleRef        types.String   `tfsdk:"module_ref"`
	AWS              *awsBlockModel `tfsdk:"aws"`
	Inputs           types.Map      `tfsdk:"inputs"`
	Secrets          types.Map      `tfsdk:"secrets"`

	InstallID types.String `tfsdk:"install_id"`
	Outputs   types.Map    `tfsdk:"outputs"`
}

// awsBlockModel is the aws {} nested block. Region/account_id are optional (the
// control plane resolves them); state_bucket is required for remote state.
type awsBlockModel struct {
	Region        types.String `tfsdk:"region"`
	AccountID     types.String `tfsdk:"account_id"`
	StateBucket   types.String `tfsdk:"state_bucket"`
	StateKey      types.String `tfsdk:"state_key"`
	DynamoDBTable types.String `tfsdk:"dynamodb_table"`
}

// stringMap converts a Terraform map attribute into a Go map, reporting element
// conversion errors as diagnostics. A null/unknown map yields nil.
func stringMap(ctx context.Context, m types.Map, diags *diag.Diagnostics) map[string]string {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	out := make(map[string]string, len(m.Elements()))
	diags.Append(m.ElementsAs(ctx, &out, false)...)
	return out
}

// urlOptions builds the SDK URLOptions from the model for AWS (Kind is set by
// the caller per operation). Cloud-specific wiring (GCP) is added alongside the
// gcp block.
func (m *stackResourceModel) urlOptions(ctx context.Context, apiURL string, diags *diag.Diagnostics) stack.URLOptions {
	opts := stack.URLOptions{
		URL:           strings.TrimRight(apiURL, "/") + "/v1/stack-runs/" + m.PhoneHomeID.ValueString(),
		Cloud:         stack.CloudAWS,
		InstallInputs: stringMap(ctx, m.Inputs, diags),
		Secrets:       stringMap(ctx, m.Secrets, diags),
	}
	if m.AWS != nil {
		opts.Backend = stack.TerraformBackend{
			Bucket:        m.AWS.StateBucket.ValueString(),
			Key:           m.AWS.StateKey.ValueString(),
			Region:        m.AWS.Region.ValueString(),
			DynamoDBTable: m.AWS.DynamoDBTable.ValueString(),
		}
	}
	return opts
}

// flattenOutputs renders the SDK Outputs into a flat string map for the computed
// `outputs` attribute. Scalars map directly; slices join on comma; sub-maps are
// flattened with a key prefix. install-input echoes are prefixed `input.`.
func flattenOutputs(out *stack.Outputs) map[string]string {
	m := map[string]string{}
	if out == nil {
		return m
	}
	m["cloud"] = string(out.Cloud)
	for k, v := range out.InstallInputs {
		m["input."+k] = v
	}
	if a := out.AWS; a != nil {
		m["account_id"] = a.AccountID
		m["region"] = a.Region
		m["vpc_id"] = a.VPCID
		m["runner_subnet_id"] = a.RunnerSubnetID
		m["runner_security_group_id"] = a.RunnerSecurityGroupID
		m["runner_iam_role_arn"] = a.RunnerIAMRoleARN
		m["runner_instance_profile_arn"] = a.RunnerInstanceProfileARN
		m["runner_asg_name"] = a.RunnerASGName
		m["runner_log_group_name"] = a.RunnerLogGroupName
		m["provision_role_arn"] = a.ProvisionRoleARN
		m["maintenance_role_arn"] = a.MaintenanceRoleARN
		m["deprovision_role_arn"] = a.DeprovisionRoleARN
		setJoined(m, "public_subnet_ids", a.PublicSubnetIDs)
		setJoined(m, "private_subnet_ids", a.PrivateSubnetIDs)
		for k, v := range a.BreakGlassRoleARNs {
			m["break_glass_role_arn."+k] = v
		}
		for k, v := range a.CustomRoleARNs {
			m["custom_role_arn."+k] = v
		}
		for k, v := range a.SecretARNs {
			m["secret."+k] = v
		}
	}
	return m
}

func setJoined(m map[string]string, key string, vals []string) {
	if len(vals) == 0 {
		return
	}
	s := append([]string(nil), vals...)
	sort.Strings(s)
	m[key] = strings.Join(s, ",")
}
