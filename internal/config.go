package install

import (
	"github.com/go-playground/validator/v10"
)

// Config exposes a set of configuration options for the install domain
// TODO(jm): update install.Provision to use these config options
type Config struct {
	// NuonAccessRoleArn is the role that we add to the sandbox EKS allowed roles so we can do other operations
	// against it
	NuonAccessRoleArn string `config:"nuon_access_role_arn" validate:"required"`

	TokenSecretNamespace string `config:"token_secret_namespace" validate:"required"`
	OrgServerRootDomain  string `config:"org_server_root_domain" validate:"required"`
	OrgAccountID         string `config:"org_account_id" validate:"required"`

	InstallationStateBucket       string `config:"installation_state_bucket" validate:"required"`
	InstallationStateBucketRegion string `config:"installation_state_bucket_region" validate:"required"`
	SandboxBucket                 string `config:"sandbox_bucket" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
