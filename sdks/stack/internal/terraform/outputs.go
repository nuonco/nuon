package terraform

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-exec/tfexec"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

// outputsToCore maps the module's `terraform output` result into core.Outputs.
// Output names mirror install-stacks/aws/outputs.tf (which in turn mirror the
// CFN phone-home payload), so this is a direct key-by-key translation.
func outputsToCore(meta map[string]tfexec.OutputMeta) (*core.Outputs, error) {
	out := &core.Outputs{
		Cloud:         core.CloudAWS,
		InstallInputs: map[string]string{},
		AWS: &core.AWSOutputs{
			BreakGlassRoleARNs: map[string]string{},
			CustomRoleARNs:     map[string]string{},
			SecretARNs:         map[string]string{},
		},
	}

	getString := func(key string, dst *string) error {
		m, ok := meta[key]
		if !ok {
			return nil
		}
		if err := json.Unmarshal(m.Value, dst); err != nil {
			return fmt.Errorf("output %q: %w", key, err)
		}
		return nil
	}
	getStrings := func(key string, dst *[]string) error {
		m, ok := meta[key]
		if !ok {
			return nil
		}
		if err := json.Unmarshal(m.Value, dst); err != nil {
			return fmt.Errorf("output %q: %w", key, err)
		}
		return nil
	}
	getStrMap := func(key string, dst *map[string]string) error {
		m, ok := meta[key]
		if !ok {
			return nil
		}
		if err := json.Unmarshal(m.Value, dst); err != nil {
			return fmt.Errorf("output %q: %w", key, err)
		}
		return nil
	}

	for key, fn := range map[string]func() error{
		"account_id":               func() error { return getString("account_id", &out.AWS.AccountID) },
		"region":                   func() error { return getString("region", &out.AWS.Region) },
		"vpc_id":                   func() error { return getString("vpc_id", &out.AWS.VPCID) },
		"runner_subnet":            func() error { return getString("runner_subnet", &out.AWS.RunnerSubnetID) },
		"runner_security_group_id": func() error { return getString("runner_security_group_id", &out.AWS.RunnerSecurityGroupID) },
		"runner_iam_role_arn":      func() error { return getString("runner_iam_role_arn", &out.AWS.RunnerIAMRoleARN) },
		"runner_instance_profile":  func() error { return getString("runner_instance_profile", &out.AWS.RunnerInstanceProfileARN) },
		"runner_asg_name":          func() error { return getString("runner_asg_name", &out.AWS.RunnerASGName) },
		"runner_log_group_name":    func() error { return getString("runner_log_group_name", &out.AWS.RunnerLogGroupName) },
		"provision_iam_role_arn":   func() error { return getString("provision_iam_role_arn", &out.AWS.ProvisionRoleARN) },
		"maintenance_iam_role_arn": func() error { return getString("maintenance_iam_role_arn", &out.AWS.MaintenanceRoleARN) },
		"deprovision_iam_role_arn": func() error { return getString("deprovision_iam_role_arn", &out.AWS.DeprovisionRoleARN) },
		"public_subnets":           func() error { return getStrings("public_subnets", &out.AWS.PublicSubnetIDs) },
		"private_subnets":          func() error { return getStrings("private_subnets", &out.AWS.PrivateSubnetIDs) },
		"break_glass_role_arns":    func() error { return getStrMap("break_glass_role_arns", &out.AWS.BreakGlassRoleARNs) },
		"custom_role_arns":         func() error { return getStrMap("custom_role_arns", &out.AWS.CustomRoleARNs) },
		"secret_arns":              func() error { return getStrMap("secret_arns", &out.AWS.SecretARNs) },
		"install_inputs":           func() error { return getStrMap("install_inputs", &out.InstallInputs) },
	} {
		if err := fn(); err != nil {
			return nil, fmt.Errorf("decode terraform outputs (%s): %w", key, err)
		}
	}

	return out, nil
}
