package configs

import (
	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	azurecredentials "github.com/powertoolsdev/mono/pkg/azure/credentials"
)

// RunnerTerraform is a terraform config that is used by the runner terraform job, to deploy runners using Terraform.
type RunnerTerraform struct {
	Plugin string `hcl:"plugin,label"`

	BundleName string `hcl:"bundle_name"`

	TerraformVersion string `hcl:"terraform_version"`

	// auth for the run itself
	AWSAuth     *awscredentials.Config   `hcl:"aws_auth,block"`
	AzureAuth *azurecredentials.Config `hcl:"azure_auth,block"`

	// Backend is used to configure where/how the backend is run
	Backend TerraformDeployBackend `hcl:"backend,block"`

	Variables map[string]string `hcl:"variables"`
	EnvVars   map[string]string `hcl:"env_vars"`
}
