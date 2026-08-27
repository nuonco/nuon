package configs

import (
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
)

type OCIRegistryType string

const (
	OCIRegistryTypeECR        OCIRegistryType = "ecr"
	OCIRegistryTypeACR        OCIRegistryType = "acr"
	OCIRegistryTypeGAR        OCIRegistryType = "gar"
	OCIRegistryTypePrivateOCI OCIRegistryType = "private_oci"
	OCIRegistryTypePublicOCI  OCIRegistryType = "public_oci"
)

type OCIRegistryAuth struct {
	Username string `hcl:"username"`
	Password string `hcl:"password"`
}

// ACRAppRegistration identifies an app registration in a registry owner's own
// tenant, which is the only way to reach an ACR that the control plane holds no
// identity in.
//
// It carries names, never credential material: azurecredentials.Config's secret
// fields are unserializable on purpose, so the material cannot ride along in a
// plan. The control plane resolves these refs, mints a short-lived registry
// token, and puts that in OCIAuth for the runner.
type ACRAppRegistration struct {
	ComponentID string `hcl:"component_id,optional"`
	TenantID    string `hcl:"tenant_id,optional"`
	ClientID    string `hcl:"client_id,optional"`

	// App secret names holding the credential. Exactly one is set.
	ClientSecretName      string `hcl:"client_secret_name,optional"`
	ClientCertificateName string `hcl:"client_certificate_name,optional"`
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

	ServiceAccountEmail      string `hcl:"service_account_email,optional"`
	WorkloadIdentityProvider string `hcl:"workload_identity_provider,optional"`

	ACRAppRegistration *ACRAppRegistration `hcl:"acr_app_registration,block"`

	// based on the type of access, either the repository (ecr) or login server (acr) will be provided.
	Repository  string `hcl:"repository"`
	LoginServer string `hcl:"login_server"`
}
