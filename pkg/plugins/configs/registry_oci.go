package configs

import (
	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	azurecredentials "github.com/powertoolsdev/mono/pkg/azure/credentials"
)

type OCIRegistryType string

const (
	OCIRegistryTypeECR OCIRegistryType = "ecr"
	OCIRegistryTypeACR OCIRegistryType = "acr"
)

type OCIRegistry struct {
	Plugin string `hcl:"plugin,label"`

	RegistryType OCIRegistryType `hcl:"registry_type,optional"`

	Tag	string			 `hcl:"tag"`
	Region	string			 `hcl:"region"`
	ECRAuth *awscredentials.Config	 `hcl:"ecr_auth,block"`
	ACRAuth *azurecredentials.Config `hcl:"acr_auth,block"`

	// based on the type of access, either the repository (ecr) or login server (acr) will be provided.
	Repository  string `hcl:"repository"`
	LoginServer string `hcl:"login_server"`
}
