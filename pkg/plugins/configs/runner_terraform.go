package configs

import (
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
)

type RunnerTerraform struct {
	Plugin string `hcl:"plugin,label"`

	BundleName string `hcl:"bundle_name"`

	TerraformVersion string `hcl:"terraform_version"`

	AWSAuth   *awscredentials.Config   `hcl:"aws_auth,block"`
	AzureAuth *azurecredentials.Config `hcl:"azure_auth,block"`

	Backend TerraformDeployBackend `hcl:"backend,block"`

	Variables map[string]string `hcl:"variables"`
	EnvVars   map[string]string `hcl:"env_vars"`
}
