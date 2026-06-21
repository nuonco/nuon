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
	// RoleConfig is the per-role payload for break-glass/custom roles.
	RoleConfig = core.RoleConfig
	// Method selects which provisioning implementation drives an install stack.
	Method = core.Method
)

// Provisioning methods.
const (
	MethodAWSSDK    = core.MethodAWSSDK
	MethodTerraform = core.MethodTerraform
)
