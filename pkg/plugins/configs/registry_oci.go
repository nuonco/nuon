package configs

import (
	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	azurecredentials "github.com/powertoolsdev/mono/pkg/azure/credentials"
)

type OCIRegistryType string

const (
	OCIRegistryTypeECR  OCIRegistryType = "ecr"
	OCIRegistryTypeACR  OCIRegistryType = "acr"
	OCIRegistryTypePrivateOCI OCIRegistryType = "private_oci"
	OCIRegistryTypePublicOCI  OCIRegistryType = "public_oci"
)

type OCIRegistryAuth struct {
	Username string `hcl:"username"`
	Password string `hcl:"password"`
}

// NOTE(jm): this is the registry config we are consolidating around for _all_ operations, as it should support all of
// the credential tooling we need and support public/private configs and more.
type OCIRegistryRepository struct {
	Plugin string `hcl:"plugin,label"`

	RegistryType OCIRegistryType `hcl:"registry_type,optional"`

	Region string `hcl:"region"`

	ECRAuth *awscredentials.Config   `hcl:"ecr_auth,block"`
	ACRAuth *azurecredentials.Config `hcl:"acr_auth,block"`
	OCIAuth *OCIRegistryAuth         `hcl:"ocr_auth,block"`

	// based on the type of access, either the repository (ecr) or login server (acr) will be provided.
	Repository  string `hcl:"repository"`
	LoginServer string `hcl:"login_server"`
}
