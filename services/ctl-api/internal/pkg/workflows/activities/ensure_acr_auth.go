package activities

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/plugins/configs"
)

// EnsureACRAuth mints an ACR refresh token into cfg.OCIAuth so the runner can
// authenticate to a registry it has no identity for.
//
// The runner only falls back to ambient Azure credentials when no token is
// supplied (see pkg/runner/registry/acr.FetchAccessInfo), which holds solely for
// registries in the same tenant as the runner's own managed identity. A vendor's
// registry lives in the vendor's tenant, and managed identity cannot cross that
// boundary, so the token has to be minted here.
//
// Only applies when the config names an app registration. A plain ACR config
// keeps the existing ambient-identity behaviour, so same-tenant registries that
// work today are untouched.
//
// Safe to call on any registry type; it is a no-op unless cfg is ACR with an app
// registration and no token already attached.
func EnsureACRAuth(ctx workflow.Context, cfg *configs.OCIRegistryRepository) error {
	if cfg == nil || cfg.RegistryType != configs.OCIRegistryTypeACR {
		return nil
	}
	if cfg.ACRAppRegistration == nil {
		return nil
	}
	if cfg.OCIAuth != nil && cfg.OCIAuth.Password != "" {
		return nil
	}

	reg := cfg.ACRAppRegistration
	token, err := AwaitGetACRAccessToken(ctx, &GetACRAccessTokenRequest{
		ComponentID:           reg.ComponentID,
		LoginServer:           cfg.LoginServer,
		TenantID:              reg.TenantID,
		ClientID:              reg.ClientID,
		ClientSecretName:      reg.ClientSecretName,
		ClientCertificateName: reg.ClientCertificateName,
	})
	if err != nil {
		return err
	}

	cfg.OCIAuth = &configs.OCIRegistryAuth{
		Username: token.Username,
		Password: token.Password,
	}

	return nil
}
