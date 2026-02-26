package preflight

import (
	"context"
	"fmt"

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
