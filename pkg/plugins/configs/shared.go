package configs

import (
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
)

type OciArchive struct {
	Image       string `hcl:"image"`
	Tag         string `hcl:"tag"`
	LoginServer string `hcl:"login_server,optional"`

	RegistryType OCIRegistryType `hcl:"registry_type"`

	ECRAuth *awscredentials.Config   `hcl:"ecr_auth,optional" validate:"required"`
	ACRAuth *azurecredentials.Config `hcl:"acr_auth,optional" validate:"required"`
}
