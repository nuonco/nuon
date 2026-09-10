package activities

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/plugins/configs"
)

// EnsureECRAuth mints an ECR authorization token into cfg.OCIAuth so a runner
// with no AWS identity can pull the artifact.
//
// An ECR config carries an assume-role for the runner to perform itself, which
// only works when the runner is in the control plane's cloud. A GCP or Azure
// runner has no AWS identity, so that walks the default chain to IMDS and 404s.
// GAR and ACR already mint credentials here for the same reason; ECR was the
// only registry type still handing the runner a role to assume.
//
// The type is switched to private_oci because the credentials are now embedded
// and the runner must not re-derive them from the assume-role path.
//
// Safe to call on any registry type; it is a no-op unless cfg is ECR without a
// token already attached.
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

	// docker.FetchAccessInfo keys the credential off LoginServer, which an ECR
	// config never sets, and AccessInfo.RepositoryURI only avoids doubling the
	// host when the repository already starts with it. Both need it scheme-less.
	cfg.RegistryType = configs.OCIRegistryTypePrivateOCI
	cfg.LoginServer = token.ServerAddress
	cfg.OCIAuth = &configs.OCIRegistryAuth{
		Username: token.Username,
		Password: token.Password,
	}
	cfg.ECRAuth = nil

	return nil
}
