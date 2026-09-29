package configs

import "github.com/nuonco/nuon/pkg/aws/credentials"

type SandboxTerraform struct {
	Plugin string `hcl:"plugin,label"`

	// TODO(jm): should be deprecated
	DirArchive *TerraformDeployDirArchive `hcl:"local_archive,block"`

	TerraformVersion string                 `hcl:"terraform_version"`
	RunType          TerraformDeployRunType `hcl:"run_type"`

	RunAuth credentials.Config `hcl:"run_auth,block"`

	Backend TerraformDeployBackend `hcl:"backend,block"`

	Outputs TerraformDeployOutputs `hcl:"outputs,block"`

	Labels    map[string]string `hcl:"labels" validate:"required"`
	Variables map[string]string `hcl:"variables"`
	EnvVars   map[string]string `hcl:"env_vars"`

	VariablesJSON string                `hcl:"variables_json"`
	Hooks         *TerraformDeployHooks `hcl:"hooks,block"`
}
