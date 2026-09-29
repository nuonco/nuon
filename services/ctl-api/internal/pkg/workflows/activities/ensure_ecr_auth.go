package activities

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/plugins/configs"
)

func EnsureECRAuth(ctx workflow.Context, cfg *configs.OCIRegistryRepository) error {
	if cfg == nil || cfg.RegistryType != configs.OCIRegistryTypeECR {
		return nil
	}
	if cfg.OCIAuth != nil && cfg.OCIAuth.Password != "" {
		return nil
	}

	token, err := AwaitGetECRAccessToken(ctx, &GetECRAccessTokenRequest{
		Credentials: cfg.ECRAuth,
	})
	if err != nil {
		return err
	}

	// why: LoginServer must be scheme-less or RepositoryURI doubles the host.
	cfg.RegistryType = configs.OCIRegistryTypePrivateOCI
	cfg.LoginServer = token.ServerAddress
	cfg.OCIAuth = &configs.OCIRegistryAuth{
		Username: token.Username,
		Password: token.Password,
	}
	cfg.ECRAuth = nil

	return nil
}
