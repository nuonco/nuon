package stack

import "github.com/nuonco/nuon/sdks/stack/internal/core"

// The per-install configuration types live in internal/core so the
// provisioning method packages (internal/awssdk, internal/terraform, …) can
// share them without importing this package. They are re-exported here as
// aliases to keep the public SDK surface stable for embedders and stack-cli.
type (
	// Config carries the per-install rendered configuration that ctl-api
	// produces alongside a stack run.
	Config = core.Config
	// SecretInput mirrors the customer-provided secret shape.
	SecretInput = core.SecretInput
	// RoleConfig is the per-role payload for AWS break-glass/custom roles.
	RoleConfig = core.RoleConfig
	// AWSConfig carries the AWS-specific install-stack inputs.
	AWSConfig = core.AWSConfig
	// GCPConfig carries the GCP-specific install-stack inputs.
	GCPConfig = core.GCPConfig
	// GCPRole is the per-role payload for GCP break-glass/custom roles.
	GCPRole = core.GCPRole
	// Method selects which provisioning implementation drives an install stack.
	Method = core.Method
	// Cloud selects which cloud provider an install stack targets.
	Cloud = core.Cloud
	// TerraformBackend configures the terraform method's remote state backend
	// (S3 for AWS, GCS for GCP) in the customer's target account.
	TerraformBackend = core.TerraformBackend
)

// Provisioning methods.
const (
	MethodSDK       = core.MethodSDK
	MethodTerraform = core.MethodTerraform
)

// Cloud providers.
const (
	CloudAWS   = core.CloudAWS
	CloudGCP   = core.CloudGCP
	CloudAzure = core.CloudAzure
)
