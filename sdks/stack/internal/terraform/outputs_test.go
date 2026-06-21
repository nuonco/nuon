package terraform

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
)

// TestOutputsToCore locks the module output names to install-stacks/aws
// outputs.tf and verifies typed decoding into core.Outputs.
func TestOutputsToCore(t *testing.T) {
	raw := map[string]string{
		"account_id":               `"123456789012"`,
		"region":                   `"us-east-1"`,
		"vpc_id":                   `"vpc-abc"`,
		"runner_subnet":            `"subnet-runner"`,
		"runner_security_group_id": `"sg-abc"`,
		"runner_iam_role_arn":      `"arn:aws:iam::123456789012:role/inst-runner"`,
		"runner_instance_profile":  `"arn:aws:iam::123456789012:instance-profile/inst-runner"`,
		"runner_asg_name":          `"inst-asg"`,
		"runner_log_group_name":    `"/nuon/inst"`,
		"provision_iam_role_arn":   `"arn:aws:iam::123456789012:role/inst-provision"`,
		"maintenance_iam_role_arn": `""`,
		"deprovision_iam_role_arn": `""`,
		"public_subnets":           `["subnet-a","subnet-b"]`,
		"private_subnets":          `["subnet-c"]`,
		"break_glass_role_arns":    `{"bg":"arn:aws:iam::123456789012:role/bg"}`,
		"custom_role_arns":         `{}`,
		"secret_arns":              `{"db_arn":"arn:aws:secretsmanager:us-east-1:123456789012:secret:db"}`,
		"install_inputs":           `{"cluster_name":"prod"}`,
	}
	meta := map[string]tfexec.OutputMeta{}
	for k, v := range raw {
		meta[k] = tfexec.OutputMeta{Value: json.RawMessage(v)}
	}

	out, err := outputsToCore(meta)
	if err != nil {
		t.Fatalf("outputsToCore: %v", err)
	}

	if out.AccountID != "123456789012" {
		t.Errorf("AccountID = %q", out.AccountID)
	}
	if out.VPCID != "vpc-abc" {
		t.Errorf("VPCID = %q", out.VPCID)
	}
	if len(out.PublicSubnetIDs) != 2 || out.PublicSubnetIDs[0] != "subnet-a" {
		t.Errorf("PublicSubnetIDs = %v", out.PublicSubnetIDs)
	}
	if out.BreakGlassRoleARNs["bg"] != "arn:aws:iam::123456789012:role/bg" {
		t.Errorf("BreakGlassRoleARNs = %v", out.BreakGlassRoleARNs)
	}
	if out.SecretARNs["db_arn"] == "" {
		t.Errorf("SecretARNs missing db_arn: %v", out.SecretARNs)
	}
	if out.InstallInputs["cluster_name"] != "prod" {
		t.Errorf("InstallInputs = %v", out.InstallInputs)
	}
}
