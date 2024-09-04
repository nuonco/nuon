package configs

import "github.com/powertoolsdev/mono/pkg/aws/credentials"

type SandboxTerraform struct {
	Plugin string `hcl:"plugin,label"`

	// TODO(jm): should be deprecated
	DirArchive *TerraformDeployDirArchive `hcl:"local_archive,block"`

	TerraformVersion string                 `hcl:"terraform_version"`
	RunType          TerraformDeployRunType `hcl:"run_type"`

	// auth for the run itself
	RunAuth credentials.Config `hcl:"run_auth,block"`

	// Backend is used to configure where/how the backend is run
	Backend TerraformDeployBackend `hcl:"backend,block"`

	// Outputs are used to control where the run outputs are synchronized to
	Outputs TerraformDeployOutputs `hcl:"outputs,block"`

	Labels    map[string]string `hcl:"labels" validate:"required"`
	Variables map[string]string `hcl:"variables"`
	EnvVars   map[string]string `hcl:"env_vars"`

	// NOTE(jm): I'm not a fan of this approach, but it's faster than building a custom map[string]interface{}
	// decoder for go-hcl. Go-HCL does not support map[string]interface{} out of the box.
	VariablesJSON string                `hcl:"variables_json"`
	Hooks         *TerraformDeployHooks `hcl:"hooks,block"`
}
