package core

// Method selects which provisioning implementation drives an install stack.
type Method string

const (
	// MethodTerraform provisions by applying the install-stacks Terraform
	// module (internal/terraform). It is currently the only supported method.
	MethodTerraform Method = "terraform"
)

// DefaultMethod is used when neither Config nor Options specifies one.
const DefaultMethod = MethodTerraform
