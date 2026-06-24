package core

// Outputs is the method-agnostic result of a successful provision. Fields
// common to every cloud live here; cloud-specific resolved values live on a
// per-cloud sub-struct (AWS, GCP, …), exactly one of which is populated,
// selected by Cloud. Every provisioning method must populate it with
// fully-resolved values so the public stack package can build an identical
// phone-home payload regardless of which method ran.
type Outputs struct {
	// Cloud identifies which sub-struct is populated.
	Cloud Cloud

	// InstallInputs echoes back the customer install inputs the run resolved.
	// Cloud-agnostic, so it lives on the common struct.
	InstallInputs map[string]string

	// AWS carries the AWS-specific outputs; populated when Cloud is aws.
	AWS *AWSOutputs
	// GCP carries the GCP-specific outputs; populated when Cloud is gcp.
	GCP *GCPOutputs
}

// AWSOutputs is the AWS-specific result of a successful provision. The field
// set mirrors install-stacks/aws/phone_home.tf so app templates resolving
// `nuon.install_stack.outputs.*` see the same keys across methods. Values are
// fully resolved (ARNs, not names).
type AWSOutputs struct {
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
}
