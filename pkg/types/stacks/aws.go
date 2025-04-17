package stacks

// AWSCloudFormationOutputs are used to define the stack outputs for a stack.
// This type can be used from anywhere, and is intentionally set up for testing here.
type AWSCloudFormationOutputs struct {
	VPCID          string `mapstructure:"vpc_id"`
	AccountID      string `mapstructure:"account_id"`
	RunnerSubnet   string `mapstructure:"runner_subnet"`
	PrivateSubnets string `mapstructure:"runner_subnet"`

	ProvisionIAMRoleARN   string `mapstructure:"provision_iam_role_arn"`
	MaintenanceIAMRoleARN string `mapstructure:"maintenance_subnet"`
	DeprovisionIAMRoleARN string `mapstructure:"deprovision_iam_role_arn"`
}
