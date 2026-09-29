package activities

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/plugins/configs"
)

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
