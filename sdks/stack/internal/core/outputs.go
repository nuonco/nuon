package core

// Outputs is the method-agnostic result of a successful provision. Every
// provisioning method (AWS SDK, Terraform, CloudFormation) must populate it
// with fully-resolved values (ARNs, not names) so the public stack package
// can build an identical phone-home payload regardless of which method ran.
//
// The field set mirrors install-stacks/aws/phone_home.tf so app templates
// resolving `nuon.install_stack.outputs.*` see the same keys across methods.
type Outputs struct {
	AccountID string
	Region    string

	VPCID                 string
	RunnerSubnetID        string
	PublicSubnetIDs       []string
	PrivateSubnetIDs      []string
	RunnerSecurityGroupID string

	RunnerIAMRoleARN         string
	RunnerInstanceProfileARN string
	RunnerASGName            string
	RunnerLogGroupName       string

	ProvisionRoleARN   string
	MaintenanceRoleARN string
	DeprovisionRoleARN string
	BreakGlassRoleARNs map[string]string
	CustomRoleARNs     map[string]string

	// SecretARNs is keyed by `<name>_arn` to match the phone-home contract.
	SecretARNs map[string]string

	// InstallInputs echoes back the customer install inputs the run resolved.
	InstallInputs map[string]string
}
