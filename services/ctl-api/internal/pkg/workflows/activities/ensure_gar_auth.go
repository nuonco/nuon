package activities

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/plugins/configs"
)

// EnsureGARAuth mints a GAR access token into cfg.OCIAuth so the runner can
// authenticate without GCP application default credentials of its own.
//
// The runner only falls back to ADC when no token is supplied (see
// pkg/runner/registry/gar.FetchAccessInfo), which holds solely for runners that
// themselves run in GCP. A control plane in GCP can hand artifacts to a runner
// in AWS or Azure, so every plan carrying a GAR repository has to embed
// credentials minted here.
//
// Safe to call on any registry type; it is a no-op unless cfg is GAR without a
// token already attached.
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
