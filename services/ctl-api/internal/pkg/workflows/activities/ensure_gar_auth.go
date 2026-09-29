package activities

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/plugins/configs"
)

func EnsureGARAuth(ctx workflow.Context, cfg *configs.OCIRegistryRepository) error {
	if cfg == nil || cfg.RegistryType != configs.OCIRegistryTypeGAR {
		return nil
	}
	if cfg.OCIAuth != nil && cfg.OCIAuth.Password != "" {
		return nil
	}

	token, err := AwaitGetGARAccessToken(ctx, &GetGARAccessTokenRequest{
		ServiceAccountEmail:      cfg.ServiceAccountEmail,
		WorkloadIdentityProvider: cfg.WorkloadIdentityProvider,
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
