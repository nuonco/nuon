package preflight

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"

	internal "github.com/nuonco/nuon/services/ctl-api/internal"
	ctlgithub "github.com/nuonco/nuon/services/ctl-api/internal/pkg/github"
)

var githubCheck = Check{
	Name:        "github",
	Description: "github app credentials",

	Fields: func(cfg *internal.Config) []Field {
		return []Field{
			{Name: "github_app_id", Value: cfg.GithubAppID, Required: true},
			{Name: "github_app_key", Value: cfg.GithubAppKey, Required: true, Secret: true},
			{Name: "github_app_key_secret_name", Value: cfg.GithubAppKeySecretName},
			// The connection callback exchanges a user-authorization code with
			// these credentials to prove the caller owns the installation, so
			// a deployment that runs the check cannot boot without them.
			{Name: "github_app_client_id", Value: cfg.GithubAppClientID, Required: !cfg.DisableVCSConnectionOwnershipCheck},
			{Name: "github_app_client_secret", Value: cfg.GithubAppClientSecret, Required: !cfg.DisableVCSConnectionOwnershipCheck, Secret: true},
			{Name: "disable_vcs_connection_ownership_check", Value: strconv.FormatBool(cfg.DisableVCSConnectionOwnershipCheck)},
			{Name: "integration_github_install_id", Value: cfg.IntegrationGithubInstallID},
		}
	},

	// Offline: building the client parses and validates the PEM key, which is
	// the failure this catches. A live API call would need an installation
	// token, which is per-repo rather than config.
	Probe: func(_ context.Context, cfg *internal.Config) (string, error) {
		if _, err := ctlgithub.New(validator.New(), cfg); err != nil {
			return "", fmt.Errorf("invalid github app credentials: %w", err)
		}

		return fmt.Sprintf("app key valid %s",
			summary("app_id", cfg.GithubAppID)), nil
	},
}
