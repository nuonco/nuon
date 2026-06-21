package core

// Method selects which provisioning implementation drives an install stack.
type Method string

const (
	// MethodSDK provisions resources directly via the AWS SDK (internal/awssdk).
	MethodSDK Method = "sdk"
	// MethodTerraform provisions by applying the install-stacks/aws Terraform
	// module (internal/terraform).
	MethodTerraform Method = "terraform"
)

// DefaultMethod is used when neither Config nor Options specifies one.
const DefaultMethod = MethodSDK
