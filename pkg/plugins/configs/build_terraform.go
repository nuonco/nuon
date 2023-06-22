package configs

// TerraformBuild is used by the terraform plugin to create an OCI archive with the build parameters.
type TerraformBuild struct {
	Plugin string `hcl:"plugin,label"`

	OutputName string `hcl:"output_name,optional"`

	Labels    map[string]string `hcl:"labels,optional"`
	Variables map[string]string `hcl:"variables,optional"`
}
